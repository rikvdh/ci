package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/rikvdh/ci/lib/config"
)

var dbHandle *gorm.DB

// Handle returns the database interface handle, it is a singleton
func Handle() *gorm.DB {
	return dbHandle
}

// Init initializes the database and auto-migrates the database tables
func Init() {
	var err error
	dbHandle, err = gorm.Open(config.Get().Dbtype, config.Get().DbConnString)
	if err != nil {
		panic("failed to connect database")
	}

	dbHandle.AutoMigrate(&User{})
	dbHandle.AutoMigrate(&Job{})
	dbHandle.AutoMigrate(&Build{})
	dbHandle.AutoMigrate(&Branch{})
	dbHandle.AutoMigrate(&Artifact{})
}
