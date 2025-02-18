package cli

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strafe/internal"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	dockerRootCmd = &cobra.Command{
		Use:   "docker",
		Short: "docker commands",
	}
	imageRootCmd = &cobra.Command{
		Use:   "image",
		Short: "commands relevant to the image",
	}
	imageExistsCmd = &cobra.Command{
		Use:   "exists",
		Short: "check if image exists",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfImage(DoesNotExist)
			color.Green("image exists!")
		},
	}
	buildImageCmd = &cobra.Command{
		Use:   "build [-D --dir] [-F --force] [-Q --quiet]",
		Short: "build the image if it does not exist",
		Long: `builds the utility image if it does not exist already.

source code is required for building the image.
if you are running this command from the root of source code (the one with the Dockerfile in it), then this command will work fine.
if you are running from a different folder, use the --dir / -D flag to provide the source code folder.

command will exit with status code 0 if the image tag already exists (docker.image.name:docker.image.tag from configuration).
use --force / -F flag to override this behaviour and build the image anyways.

logs will be streamed by default just like the normal image building process, use --quiet / -Q to shut me up

this process may take a while.`,
		Run: internal.WrapCommandWithResources(buildImage, internal.ResourceConfig{Resources: []internal.ResourceType{internal.ResourceDocker}}),
	}
	removeImageCmd = &cobra.Command{
		Use:   "remove",
		Short: "remove the image",
		Run:   internal.WrapCommandWithResources(removeImage, internal.ResourceConfig{Resources: []internal.ResourceType{internal.ResourceDocker}}),
	}
	healthImageCmd = &cobra.Command{
		Use:   "health",
		Short: "check health of utilities inside the image",
		Run:   internal.WrapCommandWithResources(healthImage, internal.ResourceConfig{Resources: []internal.ResourceType{internal.ResourceDocker}}),
	}
)

var (
	SourceFolder      string
	ImageBuildContext *bytes.Buffer
	ForceBuildImage   bool
	DisableBuildLogs  bool
)

func getDockerRootCmd() *cobra.Command {
	imageRootCmd.AddCommand(imageExistsCmd)
	buildImageCmd.PersistentFlags().StringVarP(&SourceFolder, "dir", "D", ".", "source code folder")
	buildImageCmd.PersistentFlags().BoolVarP(&ForceBuildImage, "force", "F", false, "build image even if it exists")
	buildImageCmd.PersistentFlags().BoolVarP(&DisableBuildLogs, "quiet", "Q", false, "no log stream")
	imageRootCmd.AddCommand(buildImageCmd)
	imageRootCmd.AddCommand(removeImageCmd)
	imageRootCmd.AddCommand(healthImageCmd)
	dockerRootCmd.AddCommand(imageRootCmd)
	return dockerRootCmd
}

func getImageTag() string {
	return fmt.Sprintf("%s:%s", viper.GetString(internal.DOCKER_IMAGE_NAME), viper.GetString(internal.DOCKER_IMAGE_TAG))
}

func imageExists(docker *client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(internal.TimeoutMS)*time.Millisecond)
	defer cancel()
	_, _, err := docker.ImageInspectWithRaw(ctx, viper.GetString(internal.DOCKER_IMAGE_NAME))
	return err
}

type ImageCheckCondition int

const (
	Exists       ImageCheckCondition = 0
	DoesNotExist ImageCheckCondition = 1
)

func exitIfImage(condition ImageCheckCondition) {
	docker := internal.NewDockerClient()
	defer docker.Close()
	err := imageExists(docker)
	switch condition {
	case Exists:
		if err == nil {
			color.Cyan("image already exists >.<")
			os.Exit(0)
		}
	case DoesNotExist:
		if err != nil {
			color.Red("image does not exist >///<\n%v", err)
			color.Red("try using command %s", color.MagentaString("strafe docker image build"))
			os.Exit(1)
		}
	}
}

type BuildResponse struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

func buildImage(cmd *cobra.Command, args []string) {
	if !ForceBuildImage {
		exitIfImage(Exists)
	}
	s := spinner.New(spinner.CharSets[12], 100*time.Millisecond)
	s.Prefix = "Building image "
	s.Start()
	defer s.Stop()
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	buildCtx, err := createBuildContext(SourceFolder)
	cobra.CheckErr(err)
	response, err := app.Docker.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Tags: []string{getImageTag()},
	})
	cobra.CheckErr(err)
	if !DisableBuildLogs {
		decoder := json.NewDecoder(response.Body)
		for {
			var message BuildResponse
			if err := decoder.Decode(&message); err != nil {
				if err == io.EOF {
					break
				}
				cobra.CheckErr(err)
			}

			if message.Error != "" {
				s.Stop()
				log.Error(message.Error)
				s.Start()
				continue
			}

			if message.Stream != "" {
				cleanMsg := strings.TrimSuffix(message.Stream, "\n")
				if cleanMsg != "" {
					s.Stop()
					fmt.Println(cleanMsg)
					s.Start()
				}
			}
		}
	}
}

func getImageInfo(docker *client.Client) image.Summary {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(internal.TimeoutMS)*time.Millisecond)
	defer cancel()
	filters := filters.NewArgs()
	filters.Add("reference", getImageTag())
	images, err := docker.ImageList(ctx, image.ListOptions{
		Filters: filters,
	})
	cobra.CheckErr(err)
	return images[0]
}

func removeImage(cmd *cobra.Command, args []string) {
	exitIfImage(DoesNotExist)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(color.RedString("this action will remove image %s, are you sure? [y/N] ", getImageTag()))
		s, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(s)) == "n" || strings.TrimSpace(s) == "" {
			color.Cyan("wise choice, goodbye!")
			os.Exit(0)
		}
		if strings.ToLower(strings.TrimSpace(s)) == "y" {
			color.Magenta("removing image...")
			break
		}
	}
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	docker := app.Docker
	resp, err := docker.ImageRemove(ctx, getImageInfo(docker).ID, image.RemoveOptions{Force: true})
	cobra.CheckErr(err)
	color.Green("image %s removed successfully", resp[0].Untagged)
}

// returns container id of whichever argument is not null
// if both is not null, container.CreateResponse has the priority
func getContainerId(resp *container.CreateResponse, id *string) string {
	var cid string
	if resp != nil {
		cid = resp.ID
	} else {
		cid = *id
	}
	return cid
}

// removes container by the specified container ID or the create response from ContainerCreate function.
func removeContainer(ctx context.Context, resp *container.CreateResponse, id *string, docker *client.Client) {
	var cid = getContainerId(resp, id)
	log.Infof("force removing container with the id %s", cid)
	docker.ContainerRemove(ctx, cid, container.RemoveOptions{Force: true})
}

// starts container with the specified container ID or the create response from ContainerCreate function.
func startContainer(ctx context.Context, resp *container.CreateResponse, id *string, docker *client.Client) {
	var cid = getContainerId(resp, id)
	log.Infof("starting container with the id %s", cid)
	if err := docker.ContainerStart(ctx, cid, container.StartOptions{}); err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
}

func healthImage(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	app := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	docker := app.Docker
	exitIfImage(DoesNotExist)
	log.Info("checking if exiftool works...")
	script := `#!/bin/bash
keyfinder-cli
echo "exiftool version: $(exiftool -ver)"
echo "keyfinder-cli (OK if you dont see anything, keyfinder-cli does not have version flag): "
echo "aubio version: $(aubio --version)"
echo "audiowaveform version: $(audiowaveform --version) "
echo "ffprobe version: $(ffprobe -version)"
`
	scriptFile, err := os.CreateTemp(os.TempDir(), "strafe-health-script-*.sh")
	check(err)
	defer scriptFile.Close()
	defer os.Remove(scriptFile.Name())
	_, err = io.WriteString(scriptFile, script)
	check(err)
	resp, err := docker.ContainerCreate(
		ctx,
		&container.Config{
			Image:        getImageTag(),
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
			Cmd:          []string{"/bin/bash", scriptFile.Name()},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: scriptFile.Name(),
					Target: scriptFile.Name(),
				},
			},
		},
		nil,
		nil,
		"")
	check(err)
	defer removeContainer(ctx, &resp, nil, docker)
	startContainer(ctx, &resp, nil, docker)
	statusCh, errCh := docker.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		check(err)
	case <-statusCh:
		out, err := docker.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
		check(err)
		outBytes, err := io.ReadAll(out)
		check(err)
		log.Infof("healthcheck:\n%s", string(outBytes))
	}
	color.Green("image is built and healthy!")
}

func createBuildContext(contextPath string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(contextPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(contextPath, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		header := &tar.Header{
			Name:    relPath,
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tw.Write(data); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func createTempFile(ext string) (*os.File, error) {
	baseDir := os.TempDir()
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("strafe_tmp_%d.%s", timestamp, ext)
	filepath := filepath.Join(baseDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return file, nil
}
