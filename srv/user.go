package srv

import (
	"errors"
	"qipai/dao"
	"qipai/model"
)

var User userSrv

type userSrv struct {
}

func (userSrv) Register(user *model.User) (err error) {

	if len(user.Auths) != 1 {
		err = errors.New("缺少账号信息")
		return
	}
	// 检查授权信息是否存在
	a := model.Auth{UserType: user.Auths[0].UserType, Name: user.Auths[0].Name}
	dao.Db.Where(&a).First(&a)

	if a.Model.ID > 0 {
		err = errors.New(a.Name +" 已被注册，请更换一个账号")
		return
	}

	dao.Db.Create(&user)
	return
}

func (userSrv) Bind(uid uint, auth *model.Auth) (err error) {
	return
}
