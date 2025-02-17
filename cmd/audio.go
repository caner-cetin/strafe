package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strafe/internal"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		Run:   internal.WrapCommandWithResources(processAndUploadAudio, internal.ResourceConfig{Resources: []internal.ResourceType{internal.ResourceDocker}}),
	}
	waveformPixelsPerSecond int32
	modelCheckpointName     string
	useGPU                  bool
)

func getAudioRootCmd() *cobra.Command {
	uploadCmd.PersistentFlags().Int32VarP(&waveformPixelsPerSecond, "pps", "P", 100, "waveform zoom level (pixels per second)")
	uploadCmd.PersistentFlags().StringVar(&modelCheckpointName, "model", "mel_band_roformer_karaoke_aufr33_viperx_sdr_10.1956.ckpt", "model name for audio splitter, see https://raw.githubusercontent.com/nomadkaraoke/python-audio-separator/refs/heads/main/audio_separator/models.json for full list")
	uploadCmd.PersistentFlags().BoolVar(&useGPU, "gpu", false, "use gpu during audio splitter")
	audioCmd.AddCommand(uploadCmd)
	audioCmd.PersistentFlags().StringVarP(&audioPath, "audio", "a", "", "path of audio")
	return audioCmd
}

func processAndUploadAudio(cmd *cobra.Command, args []string) {
	exitIfImage(DoesNotExist)
	s := spinner.New(spinner.CharSets[12], 100*time.Millisecond)
	s.Prefix = "initializing "
	s.Start()
	defer s.Stop()
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)

	var mounts []mount.Mount

	hostExifInfo, err := createTempFile("json")
	check(err)
	defer hostExifInfo.Close()
	defer os.Remove(hostExifInfo.Name())
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: hostExifInfo.Name(), Target: hostExifInfo.Name()})

	hostAudioPath := audioPath
	hostAudioPathSplit := strings.Split(audioPath, ".")
	hostAudioPathExt := hostAudioPathSplit[len(hostAudioPathSplit)-1]
	targetAudioPath, err := createTempFile(hostAudioPathExt)
	check(err)
	defer targetAudioPath.Close()
	defer os.Remove(targetAudioPath.Name())
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: hostAudioPath, Target: targetAudioPath.Name()})

	audioSplitScript := fmt.Sprintf(`from audio_separator.separator import Separator	
separator = Separator(output_format="%s")
separator.load_model("%s")
separator.separate("%s")
	`, strings.ToLower(hostAudioPathExt), modelCheckpointName, targetAudioPath.Name())
	audioSplitScriptFile, err := createTempFile("py")
	check(err)
	defer audioSplitScriptFile.Close()
	defer os.Remove(audioSplitScriptFile.Name())
	audioSplitScriptFile.WriteString(audioSplitScript)
	mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: audioSplitScriptFile.Name(), Target: audioSplitScriptFile.Name()})

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
	scripts := []string{fmt.Sprintf("exiftool %s -json > %s", targetAudioPath.Name(), hostExifInfo.Name())}
	scripts = append(scripts, "$HOME/.local/bin/uv venv")
	separatorPkg := "audio-separator[cpu]"
	if useGPU {
		separatorPkg = "audio-separator[gpu]"
	}
	scripts = append(scripts, fmt.Sprintf("$HOME/.local/bin/uv run --with %s %s", separatorPkg, audioSplitScriptFile.Name()))
	// scripts = append(scripts, fmt.Sprintf("audiowaveform -i %s --pixels-per-second %d > %s", targetAudioPath.Name(), waveformPixelsPerSecond, hostVocalWaveformOutput.Name()))
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
			fmt.Printf("Error scanning logs: %v\n", err)
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
		log.Debugf("exif output: %s", string(fileContents))
	}
	fmt.Println()
}
