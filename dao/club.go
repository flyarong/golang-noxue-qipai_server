package dao

import (
	"errors"
	"github.com/golang/glog"
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

func (clubDao) GetClubUsers(clubId uint)(users []model.ClubUser, err error){
	ret:=Db().Where(&model.ClubUser{ClubId:clubId}).Find(&users)
	if ret!=nil {
		glog.Errorln(ret.Error)
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

// 根据clubId删除茶楼用户
func (clubDao) DelClubUserByClubId(clubId uint) (err error) {
	ret := Db().Unscoped().Where("club_id=?", clubId).Delete(&model.ClubUser{})
	if ret.RowsAffected == 0 {
		err = errors.New("删除茶楼用户失败")
		return
	}
	return
}

func (clubDao) GetUser(clubId, uid uint)(user model.ClubUser, err error){
	ret:=Db().Where(&model.ClubUser{ClubId:clubId,Uid:uid}).First(&user)
	if ret.RecordNotFound(){
		err = errors.New("您不是该茶楼的用户！")
	}
	return
}