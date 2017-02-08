package builder

import (
	"io"
	"os"
	"time"
	"math/rand"

	"github.com/rikvdh/ci/models"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

const buildDir string = "./build"

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func startJob(job models.Job) {
	models.Handle().Model(&job).Related(&job.Build)
	models.Handle().Model(&job).Related(&job.Branch)
	targetDir := buildDir + "/" + randomString(16)
	err := cloneRepo(job.Build.Uri, job.Branch.Name, job.Reference, targetDir)
	if err != nil {
		panic(err)
	}
	cfg := readCfg(targetDir + "/.ci.yml")
}

func Run() {
	var newJobs []models.Job
	for {
		models.Handle().Where("status = ?", "new").Find(&newJobs)
		time.Sleep(time.Second * 10)
		for _, job := range newJobs {
			startJob(job)
		}
	}
}

func BoemBats() {
	img := "golang"
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	_, err = cli.ImagePull(ctx, img+":latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: img,
		Cmd:   []string{"sh", "-c", "/ci/build.sh"},
		//		Cmd:   []string{"ls", "-al", "/ci", "/build"},
	}, &container.HostConfig{
		Binds: []string{"/tmp:/ci", "/root/build:/build"},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	if _, err = cli.ContainerWait(ctx, resp.ID); err != nil {
		panic(err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	err = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		panic(err)
	}
}
