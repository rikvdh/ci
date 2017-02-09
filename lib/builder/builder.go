package builder

import (
	"time"
	"math/rand"

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
	models.Handle().Model(&job).Related(&job.Build)
	models.Handle().Model(&job).Related(&job.Branch)
	targetDir := buildDir + "/" + randomString(16)
	err := cloneRepo(job.Build.Uri, job.Branch.Name, job.Reference, targetDir)
	if err != nil {
		panic(err)
	}
	cfg := readCfg(targetDir + "/.ci.yml")
	containerId, err := startContainer(&cfg, targetDir)
	if err != nil {
		panic(err)
	}
	job.Container = containerId
	job.Status = "busy"
	models.Handle().Save(&job)
	waitContainer(containerId)
	stopContainer(containerId)
}

func Run() {
	var newJobs []models.Job
	initCtx()
	for {
		time.Sleep(time.Second * 10)

		models.Handle().Where("status = ?", "new").Find(&newJobs)
		for _, job := range newJobs {
			startJob(job)
		}
	}
}
