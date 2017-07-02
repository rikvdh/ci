package models

import (
	"github.com/jinzhu/gorm"
	// blank imports for database types
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
func Init() error {
	var err error
	dbHandle, err = gorm.Open(config.Get().Dbtype, config.Get().DbConnString)
	if err != nil {
		return err
	}

	dbHandle.AutoMigrate(&User{})
	dbHandle.AutoMigrate(&Job{})
	dbHandle.AutoMigrate(&Build{})
	dbHandle.AutoMigrate(&Branch{})
	dbHandle.AutoMigrate(&Artifact{})
	return nil
}
