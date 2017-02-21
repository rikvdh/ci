package builder

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

var cli *client.Client
var ctx context.Context

func initCtx() {
	var err error
	ctx = context.Background()
	cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}

}

func startContainer(cfg *config, path string) (string, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cfg.DockerImage,
		Cmd:   []string{"sh", "-c", "/ci/build.sh"},
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

func waitContainer(id string) {
	if _, err := cli.ContainerWait(ctx, id); err != nil {
		panic(err)
	}
}
func stopContainer(id string) {
	out, err := cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	err = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}
