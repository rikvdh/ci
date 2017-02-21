package builder

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
)

var ctx context.Context = context.Background()

func getClient() *client.Client {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	return cli
}

func fetchImage(f *os.File, cli *client.Client, cfg *config) error {
	fmt.Fprintf(f, "fetching docker image: %s\n", cfg.DockerImage)
	rc, err := cli.ImagePull(ctx, cfg.DockerImage, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("imagepull failed: %v", err)
	}
	defer rc.Close()

	if _, err := ioutil.ReadAll(rc); err != nil {
		return fmt.Errorf("reading imagepull failed: %v", err)
	}
	fmt.Fprintf(f, "done\n")

	return nil
}

func startContainer(cli *client.Client, cfg *config, path string) (string, error) {
	buildFile := "build-" + randomString(10) + ".sh"

	f, err := os.Create("/tmp/" + buildFile)
	if err != nil {
		return "", err
	}
	f.Chmod(0755)
	f.WriteString(`#!/bin/sh
cd /build
set -xe
echo "$(date) Build started"
sleep 5

`)
	for _, s := range cfg.Script {
		f.WriteString(s + "\n")
	}
	f.WriteString("\necho \"$(date) Build ended\"\n")
	f.Close()

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cfg.DockerImage,
		Cmd:   []string{"sh", "-c", "/ci/" + buildFile},
	}, &container.HostConfig{
		Binds: []string{"/tmp:/ci", path + ":/build"},
	}, nil, "")
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
		return resp.ID, err
	}

	return resp.ID, nil
}

func readContainer(f *os.File, cli *client.Client, id string) {
	out, err := cli.ContainerAttach(ctx, id, types.ContainerAttachOptions{Stdout: true, Stderr: true, Stream: true, Logs: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(f, f, out.Reader)
	out.Close()

	fmt.Fprintf(f, "removing container: %s\n", id)
	err = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}
