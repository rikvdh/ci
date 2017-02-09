package models

import (
	"github.com/jinzhu/gorm"
)

type Job struct {
	gorm.Model

	Reference string
	Status string
	Container string

	Branch   Branch
	BranchID uint `gorm:"index"`

	Build   Build
	BuildID uint `gorm:"index"`
}
