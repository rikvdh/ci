package models

import (
	"github.com/ararog/timeago"
	"github.com/jinzhu/gorm"
)

type Build struct {
	gorm.Model

	Uri string `form:"uri"`

	Jobs     []Job
	Branches []Branch

	Status     string `gorm:"-"`
	StatusTime string `gorm:"-"`
}

func (b *Build) FetchLatestStatus() {
	j := Job{}
	dbHandle.Where("build_id = ?", b.ID).Order("updated_at DESC").First(&j)
	if j.ID == 0 {
		b.Status = "unknown"
	}
	b.Status = j.Status

	if j.UpdatedAt.IsZero() {
		b.StatusTime = "n/a"
	} else {
		b.StatusTime, _ = timeago.TimeAgoFromNowWithTime(j.UpdatedAt)
	}
}
