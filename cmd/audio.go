package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
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
	waveformCmd = &cobra.Command{
		Use:   "waveform",
		Short: "calculate waveform peaks of given audio path",
		Long: `calculates waveform peaks of given audio path
.mp3, .wav, .flac, .ogg, .oga, .opus, .dat and .json formats are supported`,
		Run: calculateWaveform,
	}
	waveformPixelsPerSecond int32
)

func getAudioRootCmd() *cobra.Command {
	waveformCmd.PersistentFlags().Int32VarP(&waveformPixelsPerSecond, "pps", "P", 100, "zoom level (pixels per second)")
	audioCmd.AddCommand(waveformCmd)
	audioCmd.PersistentFlags().StringVarP(&audioPath, "audio", "a", "", "path of audio")
	return audioCmd
}

func calculateWaveform(cmd *cobra.Command, args []string) {
	exitIfImage(DoesNotExist)
	s := spinner.New(spinner.CharSets[12], 100*time.Millisecond)
	s.Prefix = "initializing "
	s.Start()
	defer s.Stop()
	docker := newDockerClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(TimeoutMS))
	defer cancel()
	hostOutput, err := createTempFile("json")
	check(err)
	containerOutput, err := createTempFile("json")
	check(err)
	log.Debugf("binding output file %s to %s", hostOutput.Name(), containerOutput.Name())
	defer hostOutput.Close()
	defer os.Remove(hostOutput.Name())
	check(err)
	hostAudioPath := audioPath
	hostAudioPathSplit := strings.Split(audioPath, ".")
	targetAudioPath, err := createTempFile(hostAudioPathSplit[len(hostAudioPathSplit)-1])
	check(err)
	log.Infof("creating container")
	s.Prefix = "creating container "
	resp, err := docker.ContainerCreate(
		ctx,
		&container.Config{
			Image:        getImageTag(),
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			Cmd:          []string{"audiowaveform", "-i", targetAudioPath.Name(), "--pixels-per-second", fmt.Sprintf("%d", waveformPixelsPerSecond), "--output-format", "json", ">", containerOutput.Name()},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: hostOutput.Name(),
					Target: containerOutput.Name(),
				},
				{
					Type:   mount.TypeBind,
					Source: hostAudioPath,
					Target: targetAudioPath.Name(),
				},
			},
		},
		nil,
		nil,
		"")
	check(err)
	defer removeContainer(ctx, &resp, nil, docker)
	log.Infof("starting container")
	s.Prefix = "starting container"
	startContainer(ctx, &resp, nil, docker)
	s.Prefix = "waiting for container to finish"
	statusCh, errCh := docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		check(err)
	case <-statusCh:
		s.Prefix = "audiowaveform executed successfully "
		out, err := docker.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
		check(err)
		outBytes, err := io.ReadAll(out)
		check(err)
		log.Infof("stdout:\n%s", string(outBytes))

		file, err := os.OpenFile(hostOutput.Name(), os.O_RDONLY, os.ModeTemporary)
		check(err)
		fileContents, err := io.ReadAll(file)
		check(err)
		log.Debugf("output: %s", string(fileContents))
	}
	fmt.Println()
	color.Cyan("successfully generated waveform!")
}
