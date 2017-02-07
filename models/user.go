package models

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model

	Username      string `form:"username"`
	Password      string
	PasswordPlain string `gorm:"-" form:"password"`
}


func (u *User) BeforeSave(scope *gorm.Scope) (err error) {
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

func (u User) IsValid() bool {
	return len(u.Username) > 0 && len(u.PasswordPlain) > 0
}
