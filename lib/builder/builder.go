package builder

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rikvdh/ci/models"
)

const buildDir string = "/home/rik/ci-build"

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
	targetDir := buildDir + "/" + randomString(16)
	err := cloneRepo(job.Build.Uri, job.Branch.Name, job.Reference, targetDir)
	if err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("cloning repository failed: %v", err))
		return
	}

	cfg := readCfg(targetDir + "/.ci.yml")
	containerID, err := startContainer(&cfg, targetDir)
	if err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("starting container failed: %v", err))
		return
	}
	job.Container = containerID
	job.SetStatus(models.StatusBusy)

	waitContainer(containerID)
	stopContainer(containerID)
}

func Run() {
	var newJobs []models.Job
	initCtx()
	for {
		time.Sleep(time.Second * 10)

		models.Handle().Preload("Branch").Preload("Build").Where("status = ?", models.StatusNew).Find(&newJobs)
		for _, job := range newJobs {
			startJob(job)
		}
	}
}
