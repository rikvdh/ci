package builder

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rikvdh/ci/lib/buildcfg"
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

func fetchImage(f *os.File, cli *client.Client, cfg *buildcfg.Config) error {
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

func startContainer(cli *client.Client, cfg *buildcfg.Config, path string) (string, error) {
	buildFile := "build-" + randomString(10) + ".sh"

	f, err := os.OpenFile("/tmp/"+buildFile, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return "", err
	}
	cfg.GetScript(f)
	f.Close()

	buildDir := "/build"
	if len(cfg.GoImportPath) > 0 {
		buildDir = "/build/src/" + cfg.GoImportPath
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cfg.DockerImage,
		Cmd:   []string{"sh", "-c", "/ci/" + buildFile},
		Env:   []string{"TRAVIS_OS_NAME=linux"},
	}, &container.HostConfig{
		Binds: []string{"/tmp:/ci", path + ":" + buildDir},
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

func readContainer(f *os.File, cli *client.Client, id string) (int, error) {
	if len(id) == 0 {
		return 0, fmt.Errorf("container '%s' is invalid", id)
	}

	out, err := cli.ContainerAttach(ctx, id, types.ContainerAttachOptions{Stdout: true, Stderr: true, Stream: true, Logs: true})
	if err != nil {
		return 0, fmt.Errorf("container attach: %v", err)
	}

	stdcopy.StdCopy(f, f, out.Reader)
	out.Close()

	cnt, err := cli.ContainerInspect(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("inspect error: %v", err)
	}
	exitcode := cnt.State.ExitCode
	fmt.Fprintf(f, "exitcode: %d\n", exitcode)

	fmt.Fprintf(f, "removing container: %s\n", id)
	err = cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
	if err != nil {
		return exitcode, fmt.Errorf("container remove: %v", err)
	}
	return exitcode, nil
}
