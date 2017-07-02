package models

import (
	"github.com/jinzhu/gorm"
)

// Artifact repesents a build artifact
//  it belongs to a Job
type Artifact struct {
	gorm.Model

	Job      Job
	JobID    uint
	FilePath string
}
