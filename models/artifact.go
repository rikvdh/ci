package models

import (
	"github.com/jinzhu/gorm"
)

type Artifact struct {
	gorm.Model

	Job      Job
	JobID    uint
	FilePath string
}
