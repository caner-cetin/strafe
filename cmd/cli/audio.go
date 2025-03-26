package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/caner-cetin/strafe/internal"
	"github.com/caner-cetin/strafe/pkg/db"
	"github.com/rs/zerolog/log"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/valyala/fastjson"
)

var (
	audioPath string
	audioCmd  = &cobra.Command{
		Use:   "audio",
		Short: "audio related commands",
	}
)

type UploadConfig struct {
	ModelCheckpoint  string
	ModelDownloadDir string
	// output directory for audio separator stems,
	// vocals and instrumentals will be directories under this directory
	OutputDir      string
	WaveformPPS    int32
	UseGPU         bool
	IsInstrumental bool
	DryRun         bool
	CoverArtPath   string
}

type ModelsConfig struct {
	Source string
}

var (
	uploadCmd = &cobra.Command{
		Use:   "upload -i audio -c cover",
		Short: "processes and uploads audio",
		Long:  `processes and uploads audio file given with -i / --input flag. requires strafe docker image.`,
		Run:   WrapCommandWithResources(processAndUploadAudio, ResourceConfig{Resources: []ResourceType{ResourceDocker, ResourceDatabase, ResourceS3}}),
	}
	uploadCfg = UploadConfig{}
	modelsCmd = &cobra.Command{
		Use:   "models",
		Short: "lists available audio separator models",
		Long: fmt.Sprintf(`lists available audio separator models, useful for command %s. Default model for the %s command is Mel-Roformer-Karaoke-Aufr33-Viperx
source %s`, color.MagentaString("audio"), color.MagentaString("audio"), color.WhiteString("https://raw.githubusercontent.com/nomadkaraoke/python-audio-separator/refs/heads/main/audio_separator/models.json")),
		Run: listModels,
	}
	modelsCfg = ModelsConfig{}
)

func getAudioRootCmd() *cobra.Command {
	uploadCmd.PersistentFlags().Int32VarP(&uploadCfg.WaveformPPS, "pps", "P", 100, "waveform zoom level (pixels per second)")
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.ModelCheckpoint, "model", "mel_band_roformer_karaoke_aufr33_viperx_sdr_10.1956.ckpt", fmt.Sprintf("model name for audio splitter, see %s for full list", color.MagentaString("strafe audio models")))
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.ModelDownloadDir, "model_file_directory", "/tmp/audio-separator-models/", "model download folder / file directory on the host machine")
	uploadCmd.PersistentFlags().StringVar(&uploadCfg.OutputDir, "audio_output_directory", "/tmp/strafe-audio-separator-audio/", "directory to write output files from audio-splitter")
	uploadCmd.PersistentFlags().BoolVar(&uploadCfg.IsInstrumental, "instrumental", false, "specify if the audio is instrumental")
	uploadCmd.PersistentFlags().BoolVar(&uploadCfg.UseGPU, "gpu", false, "use gpu during audio separation")
	uploadCmd.PersistentFlags().BoolVarP(&uploadCfg.DryRun, "dry_run", "d", false, "files and metadata will not be uploaded to S3 and database")
	uploadCmd.PersistentFlags().StringVarP(&uploadCfg.CoverArtPath, "cover_art", "c", "", "cover art for the tracks album, required if album does not exist yet.")

	modelsCmd.PersistentFlags().StringVar(&modelsCfg.Source, "src", "https://raw.githubusercontent.com/nomadkaraoke/python-audio-separator/refs/heads/main/audio_separator/models.json", "model source")

	audioCmd.AddCommand(uploadCmd)
	audioCmd.AddCommand(modelsCmd)
	audioCmd.PersistentFlags().StringVarP(&audioPath, "input", "i", "", "path of audio")
	return audioCmd
}

type audioProcessor struct {
	cfg       UploadConfig
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
			// s3 upload paths for segments
			s3 struct {
				vocal        string
				instrumental string
			}
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
	// output of exifinfo
	info        internal.ExifInfo
	db_record   db.InsertTrackParams
	audioFormat string
	spinner     *spinner.Spinner
	conditions  struct {
		// set to true if the album is uploaded for the first time
		// and the album is inserted at the same time with track is inserted
		//
		// if this is set to true and CoverArtPath in upload config is empty
		// command will throw error and exit
		shouldUploadCoverArt bool
	}
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
	defer processor.spinner.Stop()
	if err := processor.setupAudioSeparator(); err != nil {
		log.Error().Err(err).Msg("failed to setup audio separator")
		return
	}
	if err := processor.preparePaths(); err != nil {
		log.Error().Err(err).Msg("failed to prepare mount paths")
		return
	}
	processor.prepareMounts()
	defer func() {
		if err := processor.deleteTemps(); err != nil {
			log.Error().Err(err).Msg("failed to delete temporary files")
			return
		}
	}()

	if err := processor.splitAudio(); err != nil {
		log.Error().Err(err).Msg("failed to split audio file")
		return
	}
	processor.spinner.Prefix = "initializing container "

	statusCh, errCh, err := processor.runContainer()
	if err != nil {
		processor.spinner.Stop()
		log.Error().Err(err).Msg("error in container")
		return
	}
	defer func() {
		if err := removeContainer(ctx, processor.container, nil, app.Docker); err != nil {
			log.Error().Err(err).Msg("failed to remove container")
			return
		}
	}()
	select {
	case err := <-errCh:
		if err != nil {
			processor.spinner.Stop()
			log.Error().Err(err).Msg("error in container")
			return
		}
	case <-statusCh:
		processor.spinner.Stop()
		if err := processor.processResults(ctx); err != nil {
			log.Error().Err(err).Msg("failed to process results")
		}
	}
	fmt.Println(color.GreenString("goodbye!"))
}

func (p *audioProcessor) setupAudioSeparator() error {
	uvPath, err := exec.LookPath("uv")
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("uv is not installed! %w", err)
	}
	p.spinner.Prefix = "checking installed pip packages "
	output, err := exec.Command(uvPath, "pip", "list").Output()
	if err != nil {
		return fmt.Errorf("installed package check failed: %w", err)
	}
	separatorPkg := "audio-separator"
	if !strings.Contains(string(output), separatorPkg) {
		p.spinner.Stop()
		fmt.Printf("audio-separator package is not installed, should we install it? [Y/n] ")
		conf, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ToLower(strings.TrimSpace(conf)) == "n" {
			return fmt.Errorf("audio separator is required for vocal and instrument split")
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
			return fmt.Errorf("failed to install %s: %w", pkgSpec, err)
		}
	}
	return nil
}

func (p *audioProcessor) preparePaths() error {
	var err error
	audioSeparatorOutputDirectoryNoSuffix := strings.TrimSuffix(uploadCfg.OutputDir, string(os.PathSeparator))
	hostAudioSplitByPath := strings.Split(audioPath, string(os.PathSeparator))
	hostAudioPathSplit := strings.Split(audioPath, ".")
	p.audioFormat = hostAudioPathSplit[len(hostAudioPathSplit)-1]
	if p.audioFormat == "" {
		return fmt.Errorf("cannot determine audio format from file extension")
	}
	fileName := strings.ReplaceAll(hostAudioSplitByPath[len(hostAudioSplitByPath)-1], p.audioFormat, "")
	p.paths.stems.vocal.filename = fmt.Sprintf("%s_vocals", fileName)
	p.paths.stems.vocal.path = fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, p.paths.stems.vocal.filename, p.audioFormat)
	p.paths.stems.instrumental.filename = fmt.Sprintf("%s_instrumentals", fileName)
	p.paths.stems.instrumental.path = fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, p.paths.stems.instrumental.filename, p.audioFormat)

	p.paths.segments.instrumental, err = os.MkdirTemp(os.TempDir(), "strafe-instrumental-segments-*")
	if err != nil {
		return fmt.Errorf("failed to create instrumental segments directory: %w", err)
	}
	p.paths.segments.vocal, err = os.MkdirTemp(os.TempDir(), "strafe-vocal-segments-*")
	if err != nil {
		return fmt.Errorf("failed to create vocal segments directory: %w", err)
	}

	p.paths.audio = audioPath
	if _, err = os.Stat(p.paths.audio); os.IsNotExist(err) {
		return fmt.Errorf("audio file %s not found: %w", p.paths.audio, err)
	}
	if p.paths.key, err = createTempFileReturnPath("txt"); err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	if p.paths.tempo, err = createTempFileReturnPath("txt"); err != nil {
		return fmt.Errorf("failed to create tempo file: %w", err)
	}
	if p.paths.duration, err = createTempFileReturnPath("txt"); err != nil {
		return fmt.Errorf("failed to create duration file: %w", err)
	}
	if p.paths.waveform.instrumental, err = createTempFileReturnPath("json"); err != nil {
		return fmt.Errorf("failed to create instrumental waveform file: %w", err)
	}
	if p.paths.waveform.vocal, err = createTempFileReturnPath("json"); err != nil {
		return fmt.Errorf("failed to create vocal waveform file: %w", err)
	}
	if p.paths.exif, err = createTempFileReturnPath("json"); err != nil {
		return fmt.Errorf("failed to create exif file: %w", err)
	}
	if p.paths.entrypoint, err = createTempFileReturnPath("sh"); err != nil {
		return fmt.Errorf("failed to create entrypoint file: %w", err)
	}
	return nil
}

func (p *audioProcessor) prepareMounts() {
	bind := func(path ...string) {
		for _, pth := range path {
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
func (p *audioProcessor) deleteTemps() error {
	var err error
	if err = os.RemoveAll(p.paths.segments.instrumental); err != nil {
		return fmt.Errorf("failed to remove instrumental segments directory: %w", err)
	}
	if err = os.RemoveAll(p.paths.segments.vocal); err != nil {
		return fmt.Errorf("failed to remove vocal segments directory: %w", err)
	}
	if err = os.Remove(p.paths.key); err != nil {
		return fmt.Errorf("failed to remove key file: %w", err)
	}
	if err = os.Remove(p.paths.tempo); err != nil {
		return fmt.Errorf("failed to remove tempo file: %w", err)
	}
	if err = os.Remove(p.paths.duration); err != nil {
		return fmt.Errorf("failed to remove duration file: %w", err)
	}
	if err = os.Remove(p.paths.waveform.instrumental); err != nil {
		return fmt.Errorf("failed to remove instrumental waveform file: %w", err)
	}
	if err = os.Remove(p.paths.waveform.vocal); err != nil {
		return fmt.Errorf("failed to remove vocal waveform file: %w", err)
	}
	if err = os.Remove(p.paths.entrypoint); err != nil {
		return fmt.Errorf("failed to remove entrypoint file: %w", err)
	}
	if err = os.Remove(p.paths.exif); err != nil {
		return fmt.Errorf("failed to remove exif file: %w", err)
	}
	return nil
}

func (p *audioProcessor) splitAudio() error {
	uvxPath, err := exec.LookPath("uvx")
	if err != nil {
		return fmt.Errorf("uvx is not installed: %w", err)
	}
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
					return fmt.Errorf("failed to split audio %s: %w", audioPath, err)
				}
			}
		}
	} else {
		if err := cd.Run(); err != nil {
			return fmt.Errorf("failed to split audio %s: %w", audioPath, err)
		}
	}
	p.spinner.Start()
	return nil
}

func (p *audioProcessor) runContainer() (<-chan container.WaitResponse, <-chan error, error) {
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
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write entrypoint script: %w", err)
	}
	p.spinner.Prefix = "creating container "
	resp, err := p.app.Docker.ContainerCreate(p.ctx, &container.Config{
		Image:        getImageTag(),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          []string{"/bin/bash", p.paths.entrypoint},
	}, &container.HostConfig{Mounts: p.mounts}, nil, nil, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create container: %w", err)
	}
	p.container = &resp
	p.spinner.Prefix = "starting container"
	if err := startContainer(p.ctx, &resp, nil, p.app.Docker); err != nil {
		return nil, nil, fmt.Errorf("failed to start the container: %w", err)
	}
	stdout, err := p.app.Docker.ContainerLogs(p.ctx, resp.ID, container.LogsOptions{ShowStdout: true, Follow: true, ShowStderr: true, Timestamps: true})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container logs: %w", err)
	}
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
	statusCh, errCh := p.app.Docker.ContainerWait(p.ctx, resp.ID, container.WaitConditionNotRunning)
	return statusCh, errCh, nil
}

func (p *audioProcessor) processResults(ctx context.Context) error {
	if err := p.loadExifInfo(); err != nil {
		return fmt.Errorf("failed to load exif info: %w", err)
	}
	log.Info().
		Str("artist", p.info.Artist).
		Str("title", p.info.Title).
		Int("year", p.info.Year).
		Str("album", p.info.Album).
		Msg("processing audio")
	if err := p.loadWaveforms(); err != nil {
		return fmt.Errorf("failed to load waveforms: %w", err)
	}
	if err := p.loadDuration(); err != nil {
		return fmt.Errorf("failed to load duration: %w", err)
	}
	if err := p.loadKey(); err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}
	if err := p.loadTempo(); err != nil {
		return fmt.Errorf("failed to load tempo: %w", err)
	}
	if !uploadCfg.DryRun {
		if err := p.loadOrCreateAlbum(ctx); err != nil {
			return fmt.Errorf("failed to load or create album: %w", err)
		}
		p.db_record.ID = uuid.NewString()
		p.db_record.Instrumental = pgtype.Bool{Bool: uploadCfg.IsInstrumental, Valid: true}
		p.db_record.AlbumName = pgtype.Text{String: p.info.Album, Valid: true}
		if err := p.upload(); err != nil {
			return fmt.Errorf("failed to upload: %w", err)
		}
		if err := p.app.DB.InsertTrack(ctx, p.db_record); err != nil {
			return fmt.Errorf("failed to insert track: %w", err)
		}
	}
	return nil
}

func (p *audioProcessor) loadExifInfo() error {
	// 1 element array, as we pass one single audio
	exifInfoArrayBytes, err := os.ReadFile(p.paths.exif)
	if err != nil {
		return fmt.Errorf("failed to read exif info: %w", err)
	}
	exifInfoObjectBytes := fastjson.MustParseBytes(exifInfoArrayBytes).GetArray()[0].GetObject().MarshalTo(nil)
	var exifInfo internal.ExifInfo
	if err = json.Unmarshal(exifInfoObjectBytes, &exifInfo); err != nil {
		return fmt.Errorf("failed to unmarshal exif info: %w", err)
	}
	p.info = exifInfo
	p.db_record.Info = exifInfoObjectBytes
	return nil
}

func (p *audioProcessor) loadWaveforms() error {
	instrumentalWFBytes, err := os.ReadFile(p.paths.waveform.instrumental)
	if err != nil {
		return fmt.Errorf("failed to read instrumental waveform: %w", err)
	}
	vocalWFBytes, err := os.ReadFile(p.paths.waveform.vocal)
	if err != nil {
		return fmt.Errorf("failed to read vocal waveform: %w", err)
	}
	instrumentalWF := fastjson.MustParseBytes(instrumentalWFBytes).GetObject().Get("data").MarshalTo(nil)
	vocalWF := fastjson.MustParseBytes(vocalWFBytes).GetObject().Get("data").MarshalTo(nil)
	p.db_record.InstrumentalWaveform, err = internal.CompressJSON(instrumentalWF)
	if err != nil {
		return fmt.Errorf("failed to compress instrumental waveform: %w", err)
	}
	p.db_record.VocalWaveform, err = internal.CompressJSON(vocalWF)
	if err != nil {
		return fmt.Errorf("failed to compress vocal waveform: %w", err)
	}
	return nil
}

func (p *audioProcessor) loadDuration() error {
	var duration pgtype.Numeric
	durationBytes, err := os.ReadFile(p.paths.duration)
	if err != nil {
		return fmt.Errorf("failed to read duration: %w", err)
	}
	if err = duration.Scan(strings.TrimSpace(string(durationBytes))); err != nil {
		return fmt.Errorf("failed to scan duration: %w", err)
	}
	p.db_record.TotalDuration = duration
	return nil
}

func (p *audioProcessor) loadKey() error {
	keyBytes, err := os.ReadFile(p.paths.key)
	if err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}
	p.db_record.Key = pgtype.Text{String: strings.TrimSpace(strings.ReplaceAll(string(keyBytes), "\n", "")), Valid: true}
	return nil
}

func (p *audioProcessor) loadTempo() error {
	tempoBytes, err := os.ReadFile(p.paths.tempo)
	if err != nil {
		return fmt.Errorf("failed to read tempo: %w", err)
	}
	var tempo pgtype.Numeric
	if err = tempo.Scan(strings.TrimSpace(strings.ReplaceAll(string(tempoBytes), "bpm", ""))); err != nil {
		return fmt.Errorf("failed to scan tempo: %w", err)
	}
	p.db_record.Tempo = tempo
	return nil
}

// inserts album if the name does not exist
func (p *audioProcessor) loadOrCreateAlbum(ctx context.Context) error {
	var albumId string
	tx, err := p.app.Conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error().Err(err).Msg("failed to rollback transaction")
		}
	}()
	qtx := p.app.DB.WithTx(tx)
	albumId, err = qtx.GetAlbumIDByNameAndArtist(ctx,
		db.GetAlbumIDByNameAndArtistParams{
			Name:   pgtype.Text{String: p.info.Album, Valid: true},
			Artist: pgtype.Text{String: p.info.Artist, Valid: true}},
	)
	p.conditions.shouldUploadCoverArt = false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if p.cfg.CoverArtPath == "" {
				return fmt.Errorf("corresponding album does not exist in database, and cover art path is not given with the command")
			}
			albumId, err = qtx.InsertAlbum(ctx, db.InsertAlbumParams{
				ID:     uuid.NewString(),
				Name:   pgtype.Text{String: p.info.Album, Valid: true},
				Cover:  pgtype.Text{String: p.coverArtS3Key(), Valid: true},
				Artist: pgtype.Text{String: p.info.Artist, Valid: true},
			})
			if err != nil {
				return fmt.Errorf("failed to insert album: %w", err)
			}
			err = tx.Commit(ctx)
			if err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
			p.conditions.shouldUploadCoverArt = true
		} else {
			return fmt.Errorf("failed to get album id: %w", err)
		}
	}
	p.db_record.AlbumID = pgtype.Text{String: albumId, Valid: true}
	p.db_record.AlbumName = pgtype.Text{String: p.info.Album, Valid: true}
	return nil
}

func (p *audioProcessor) coverArtS3Key() string {
	coverArtPathSplit := strings.Split(uploadCfg.CoverArtPath, string(os.PathSeparator))
	// last trim suffix for sanity check
	return fmt.Sprintf("%s/%s/%s", p.info.Artist, p.info.Album, strings.TrimSuffix(coverArtPathSplit[len(coverArtPathSplit)-1], string(os.PathSeparator)))
}
func (p *audioProcessor) upload() error {
	if !viper.IsSet(internal.S3_BUCKET_NAME) {
		return fmt.Errorf("s3 bucket name is not set")
	}
	vocals, err := os.ReadDir(p.paths.segments.vocal)
	if err != nil {
		return fmt.Errorf("failed to read vocal segments directory: %w", err)
	}
	instrumentals, err := os.ReadDir(p.paths.segments.instrumental)
	if err != nil {
		return fmt.Errorf("failed to read instrumental segments directory: %w", err)
	}
	uploadSegments := func(s3Folder string, files []os.DirEntry) error {
		for _, segment := range files {
			var s3path = fmt.Sprintf("%s/%s/%s/%s/%s", p.info.Artist, p.info.Album, p.info.Title, s3Folder, segment.Name())
			var segmentBytes []byte
			if s3Folder == "vocal" {
				p.paths.segments.s3.vocal = s3path
				p.db_record.VocalFolderPath = pgtype.Text{String: s3path, Valid: true}
				segmentBytes, err = os.ReadFile(fmt.Sprintf("%s/%s", strings.TrimSuffix(p.paths.segments.vocal, string(os.PathSeparator)), segment.Name()))
			} else {
				p.paths.segments.s3.instrumental = s3path
				p.db_record.InstrumentalFolderPath = pgtype.Text{String: s3path, Valid: true}
				segmentBytes, err = os.ReadFile(fmt.Sprintf("%s/%s", strings.TrimSuffix(p.paths.segments.instrumental, string(os.PathSeparator)), segment.Name()))
			}
			if err != nil {
				return fmt.Errorf("failed to read segment file: %w", err)
			}
			_, err = p.app.UploadObject(p.ctx, viper.GetString(internal.S3_BUCKET_NAME), s3path, segmentBytes)
			if err != nil {
				return fmt.Errorf("failed to upload segment to s3: %w", err)
			}
		}
		return nil
	}
	if !uploadCfg.IsInstrumental {
		if err := uploadSegments("vocal", vocals); err != nil {
			return fmt.Errorf("failed to upload vocal segments: %w", err)
		}
	}
	if err := uploadSegments("instrumental", instrumentals); err != nil {
		return fmt.Errorf("failed to upload instrumental segments: %w", err)
	}
	if p.conditions.shouldUploadCoverArt {
		coverArtBytes, err := os.ReadFile(uploadCfg.CoverArtPath)
		if err != nil {
			return fmt.Errorf("failed to read cover art file: %w", err)
		}
		_, err = p.app.UploadObject(p.ctx, viper.GetString(internal.S3_BUCKET_NAME), p.coverArtS3Key(), coverArtBytes)
		if err != nil {
			return fmt.Errorf("failed to upload cover art to s3: %w", err)
		}
	}
	return nil
}

func listModels(cmd *cobra.Command, args []string) {
	req, err := http.NewRequest(http.MethodGet, modelsCfg.Source, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to create request to the source")
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("failed to establish connection to the source")
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", resp.StatusCode).Msg("cannot establish connection to the source")
		return
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read response body for models")
		return
	}
	var data = fastjson.MustParse(string(respBytes))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)
	t.AppendHeader(table.Row{"Category", "Model Name", "File Name", "Config/Related File"})
	processModels := func(category string, objName string) {
		data.GetObject(objName).Visit(func(key []byte, v *fastjson.Value) {
			modelName := strings.TrimSpace(strings.Split(string(key), ":")[1])

			switch v.Type() {
			case fastjson.TypeObject:
				v.GetObject().Visit(func(subKey []byte, subV *fastjson.Value) {
					t.AppendRow(table.Row{
						category,
						modelName,
						string(subKey),
						subV.String(), // config file
					})
				})
			case fastjson.TypeString:
				t.AppendRow(table.Row{
					category,
					"",
					modelName,
					v.String(), // again, config file
				})
			}
		})
	}

	processModels("VR", "vr_download_list")
	processModels("MDX", "mdx_download_list")
	processModels("MDX23C", "mdx23c_download_list")
	processModels("Roformer", "roformer_download_list")

	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Category", WidthMax: 10},
		{Name: "Model Name", WidthMax: 80},
		{Name: "File Name", WidthMax: 70},
		{Name: "Config/Related File", WidthMax: 70},
	})

	t.SetRowPainter(func(row table.Row) text.Colors {
		if row[0] != "" {
			return text.Colors{text.BgHiBlack}
		}
		return nil
	})

	t.Render()
}
