package models

import (
	//"github.com/ararog/timeago"
	"github.com/jinzhu/gorm"
	"time"
)

type Branch struct {
	gorm.Model

	Name          string
	LastReference string

	Build   Build
	BuildID uint `gorm:"index"`

	Jobs []Job

	Status     string
	StatusTime time.Time
}

func (b *Branch) UpdateStatus(s string, t time.Time) {
	b.Status = s
	b.StatusTime = t
	dbHandle.Save(b)
	if b.Build.ID > 0 {
		b.Build.UpdateStatus()
	}
}

func UpdateBranchStatus(branchId uint, s string, t time.Time) {
	b := Branch{}
	dbHandle.Preload("Build").Where("id = ?", branchId).First(&b)
	if b.ID > 0 {
		b.UpdateStatus(s, t)
	}
}
