package models

import (
	"github.com/jinzhu/gorm"
)

type Job struct {
	gorm.Model

	Reference string
	Status    string
	Container string
	Message   string

	Branch   Branch
	BranchID uint `gorm:"index"`

	Build   Build
	BuildID uint `gorm:"index"`
}

const (
	StatusNew    = "new"
	StatusBusy   = "busy"
	StatusFailed = "failed"
	StatusPassed = "passed"
	StatusError  = "error"
)

func (j *Job) SetStatus(status string, message ...string) error {
	j.Status = status
	if len(message) == 1 {
		j.Message = message[0]
	}
	return Handle().Save(j).Error
}
