package models

import (
	"fmt"
	"github.com/ararog/timeago"
	"github.com/jinzhu/gorm"
	"time"
)

type Job struct {
	gorm.Model

	Reference string
	Status    string
	Container string
	Message   string
	Start     time.Time
	End       time.Time
	Tag       string
	BuildDir  string

	Branch   Branch
	BranchID uint `gorm:"index"`

	Build   Build
	BuildID uint `gorm:"index"`

	StatusTime string `gorm:"-"`

	Artifacts []Artifact
}

func (j *Job) SetStatusTime() {
	if j.Start.IsZero() {
		j.StatusTime = "not started"
	} else if j.End.IsZero() {
		j.StatusTime, _ = timeago.TimeAgoFromNowWithTime(j.Start)
	} else {
		j.StatusTime = j.End.Sub(j.Start).String()
	}
}

func (j *Job) StoreTag(tag string) {
	j.Tag = tag
	Handle().Save(j)
}

func (j *Job) SetStatus(status string, message ...string) error {
	t := time.Now()
	j.Status = status

	// Error, failed and passed are final statusses, add the end-time
	if status == StatusError || status == StatusFailed || status == StatusPassed {
		j.End = t
	}
	if len(message) == 1 {
		j.Message = message[0]
	}
	err := Handle().Save(j).Error
	UpdateBranchStatus(j.BranchID, status, t)
	return err
}

func GetJobById(jobID int, err error) (*Job, error) {
	if err != nil {
		return nil, err
	}

	item := Job{}
	dbHandle.Preload("Branch").Preload("Build").Preload("Artifacts").Where("id = ?", jobID).First(&item)

	if item.ID > 0 {
		item.SetStatusTime()
		return &item, nil
	}
	return nil, fmt.Errorf("error finding job %d", jobID)
}
