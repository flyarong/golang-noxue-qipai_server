package dao

import (
	"errors"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"qipai/model"
)

var User userDao

type userDao struct {
}

func (userDao) Get(uid uint) (user model.User, err error) {
	ret := Db().First(&user, uid)
	if ret.Error != nil {
		err = errors.New("查询用户数据出错")
		glog.Errorln(ret.Error)
		return
	}

	if ret.RecordNotFound() {
		err = errors.New("该用户不存在")
		return
	}
	return
}

// 扣除房卡
func (userDao) TakeCard(uid uint, cards int) (err error) {
	ret:=Db().Model(&model.User{}).Where("id=?", uid).Update("card", gorm.Expr("card-?", cards))
	if ret.Error != nil {
		err = errors.New("扣除房卡出错")
		glog.Errorln(ret.Error)
		return
	}

	if ret.RecordNotFound() {
		err = errors.New("扣除房卡失败")
		return
	}
	return
}

// 是否是特殊用户
func (userDao) IsSpecialUser(uid uint)(ok bool){
	var user model.SpecialUser
	ret := Db().First(&user,uid)
	ok = !ret.RecordNotFound()
	return
}
