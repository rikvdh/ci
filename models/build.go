package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Build struct {
	gorm.Model

	URI        string `form:"uri"`
	Personal   bool   `form:"personal"`
	UserID     uint   `gorm:"index" form:"userID"`
	User       User
	Jobs       []Job
	Branches   []Branch
	Status     string
	StatusTime time.Time
}

func (b *Build) IsValid() error {
	if len(b.URI) == 0 {
		return errors.New("missing repository URI")
	}
	build := Build{}
	if dbHandle.Where("uri = ?", b.URI).First(&build); build.ID > 0 {
		return errors.New("build with the provided URI already exists")
	}
	return nil
}

func BuildWithBranches(buildID int, userID uint) (*Build, error) {
	item := Build{}
	err := dbHandle.Preload("Branches", func(db *gorm.DB) *gorm.DB {
		return db.Order("branches.Enabled DESC, branches.status_time DESC")
	}).Where("((personal = 1 AND user_id = ?) OR personal = 0) AND id = ?", userID, buildID).First(&item).Error
	if err != nil {
		return nil, err
	}
	item.URI = cleanReponame(item.URI)
	for k := range item.Branches {
		item.Branches[k].LastReference = item.Branches[k].LastReference[:7]
	}
	return &item, nil
}

func BuildList(userID uint, err error) ([]Build, error) {
	if err != nil {
		return nil, err
	}
	var builds []Build
	err = dbHandle.Where("(personal = 1 AND user_id = ?) OR personal = 0", userID).Order("updated_at DESC").Find(&builds).Error
	if err != nil {
		return nil, err
	}
	for k := range builds {
		builds[k].URI = cleanReponame(builds[k].URI)
	}
	return builds, nil
}

func (b *Build) UpdateStatus() {
	var branches []Branch
	var status string
	var time time.Time

	dbHandle.Where("build_id = ?", b.ID).Find(&branches)
	if len(branches) > 0 {
		status = StatusPassed
		for _, branch := range branches {
			if branch.Status == StatusBusy || branch.Status == StatusNew {
				status = branch.Status
				time = branch.StatusTime
			}
			if status == StatusPassed {
				if time.Sub(branch.StatusTime).Seconds() < 0.0 {
					time = branch.StatusTime
				}

				if branch.Status == StatusError || branch.Status == StatusFailed {
					status = branch.Status
					time = branch.StatusTime
				}
			}
		}
	} else {
		status = StatusUnknown
		time = b.CreatedAt
	}

	b.Status = status
	b.StatusTime = time
	dbHandle.Save(b)
}

func cleanReponame(remote string) string {
	remote = strings.Replace(remote, ".git", "", -1)
	if strings.Contains(remote, ":") && strings.Contains(remote, "@") {
		rem := remote[strings.Index(remote, "@")+1:]
		return strings.Replace(rem, ":", "/", 1)
	}
	u, err := url.Parse(remote)
	if err != nil {
		return remote
	}
	return u.Hostname() + u.Path
}
