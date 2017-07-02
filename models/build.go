package models

import (
	"errors"
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
