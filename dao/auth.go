package dao

import (
	"errors"
	"qipai/enum"
	"qipai/model"
)

type authDao struct {
}

var Auth authDao

func (authDao) Get(userType enum.UserType, name string) (auth model.Auth, err error) {
	ret := Db().Where("user_type=? and name=?", userType, name).First(&auth)
	if ret.RecordNotFound() {
		err = errors.New("账号或密码错误")
		return
	}
	return
}
