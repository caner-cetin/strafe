package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strafe/internal"
	"strafe/pkg/db"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/valyala/fastjson"
)

var (
	audioPath string
	audioCmd  = &cobra.Command{
		Use:   "audio",
		Short: "audio related commands",
		Long:  fmt.Sprintf(`audio related commands, all commands under this tree requires strafe image to be built, use command %s if required`, color.MagentaString("strafe docker build")),
	}
)

type AudioProcessConfig struct {
	ModelCheckpoint  string
	ModelDownloadDir string
	// output directory for audio separator stems,
	// vocals and instrumentals will be directories under this directory
	OutputDir      string
	WaveformPPS    int32
	UseGPU         bool
	IsInstrumental bool
	DryRun         bool
}

var (
	uploadCmd = &cobra.Command{
		Use:   "upload -a audio",
		Short: "processes and uploads audio",
		Long:  `processes and uploads audio file given with -a / --audio flag`,
		Run:   WrapCommandWithResources(processAndUploadAudio, ResourceConfig{Resources: []ResourceType{ResourceDocker}}),
	}
	uploadCfg = AudioProcessConfig{}
)

func getAudioRootCmd() *cobra.Command {
	uploadCmd.PersistentFlags().Int32VarP(&uploadCfg.WaveformPPS, "pps", "P", 100, "waveform zoom level (pixels per second)")
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.ModelCheckpoint, "model", "mel_band_roformer_karaoke_aufr33_viperx_sdr_10.1956.ckpt", "model name for audio splitter, see https://raw.githubusercontent.com/nomadkaraoke/python-audio-separator/refs/heads/main/audio_separator/models.json for full list")
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.ModelDownloadDir, "model_file_directory", "/tmp/audio-separator-models/", "model download folder / file directory on the host machine")
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.OutputDir, "audio_output_directory", "/tmp/strafe-audio-separator-audio/", "directory to write output files from audio-splitter")
	uploadCmd.PersistentFlags().BoolVar(&uploadCfg.IsInstrumental, "instrumental", false, "specify if the audio is instrumental")
	uploadCmd.PersistentFlags().BoolVar(&uploadCfg.UseGPU, "gpu", false, "use gpu during audio splitter")
	uploadCmd.PersistentFlags().BoolVarP(&uploadCfg.DryRun, "dry_run", "D", false, "files and metadata will not be uploaded to S3 and database")
	audioCmd.AddCommand(uploadCmd)
	audioCmd.PersistentFlags().StringVarP(&audioPath, "audio", "a", "", "path of audio")
	return audioCmd
}

type audioProcessor struct {
	cfg       AudioProcessConfig
	app       internal.AppCtx
	ctx       context.Context
	container *container.CreateResponse
	mounts    []mount.Mount
	// output paths
	paths struct {
		audio string
		// stem paths for audio-separator
		stems struct {
			vocal struct {
				filename string
				path     string
			}
			instrumental struct {
				filename string
				path     string
			}
		}
		// segment directories for ffmpeg playlists and segments, deleted at the end of run
		segments struct {
			vocal        string
			instrumental string
		}
		// keyfinder-cli output
		key string
		// aubio tempo output
		tempo string
		// ffprobe duration output
		duration string
		// audiowaveform outputs
		waveform struct {
			vocal        string
			instrumental string
		}
		exif string
		// entrypoint bash script for docker container
		entrypoint string
	}
	audioFormat string
	spinner     *spinner.Spinner
}

func processAndUploadAudio(cmd *cobra.Command, args []string) {
	exitIfImage(DoesNotExist)
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)

	processor := &audioProcessor{
		cfg:     uploadCfg,
		app:     app,
		ctx:     cmd.Context(),
		spinner: spinner.New(spinner.CharSets[12], 100*time.Millisecond),
	}
	processor.spinner.Prefix = "initializing "
	processor.spinner.Start()
	processor.setupAudioSeparator()
	processor.preparePaths()
	processor.prepareMounts()
	defer processor.deleteTemps()

	processor.splitAudio()
	processor.spinner.Prefix = "initializing container "

	statusCh, errCh := processor.runContainer()
	defer removeContainer(ctx, processor.container, nil, app.Docker)
	select {
	case err := <-errCh:
		check(err)
	case <-statusCh:
		processor.spinner.Stop()
		processor.processResults(ctx)
	}
	fmt.Println(color.GreenString("goodbye!"))
}

func (p *audioProcessor) setupAudioSeparator() error {
	uvPath, err := exec.LookPath("uv")
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		log.Fatalf("uv is not installed! %v", err)
	}
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		log.Fatalf("uv is not installed! %v", err)
	}
	p.spinner.Prefix = "checking installed pip packages "
	output, err := exec.Command(uvPath, "pip", "list").Output()
	if err != nil {
		log.Fatalf("failed to check installed packages: %s", err.Error())
	}
	separatorPkg := "audio-separator"
	if !strings.Contains(string(output), separatorPkg) {
		p.spinner.Stop()
		fmt.Printf("audio-separator package is not installed, should we install it? [Y/n] ")
		conf, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ToLower(strings.TrimSpace(conf)) == "n" {
			log.Fatalf("audio separator is required for vocal and instrument split")
		}
		p.spinner.Start()
		p.spinner.Prefix = "installing audio separator package "
		pkgSpec := "audio-separator[cpu]"
		if uploadCfg.UseGPU {
			pkgSpec = "audio-separator[gpu]"
		}
		cd := exec.Command(uvPath, "pip", "install", "--system", "onnxruntime", pkgSpec)
		cd.Stdout = os.Stdout
		cd.Stderr = os.Stderr
		if err := cd.Run(); err != nil {
			log.Fatalf("failed to install %s: %s", pkgSpec, err.Error())
		}
	}
	return nil
}

func (p *audioProcessor) preparePaths() {
	var err error
	audioSeparatorOutputDirectoryNoSuffix := strings.TrimSuffix(uploadCfg.OutputDir, string(os.PathSeparator))
	hostAudioSplitByPath := strings.Split(audioPath, string(os.PathSeparator))
	hostAudioPathSplit := strings.Split(audioPath, ".")
	p.audioFormat = hostAudioPathSplit[len(hostAudioPathSplit)-1]
	if p.audioFormat == "" {
		log.Fatal("could not determine audio format from file extension")
	}
	fileName := strings.Replace(hostAudioSplitByPath[len(hostAudioSplitByPath)-1], p.audioFormat, "", -1)
	p.paths.stems.vocal.filename = fmt.Sprintf("%s_vocals", fileName)
	p.paths.stems.vocal.path = fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, p.paths.stems.vocal.filename, p.audioFormat)
	p.paths.stems.instrumental.filename = fmt.Sprintf("%s_instrumentals", fileName)
	p.paths.stems.instrumental.path = fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, p.paths.stems.instrumental.filename, p.audioFormat)

	p.paths.segments.instrumental, err = os.MkdirTemp(os.TempDir(), "strafe-instrumental-segments-*")
	check(err)
	p.paths.segments.vocal, err = os.MkdirTemp(os.TempDir(), "strafe-vocal-segments-*")
	check(err)

	p.paths.audio = audioPath
	if _, err = os.Stat(p.paths.audio); os.IsNotExist(err) {
		log.Fatalf("audio file not found: %s", p.paths.audio)
	}
	if p.paths.key, err = createTempFileReturnPath("txt"); err != nil {
		check(err)
	}
	if p.paths.tempo, err = createTempFileReturnPath("txt"); err != nil {
		check(err)
	}
	if p.paths.duration, err = createTempFileReturnPath("txt"); err != nil {
		check(err)
	}
	if p.paths.waveform.instrumental, err = createTempFileReturnPath("json"); err != nil {
		check(err)
	}
	if p.paths.waveform.vocal, err = createTempFileReturnPath("json"); err != nil {
		check(err)
	}
	if p.paths.exif, err = createTempFileReturnPath("json"); err != nil {
		check(err)
	}
	if p.paths.entrypoint, err = createTempFileReturnPath("sh"); err != nil {
		check(err)
	}
}

func (p *audioProcessor) prepareMounts() {
	bind := func(path ...string) {
		for _, pth := range path {
			if pth == "" {
				log.Fatalf("empty path found when preparing mounts")
			}
			p.mounts = append(p.mounts, mount.Mount{Type: mount.TypeBind, Source: pth, Target: pth})
		}
	}
	bind(
		p.paths.audio,
		p.paths.key,
		p.paths.tempo,
		p.paths.duration,
		p.paths.exif,
		p.paths.stems.vocal.path,
		p.paths.stems.instrumental.path,
		p.paths.segments.vocal,
		p.paths.segments.instrumental,
		p.paths.waveform.vocal,
		p.paths.waveform.instrumental,
		p.paths.entrypoint,
	)
}
func (p *audioProcessor) deleteTemps() {
	check(os.RemoveAll(p.paths.segments.instrumental))
	check(os.RemoveAll(p.paths.segments.vocal))
	check(os.Remove(p.paths.key))
	check(os.Remove(p.paths.tempo))
	check(os.Remove(p.paths.duration))
	check(os.Remove(p.paths.waveform.instrumental))
	check(os.Remove(p.paths.waveform.vocal))
	check(os.Remove(p.paths.entrypoint))
	check(os.Remove(p.paths.exif))
}

func (p *audioProcessor) splitAudio() {
	uvxPath, err := exec.LookPath("uvx")
	check(err)
	cdArgs := []string{
		"--with", "onnxruntime",
		"audio-separator",
		"-m", uploadCfg.ModelCheckpoint,
		"--output_format", strings.ToUpper(p.audioFormat),
		"--output_dir", uploadCfg.OutputDir,
		"--custom_output_names", fmt.Sprintf(`{"Vocals": "%s", "Instrumental": "%s"}`, p.paths.stems.vocal.filename, p.paths.stems.instrumental.filename),
		audioPath,
	}
	cd := exec.Command(uvxPath, cdArgs...)
	cd.Stdout = os.Stdout
	cd.Stderr = os.Stderr
	p.spinner.Stop()
	if _, err := os.Stat(p.paths.stems.vocal.path); err == nil {
		if _, err := os.Stat(p.paths.stems.instrumental.path); err == nil {
			fmt.Printf("vocal and instrumental stem files exists, should we split the audio anyways? [y/N] ")
			conf, _ := bufio.NewReader(os.Stdin).ReadString('\n')
			if strings.ToLower(strings.TrimSpace(conf)) == "y" {
				if err := cd.Run(); err != nil {
					log.Fatalf("failed to split audio %s with the error %s", audioPath, err.Error())
				}
			}
		}
	} else {
		if err := cd.Run(); err != nil {
			log.Fatalf("failed to split audio %s with the error %s", audioPath, err.Error())
		}
	}
	p.spinner.Start()
}

func (p *audioProcessor) runContainer() (<-chan container.WaitResponse, <-chan error) {
	scripts := []string{
		fmt.Sprintf(`exiftool "%s" -json > "%s"`, audioPath, p.paths.exif),
		fmt.Sprintf(`audiowaveform -i "%s" --pixels-per-second %d --output-format json > "%s"`, p.paths.stems.vocal.path, uploadCfg.WaveformPPS, p.paths.waveform.vocal),
		fmt.Sprintf(`audiowaveform -i "%s" --pixels-per-second %d --output-format json > "%s"`, p.paths.stems.instrumental.path, uploadCfg.WaveformPPS, p.paths.waveform.instrumental),
		fmt.Sprintf(`aubio tempo -i "%s" > "%s"`, audioPath, p.paths.tempo),
		fmt.Sprintf(`keyfinder-cli "%s" > "%s"`, audioPath, p.paths.key),
		fmt.Sprintf(`ffmpeg -i "%s" -c:a aac -b:a 320k -f segment -segment_time 40 -segment_list "%s/playlist.m3u8" -segment_format mpegts "%s/%%03d.ts"`, p.paths.stems.instrumental.path, p.paths.segments.instrumental, p.paths.segments.instrumental),
		fmt.Sprintf(`ffmpeg -i "%s" -c:a aac -b:a 320k -f segment -segment_time 40 -segment_list "%s/playlist.m3u8" -segment_format mpegts "%s/%%03d.ts"`, p.paths.stems.vocal.path, p.paths.segments.vocal, p.paths.segments.vocal),
		fmt.Sprintf(`ffprobe -i "%s" -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 -v error > "%s"`, audioPath, p.paths.duration),
	}
	var err error
	err = os.WriteFile(p.paths.entrypoint, []byte(strings.Join(scripts, "\n")), 0755)
	check(err)
	log.Infof("creating container")
	p.spinner.Prefix = "creating container "
	resp, err := p.app.Docker.ContainerCreate(p.ctx, &container.Config{
		Image:        getImageTag(),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          []string{"/bin/bash", p.paths.entrypoint},
	}, &container.HostConfig{Mounts: p.mounts}, nil, nil, "")
	p.container = &resp
	check(err)
	log.Infof("starting container")
	p.spinner.Prefix = "starting container"
	startContainer(p.ctx, &resp, nil, p.app.Docker)
	stdout, err := p.app.Docker.ContainerLogs(p.ctx, resp.ID, container.LogsOptions{ShowStdout: true, Follow: true, ShowStderr: true, Timestamps: true})
	check(err)
	go func(out io.ReadCloser, s *spinner.Spinner) {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			line := scanner.Text()
			s.Stop()
			fmt.Println(line)
			s.Start()
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("error scanning logs: %v\n", err)
		}
	}(stdout, p.spinner)
	p.spinner.Prefix = "waiting for container to finish"
	return p.app.Docker.ContainerWait(p.ctx, resp.ID, container.WaitConditionNotRunning)
}

func (p *audioProcessor) processResults(ctx context.Context) error {
	exifInfoBytes, err := os.ReadFile(p.paths.exif)
	check(err)
	var exifInfo internal.ExifInfo
	if err = json.Unmarshal(fastjson.MustParseBytes(exifInfoBytes).GetArray()[0].GetObject().MarshalTo(nil), &exifInfo); err != nil {
		check(err)
	}

	log.Info(color.CyanString(`finished processing %s - %s [%d - %s]`, exifInfo.Artist, exifInfo.Title, exifInfo.Year, exifInfo.Album))

	keyBytes, err := os.ReadFile(p.paths.key)
	check(err)
	tempoBytes, err := os.ReadFile(p.paths.tempo)
	check(err)
	var tempo pgtype.Numeric
	if err = tempo.Scan(strings.TrimSpace(strings.Replace(string(tempoBytes), "bpm", "", -1))); err != nil {
		check(err)
	}
	instrumentalSegmentFiles, err := os.ReadDir(p.paths.segments.instrumental)
	check(err)
	vocalSegmentFiles, err := os.ReadDir(p.paths.segments.vocal)
	check(err)
	instrumentalWFBytes, err := os.ReadFile(p.paths.waveform.instrumental)
	check(err)
	vocalWFBytes, err := os.ReadFile(p.paths.waveform.vocal)
	check(err)
	instrumentalWF := fastjson.MustParseBytes(instrumentalWFBytes).GetObject().Get("data").MarshalTo(nil)
	vocalWF := fastjson.MustParseBytes(vocalWFBytes).GetObject().Get("data").MarshalTo(nil)
	compressedInstrumentalWF, err := internal.CompressJSON(instrumentalWF)
	check(err)
	compressedVocalWF, err := internal.CompressJSON(vocalWF)
	check(err)
	var duration pgtype.Numeric
	durationBytes, err := os.ReadFile(p.paths.duration)
	check(err)
	if err = duration.Scan(strings.TrimSpace(string(durationBytes))); err != nil {
		check(err)
	}
	if !uploadCfg.DryRun {
		var albumId string
		tx, err := p.app.Conn.BeginTx(ctx, pgx.TxOptions{})
		defer tx.Rollback(ctx)
		check(err)
		qtx := p.app.DB.WithTx(tx)
		albumId, err = qtx.GetAlbumIDByName(ctx, pgtype.Text{String: exifInfo.Album})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				albumId, err = qtx.InsertAlbum(ctx, db.InsertAlbumParams{
					ID:    uuid.NewString(),
					Name:  pgtype.Text{String: exifInfo.Album},
					Cover: pgtype.Text{String: ""},
				})
				check(err)
			} else {
				log.Fatal(err)
			}
		}
		err = tx.Commit(ctx)
		check(err)
		p.app.DB.InsertTrack(ctx, db.InsertTrackParams{
			ID:                     uuid.NewString(),
			VocalFolderPath:        pgtype.Text{String: ""},
			InstrumentalFolderPath: pgtype.Text{String: ""},
			AlbumID:                pgtype.Text{String: albumId},
			TotalDuration:          duration,
			Info:                   exifInfoBytes,
			Instrumental:           pgtype.Bool{Bool: uploadCfg.IsInstrumental},
			Tempo:                  tempo,
			Key:                    pgtype.Text{String: string(keyBytes)},
			VocalWaveform:          compressedVocalWF,
			InstrumentalWaveform:   compressedInstrumentalWF,
			AlbumName:              pgtype.Text{String: exifInfo.Album},
		})
	} else {
		log.WithFields(log.Fields{
			"instrumental_segment_file_count": len(instrumentalSegmentFiles),
			"vocal_segment_file_count":        len(vocalSegmentFiles),
			"tempo":                           string(tempoBytes),
			"song_written_with_key":           string(keyBytes),
			"exif_info":                       string(exifInfoBytes),
		}).Debug("dry run :: results")
	}
	return nil
}
