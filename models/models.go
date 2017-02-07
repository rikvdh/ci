package models

import (
	"github.com/jinzhu/gorm"
)

var dbHandle *gorm.DB

func Handle() *gorm.DB {
	return dbHandle
}

func Migrate(db *gorm.DB) {
	dbHandle = db

	db.AutoMigrate(&User{})
}
