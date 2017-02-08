package models

import (
	"github.com/jinzhu/gorm"
)

type Branch struct {
	gorm.Model

	Name string
	LastReference string

	Build   Build
	BuildID uint `gorm:"index"`

	Jobs    []Job

	Status string `gorm:"-"`
}

func (b *Branch) FetchLatestStatus() {
	j := Job{}
	dbHandle.Where("build_id = ? AND branch_id = ?", b.BuildID, b.ID).Order("updated_at DESC").First(&j)
	if j.ID == 0 {
		b.Status = "unknown"
	}
	b.Status = j.Status
}
