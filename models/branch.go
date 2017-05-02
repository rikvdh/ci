package models

import (
	//"github.com/ararog/timeago"
	"fmt"
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

func UpdateBranchStatus(branchID uint, s string, t time.Time) {
	b := Branch{}
	dbHandle.Preload("Build").Where("id = ?", branchID).First(&b)
	if b.ID > 0 {
		b.UpdateStatus(s, t)
	}
}

func GetBranchByID(branchID int, err error) (*Branch, error) {
	if err != nil {
		return nil, err
	}

	item := Branch{}
	dbHandle.Preload("Jobs", func(db *gorm.DB) *gorm.DB {
		return db.Order("jobs.id DESC")
	}).Preload("Build").Where("id = ?", branchID).First(&item)

	if item.ID > 0 {
		return &item, nil
	}
	return nil, fmt.Errorf("error finding branch %d", branchID)
}
