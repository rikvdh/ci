package indexer

import (
	"fmt"
	"time"

	"github.com/rikvdh/ci/models"
	"srcd.works/go-git.v4"
	"srcd.works/go-git.v4/config"
	"srcd.works/go-git.v4/plumbing"
	"srcd.works/go-git.v4/storage/memory"
)

type Branch struct {
	Hash string
	Name string
}

func RemoteBranches(repo string) ([]Branch, error) {
	// Create a new repository
	r, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return nil, fmt.Errorf("git init error: %v", err)
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "r",
		URL:  repo,
	})
	if err != nil {
		return nil, fmt.Errorf("create remote error: %v", err)
	}

	rem, err := r.Remote("r")
	if err != nil {
		return nil, fmt.Errorf("remote err: %v", err)
	}

	err = rem.Fetch(&git.FetchOptions{})
	if err != nil {
		return nil, err
	}

	refs, err := r.References()
	if err != nil {
		return nil, err
	}

	var branches []Branch

	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference && !ref.IsTag() {
			branches = append(branches, Branch{
				Hash: ref.Hash().String(),
				Name: ref.Name().Short(),
			})
		}
		return nil
	})

	return branches, nil
}

func ScheduleJob(buildID, branchID uint, ref string) {
	fmt.Println("Scheduling job for build", buildID, "on branch", branchID)
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
				fmt.Printf("error reading branches from %s: %v\n", build.Uri, err)
			}
			for _, branch := range branches {
				checkBranch(build.ID, branch)
			}
		}
		time.Sleep(time.Second * 5)
	}
}
