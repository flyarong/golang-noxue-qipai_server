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
	Avatar  string `gorm:"size:255" json:"avatar"`
	Mobile  string `gorm:"size:20" json:"mobile"`
	Sex     int    `json:"sex"` // 普通用户性别，1为男性，2为女性
	Ip      string `gorm:"size:22" json:"ip"`
	Address string `gorm:"size:100" json:"address"`
	Card    int    `json:"card"`
	Auths   []Auth ` json:"-"`
}

func (this *Auth) BeforeSave() (err error) {
	this.Pass = utils.PassEncode(this.Pass)
	return
}
