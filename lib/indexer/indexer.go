package indexer

import (
	"fmt"
	"strings"
	"time"

	"code.gitea.io/git"
	"github.com/Sirupsen/logrus"
	"github.com/rikvdh/ci/models"
)

const remoteName string = "rem"

type Branch struct {
	Hash string
	Name string
}

func RemoteBranches(repo string) ([]Branch, error) {
	s, err := git.NewCommand("-c", "core.askpass=true", "ls-remote", "-h", repo).RunTimeout(time.Second * 5)
	if err != nil {
		return nil, err
	}

	var branches []Branch

	references := strings.Split(s, "\n")

	for _, ref := range references {
		refsplit := strings.Fields(ref)
		if len(refsplit) == 2 {
			branches = append(branches, Branch{
				Hash: refsplit[0],
				Name: strings.Replace(refsplit[1], "refs/heads/", "", 1),
			})
		}
	}
	if len(branches) > 0 {
		return branches, nil
	}
	return nil, fmt.Errorf("no remote branches found")
}

func ScheduleJob(buildID, branchID uint, ref string) {
	logrus.Infof("Scheduling job for build %d on branch %d", buildID, branchID)
	job := models.Job{
		BuildID:   buildID,
		BranchID:  branchID,
		Status:    models.StatusNew,
		Reference: ref,
	}
	models.Handle().Create(&job)
}

func checkBranch(buildID uint, branch Branch) {
	dbBranch := models.Branch{}
	models.Handle().Where("name = ? AND build_id = ?", branch.Name, buildID).First(&dbBranch)

	if dbBranch.ID > 0 && dbBranch.LastReference != branch.Hash {
		dbBranch.LastReference = branch.Hash
		models.Handle().Save(&dbBranch)
		ScheduleJob(buildID, dbBranch.ID, branch.Hash)
	} else if dbBranch.ID == 0 {
		dbBranch.Name = branch.Name
		dbBranch.BuildID = buildID
		dbBranch.LastReference = branch.Hash
		models.Handle().Create(&dbBranch)
		ScheduleJob(buildID, dbBranch.ID, branch.Hash)
	}
}

func Run() {
	for {
		var builds []models.Build
		models.Handle().Find(&builds)
		for _, build := range builds {
			branches, err := RemoteBranches(build.Uri)
			if err != nil {
				logrus.Warnf("error reading branches from %s: %v", build.Uri, err)
			}
			for _, branch := range branches {
				checkBranch(build.ID, branch)
			}
		}
		time.Sleep(time.Second * 5)
	}
}
