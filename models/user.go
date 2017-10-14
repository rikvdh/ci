package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model

	Username      string
	Password      string
	PasswordPlain string `gorm:"-"`
	APIKey        string
}

func ValidAPIKey(key string) bool {
	if key == "" {
		return false
	}
	item := User{}
	dbHandle.Where("api_key = ?", key).First(&item)
	if item.ID > 0 {
		return true
	}
	return false
}

func (u *User) BeforeSave(scope *gorm.Scope) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordPlain), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	scope.SetColumn("password", hashedPassword)
	return nil
}

func (u *User) ValidPassword() bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(u.PasswordPlain))
	if err == nil {
		return true
	}
	return false
}

// IsValid validates the user
func (u User) IsValid() bool {
	return len(u.Username) > 0 && len(u.PasswordPlain) > 0
}
