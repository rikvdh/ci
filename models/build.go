package models

import (
	"github.com/jinzhu/gorm"
)

type Build struct {
	gorm.Model

	Uri string `form:"uri"`
	DefaultBranch string

	Jobs []Job
	Branches []Branch
}
