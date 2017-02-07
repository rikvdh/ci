package models

import (
	"github.com/jinzhu/gorm"
)

type Build struct {
	gorm.Model

	Uri string
	DefaultBranch string

	Jobs []Job
	Branches []Branch
}
