package model

import "github.com/jinzhu/gorm"

type UserType int32

const (
	Mobile UserType = 1
	WeChat UserType = 2
)

type Auth struct {
	gorm.Model
	UserType UserType `gorm:"type:int;not null"`
	Name     string   `gorm:"size:50"`
	Pass     string   `gorm:"size:50"`
	Verified bool
	UserId   int
	User     User
}

type User struct {
	gorm.Model
	Name    string `gorm:"size:20"`
	Mobile  string `gorm:"size:20"`
	Ip      string `gorm:"size:20"`
	Address string `gorm:"size:50"`
	Auths   []Auth
}
