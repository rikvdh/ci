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
}
