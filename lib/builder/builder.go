package builder

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/client"
	"github.com/rikvdh/ci/lib/buildcfg"
	"github.com/rikvdh/ci/lib/config"
	"github.com/rikvdh/ci/models"
	"path/filepath"
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
func startJob(f *os.File, cli *client.Client, job models.Job) bool {
	targetDir := buildDir + "/" + randomString(16)

	fmt.Fprintf(f, "starting build job %d\n", job.ID)
	job.Start = time.Now()

	if err := cloneRepo(f, job.Build.Uri, job.Branch.Name, job.Reference, targetDir); err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("cloning repository failed: %v", err))
		return false
	}

	fmt.Fprintf(f, "reading configuration\n")
	cfg := buildcfg.Read(targetDir, job.Build.Uri)

	if err := fetchImage(f, cli, &cfg); err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("fetch image failed: %v", err))
	}

	fmt.Fprintf(f, "starting container...\n")
	containerID, err := startContainer(cli, &cfg, targetDir)
	if err != nil {
		job.SetStatus(models.StatusError, fmt.Sprintf("starting container failed: %v", err))
		return false
	}
	fmt.Fprintf(f, "container started, ID: %s\n", containerID)

	job.Container = containerID
	job.SetStatus(models.StatusBusy)
	return true
}

func waitForJob(f *os.File, cli *client.Client, job models.Job) {
	models.Handle().First(&job, job.ID)
	code, err := readContainer(f, cli, job.Container)
	if err != nil {
		job.SetStatus(models.StatusError, err.Error())
	} else if code != 0 {
		job.SetStatus(models.StatusFailed, fmt.Sprintf("build failed with code: %d", code))
	} else {
		job.SetStatus(models.StatusPassed)
	}
	runningJobs--
	buildEvent <- runningJobs
	cli.Close()
}

func GetEventChannel() chan uint {
	return buildEvent
}

func retakeRunningJobs() {
	var jobs []models.Job
	models.Handle().Preload("Branch").Preload("Build").Where("status = ?", models.StatusBusy).Find(&jobs)
	for _, job := range jobs {
		f, err := os.OpenFile(buildDir+"/"+strconv.Itoa(int(job.ID))+".log", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			job.SetStatus(models.StatusError, fmt.Sprintf("creating logfile failed: %v", err))
			continue
		}
		defer f.Close()
		cli := getClient()
		go waitForJob(f, cli, job)
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
				cli := getClient()

				f, err := os.OpenFile(buildDir+"/"+strconv.Itoa(int(job.ID))+".log", os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					job.SetStatus(models.StatusError, fmt.Sprintf("creating logfile failed: %v", err))
					continue
				}
				defer f.Close()

				started := startJob(f, cli, job)
				if started {
					go waitForJob(f, cli, job)
					runningJobs++
					buildEvent <- runningJobs
				}
			} else {
				time.Sleep(time.Second * 5)
			}
		} else {
			fmt.Printf("Job ratelimiter: %d/%d\n", runningJobs, config.Get().ConcurrentBuilds)
			time.Sleep(time.Second * 5)
		}
	}
}
