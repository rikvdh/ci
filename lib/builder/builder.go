package builder

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/rikvdh/ci/lib/buildcfg"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/models"
)

var runningJobs uint
var buildDir string
var buildEvent chan uint

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// GetLog retrieves a log for a job
func GetLog(job *models.Job) string {
	d, err := ioutil.ReadFile(buildDir + "/" + strconv.Itoa(int(job.ID)) + ".log")
	if err != nil {
		return ""
	}
	return string(d)
}

// Returns boolean true when the job is started
func startJob(f *os.File, job models.Job) {
	fmt.Fprintf(f, "starting build job %d\n", job.ID)
	job.BuildDir = buildDir + "/" + randomString(16)
	job.Start = time.Now()
	job.SetStatus(models.StatusBusy)

	tag, err := cloneRepo(f, job.Build.Uri, job.Branch.Name, job.Reference, job.BuildDir)
	if err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("cloning repository failed: %v", err))
		return
	}
	job.StoreTag(tag)

	fmt.Fprintf(f, "reading configuration\n")
	cfg := buildcfg.Read(job.BuildDir, job.Build.Uri)

	cli := getClient()
	if err := fetchImage(f, cli, &cfg); err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("fetch image failed: %v", err))
		return
	}

	fmt.Fprintf(f, "starting container...\n")
	containerID, err := startContainer(cli, &cfg, job.BuildDir)
	if err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("starting container failed: %v", err))
		return
	}
	fmt.Fprintf(f, "container started, ID: %s\n", containerID)

	job.Container = containerID
	job.SetStatus(models.StatusBusy)

	go waitForJob(f, cli, &job, &cfg)
	runningJobs++
	buildEvent <- runningJobs
}

func waitForJob(f *os.File, cli *client.Client, job *models.Job, cfg *buildcfg.Config) {
	logrus.Infof("Wait for job %d", job.ID)
	models.Handle().First(&job, job.ID)
	code, err := readContainer(f, cli, job.Container)
	if err != nil {
		job.SetStatus(models.StatusError, err.Error())
	} else if code != 0 {
		job.SetStatus(models.StatusFailed, fmt.Sprintf("build failed with code: %d", code))
	} else {
		handleArtifacts(f, job, cfg)
	}
	runningJobs--
	buildEvent <- runningJobs
	cli.Close()
}

func GetEventChannel() *chan uint {
	return &buildEvent
}

func retakeRunningJobs() {
	var jobs []models.Job
	models.Handle().Preload("Branch").Preload("Build").Where("status = ?", models.StatusBusy).Find(&jobs)
	for _, job := range jobs {
		logrus.Infof("Retake job %d", job.ID)
		f, err := os.OpenFile(buildDir+"/"+strconv.Itoa(int(job.ID))+".log", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			job.SetStatus(models.StatusError, fmt.Sprintf("reopening logfile failed: %v", err))
			continue
		}
		defer f.Close()

		cli := getClient()
		go waitForJob(f, cli, &job, nil)
		runningJobs++
	}
}

// Run is the build-runner, it starts containers and runs up to 5 parallel builds
func Run() {
	buildEvent = make(chan uint)
	buildDir, _ = filepath.Abs(config.Get().BuildDir)
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		os.Mkdir(buildDir, 0755)
	}

	retakeRunningJobs()

	for {
		if runningJobs < config.Get().ConcurrentBuilds {
			var job models.Job

			models.Handle().Preload("Branch").Preload("Build").Where("status = ?", models.StatusNew).First(&job)
			if job.ID > 0 {
				f, err := os.OpenFile(buildDir+"/"+strconv.Itoa(int(job.ID))+".log", os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					job.SetStatus(models.StatusError, fmt.Sprintf("creating logfile failed: %v", err))
					continue
				}
				defer f.Close()

				startJob(f, job)
			} else {
				time.Sleep(time.Second * 5)
			}
		} else {
			logrus.Infof("Job ratelimiter: %d/%d", runningJobs, config.Get().ConcurrentBuilds)
			time.Sleep(time.Second * 5)
		}
	}
}
