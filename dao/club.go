package dao

import (
	"errors"
	"qipai/model"
)

var Club clubDao

type clubDao struct {
}

func (clubDao) Get(clubId uint) (club model.Club, err error) {
	ret := Db().First(&club, clubId)
	if ret.RecordNotFound() {
		err = errors.New("该茶楼不存在")
		return
	}
	return
}

func (clubDao) Del(clubId uint) (err error) {
	ret := Db().Delete(&model.Club{}, clubId)
	if ret.RowsAffected == 0 {
		err = errors.New("删除茶楼失败")
		return
	}
	return
}
