package model

import (
	"github.com/jinzhu/gorm"
	"qipai/enum"
	"qipai/utils"
)



type Auth struct {
	gorm.Model
	UserType enum.UserType `gorm:"type:int;not null"`
	Name     string        `gorm:"size:50"`
	Pass     string        `gorm:"size:50" json:"-"`
	Verified bool
	UserId   int
	User     User `gorm:"foreignkey:UserId" json:"-"`
	Thirdly  bool `json:"-"`
}

type User struct {
	gorm.Model
	Nick    string `gorm:"size:20" json:"nick"`
	Avatar  string `gorm:"size:120" json:"avatar"`
	Mobile  string `gorm:"size:20" json:"mobile"`
	Ip      string `gorm:"size:20" json:"ip"`
	Address string `gorm:"size:50" json:"address"`
	Card    int    `json:"card"`
	Auths   []Auth ` json:"-"`
}

func (this *Auth) BeforeSave() (err error) {
	this.Pass = utils.PassEncode(this.Pass)
	return
}
