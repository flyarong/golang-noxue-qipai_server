package dao

import (
	"errors"
	"github.com/golang/glog"
	"qipai/model"
)

var Room roomDao

type roomDao struct {
}

func (roomDao) Get(roomId uint) (room model.Room, err error) {
	if ret := Db().First(&room, roomId); ret.Error != nil || ret.RecordNotFound() {
		err = errors.New("该房间不存在")
		return
	}
	return
}

func (roomDao) IsRoomPlayer(rid, uid uint) bool {
	var n int
	Db().Model(&model.Player{}).Where(&model.Player{Uid: uid, RoomId: rid}).Count(&n)
	return n > 0
}

// 房间中所有坐下的玩家
func (roomDao) PlayersSitDown(roomId uint) (players []model.Player) {
	Db().Where(&model.Player{RoomId: roomId}).Where("desk_id>0").Find(&players)
	return
}
func (roomDao) Exists(roomId uint) bool {
	var n int
	Db().Model(&model.Room{}).Where("id=?", roomId).Count(&n)
	return n > 0
}

// 删除房间信息
func (roomDao) Delete(roomId uint) (err error) {
	res := Db().Unscoped().Where("id=?", roomId).Delete(&model.Room{})
	if res.Error != nil {
		glog.Errorln(res.Error)
		err = errors.New("解散房间出错")
		return
	}
	return
}
