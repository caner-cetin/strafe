package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strafe/internal"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/fatih/color"
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

var (
	uploadCmd = &cobra.Command{
		Use:   "upload -a audio",
		Short: "processes and uploads audio",
		Long:  `processes and uploads audio file given with -a / --audio flag`,
		Run:   WrapCommandWithResources(processAndUploadAudio, ResourceConfig{Resources: []ResourceType{ResourceDocker}}),
	}
	modelDownloadFolder           string
	modelCheckpointName           string
	audioSeparatorOutputDirectory string
	waveformPixelsPerSecond       int32
	useGPU                        bool
)

func getAudioRootCmd() *cobra.Command {
	uploadCmd.PersistentFlags().Int32VarP(&waveformPixelsPerSecond, "pps", "P", 100, "waveform zoom level (pixels per second)")
	uploadCmd.PersistentFlags().StringVar(&modelCheckpointName, "model", "mel_band_roformer_karaoke_aufr33_viperx_sdr_10.1956.ckpt", "model name for audio splitter, see https://raw.githubusercontent.com/nomadkaraoke/python-audio-separator/refs/heads/main/audio_separator/models.json for full list")
	uploadCmd.PersistentFlags().StringVar(&modelDownloadFolder, "model_file_directory", "/tmp/audio-separator-models/", "model download folder / file directory on the host machine")
	uploadCmd.PersistentFlags().StringVar(&audioSeparatorOutputDirectory, "audio_output_directory", "/tmp/strafe-audio-separator-audio/", "directory to write output files from audio-splitter")

	uploadCmd.PersistentFlags().BoolVar(&useGPU, "gpu", false, "use gpu during audio splitter")
	audioCmd.AddCommand(uploadCmd)
	audioCmd.PersistentFlags().StringVarP(&audioPath, "audio", "a", "", "path of audio")
	return audioCmd
}

func processAndUploadAudio(cmd *cobra.Command, args []string) {
	exitIfImage(DoesNotExist)
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)

	var mounts []mount.Mount

	uvPath, err := exec.LookPath("uv")
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		log.Fatalf("uv is not installed! %v", err)
	}
	uvxPath, err := exec.LookPath("uvx")
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		log.Fatalf("uv is not installed! %v", err)
	}
	s := spinner.New(spinner.CharSets[12], 100*time.Millisecond)
	s.Prefix = "initializing "
	s.Start()
	defer s.Stop()
	s.Prefix = "checking installed pip packages "
	output, err := exec.Command(uvPath, "pip", "list").Output()
	if err != nil {
		log.Fatalf("failed to check installed packages: %s", err.Error())
	}
	separatorPkg := "audio-separator"
	if !strings.Contains(string(output), separatorPkg) {
		s.Stop()
		fmt.Printf("audio-separator package is not installed, should we install it? [Y/n] ")
		conf, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.ToLower(strings.TrimSpace(conf)) == "n" {
			log.Fatalf("audio separator is required for vocal and instrument split")
		}
		s.Start()
		s.Prefix = "installing audio separator package "
		pkgSpec := "audio-separator[cpu]"
		if useGPU {
			pkgSpec = "audio-separator[gpu]"
		}
		cd := exec.Command(uvPath, "pip", "install", "--system", "onnxruntime", pkgSpec)
		cd.Stdout = os.Stdout
		cd.Stderr = os.Stderr
		if err := cd.Run(); err != nil {
			log.Fatalf("failed to install %s: %s", pkgSpec, err.Error())
		}
	}
	audioSeparatorOutputDirectoryNoSuffix := strings.TrimSuffix(audioSeparatorOutputDirectory, string(os.PathSeparator))
	hostAudioSplitByPath := strings.Split(audioPath, string(os.PathSeparator))
	hostAudioPathSplit := strings.Split(audioPath, ".")
	hostAudioPathExt := hostAudioPathSplit[len(hostAudioPathSplit)-1]
	fileName := strings.Replace(hostAudioSplitByPath[len(hostAudioSplitByPath)-1], hostAudioPathExt, "", -1)
	vocalStemFilename := fmt.Sprintf("%s_vocals", fileName)
	instrumentalStemFilename := fmt.Sprintf("%s_instrumentals", fileName)
	vocalStemFilepath := fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, vocalStemFilename, hostAudioPathExt)
	instrumentalStemFilepath := fmt.Sprintf("%s/%s.%s", audioSeparatorOutputDirectoryNoSuffix, instrumentalStemFilename, hostAudioPathExt)
	cdArgs := []string{
		"--with", "onnxruntime",
		"audio-separator",
		"-m", modelCheckpointName,
		"--output_format", strings.ToUpper(hostAudioPathExt),
		"--output_dir", audioSeparatorOutputDirectory,
		"--custom_output_names", fmt.Sprintf(`{"Vocals": "%s", "Instrumental": "%s"}`, vocalStemFilename, instrumentalStemFilename),
		audioPath,
	}
	cd := exec.Command(uvxPath, cdArgs...)
	cd.Stdout = os.Stdout
	cd.Stderr = os.Stderr
	s.Stop()
	if _, err := os.Stat(vocalStemFilepath); err == nil {
		if _, err := os.Stat(instrumentalStemFilepath); err == nil {
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
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: audioPath, Target: audioPath})
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: vocalStemFilepath, Target: vocalStemFilepath})
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: instrumentalStemFilepath, Target: instrumentalStemFilepath})

	s.Start()
	s.Prefix = "initializing container "

	hostExifInfo, err := createTempFile("json")
	check(err)
	defer hostExifInfo.Close()
	defer os.Remove(hostExifInfo.Name())
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: hostExifInfo.Name(), Target: hostExifInfo.Name()})

	hostInstrumentalWaveformOutput, err := createTempFile("json")
	check(err)
	defer hostInstrumentalWaveformOutput.Close()
	defer os.Remove(hostInstrumentalWaveformOutput.Name())
	hostVocalWaveformOutput, err := createTempFile("json")
	check(err)
	defer hostVocalWaveformOutput.Close()
	defer os.Remove(hostVocalWaveformOutput.Name())
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: hostVocalWaveformOutput.Name(), Target: hostVocalWaveformOutput.Name()})
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: hostInstrumentalWaveformOutput.Name(), Target: hostInstrumentalWaveformOutput.Name()})

	scriptFile, err := createTempFile("sh")
	check(err)
	scripts := []string{
		fmt.Sprintf(`exiftool "%s" -json > "%s"`, audioPath, hostExifInfo.Name()),
		fmt.Sprintf(`audiowaveform -i "%s" --pixels-per-second %d --output-format json > "%s"`, vocalStemFilepath, waveformPixelsPerSecond, hostVocalWaveformOutput.Name()),
		fmt.Sprintf(`audiowaveform -i "%s" --pixels-per-second %d --output-format json > "%s"`, instrumentalStemFilepath, waveformPixelsPerSecond, hostInstrumentalWaveformOutput.Name()),
	}
	_, err = scriptFile.WriteString(strings.Join(scripts, "\n"))
	check(err)
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: scriptFile.Name(), Target: scriptFile.Name()})

	log.Infof("creating container")
	s.Prefix = "creating container "
	resp, err := app.Docker.ContainerCreate(ctx, &container.Config{
		Image:        getImageTag(),
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          []string{"/bin/bash", scriptFile.Name()},
	}, &container.HostConfig{Mounts: mounts}, nil, nil, "")
	check(err)
	defer removeContainer(ctx, &resp, nil, app.Docker)
	log.Infof("starting container")
	s.Prefix = "starting container"
	startContainer(ctx, &resp, nil, app.Docker)
	stdout, err := app.Docker.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, Follow: true, ShowStderr: true, Timestamps: true})
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
	}(stdout, s)
	s.Prefix = "waiting for container to finish"
	statusCh, errCh := app.Docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		check(err)
	case <-statusCh:
		s.Stop()
		file, err := os.OpenFile(hostExifInfo.Name(), os.O_RDONLY, os.ModeTemporary)
		check(err)
		fileContents, err := io.ReadAll(file)
		check(err)
		var audioInfo internal.ExifInfo
		if err = json.Unmarshal(fastjson.MustParseBytes(fileContents).GetArray()[0].GetObject().MarshalTo(nil), &audioInfo); err != nil {
			check(err)
		}
		log.Info(color.CyanString(`finished processing %s - %s [%d - %s]`, audioInfo.Artist, audioInfo.Title, audioInfo.Year, audioInfo.Album))
	}
	fmt.Println(color.GreenString("goodbye!"))
}
