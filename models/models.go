package models

import (
	"github.com/rikvdh/ci/lib/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var dbHandle *gorm.DB

func Handle() *gorm.DB {
	return dbHandle
}

func Init() {
	dbHandle, err := gorm.Open(config.Get().Dbtype, config.Get().DbConnString)
	if err != nil {
		panic("failed to connect database")
	}

	dbHandle.AutoMigrate(&User{})
	dbHandle.AutoMigrate(&Job{})
	dbHandle.AutoMigrate(&Build{})
	dbHandle.AutoMigrate(&Branch{})
}
