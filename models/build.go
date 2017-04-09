package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Build struct {
	gorm.Model

	Uri string `form:"uri"`

	Jobs     []Job
	Branches []Branch

	Status     string
	StatusTime time.Time
}

func (b *Build) UpdateStatus() {
	var branches []Branch
	var status string
	var time time.Time

	dbHandle.Where("build_id = ?", b.ID).Find(&branches)
	if len(branches) > 0 {
		status = StatusPassed
		for _, branch := range branches {
			if status == StatusPassed {
				if time.Sub(branch.StatusTime).Seconds() < 0.0 {
					time = branch.StatusTime
				}

				if branch.Status == StatusBusy || branch.Status == StatusNew {
					status = branch.Status
					time = branch.StatusTime
				} else if branch.Status == StatusError || branch.Status == StatusFailed {
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
