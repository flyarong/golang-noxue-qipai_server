package dao

import (
	"errors"
	"github.com/golang/glog"
	"qipai/model"
)

var Game gameDao

type gameDao struct{
}

func (this *gameDao) GetGames(roomId uint, current int) (games []model.Game, err error) {
	if Db().Where(&model.Game{RoomId: roomId, Current: current}).Find(&games).Error != nil {
		err = errors.New("获取游戏信息失败")
		return
	}
	return
}

func (this *gameDao) Players(roomId uint) (players []model.Player) {
	Db().Where(&model.Player{RoomId: roomId}).Find(&players)
	return
}

func (this *gameDao) GetCurrentGames(roomId uint) (game []model.Game, err error) {
	room, e := Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game, err = Game.GetGames(roomId, room.Current)
	return
}

func (gameDao) GetGame(roomId, uid uint, current int) (game model.Game, err error) {

	if Db().Where(&model.Game{RoomId: roomId, PlayerId: uid, Current: current}).First(&game).RecordNotFound() {
		err = errors.New("获取游戏数据失败")
		return
	}
	return
}

func (this *gameDao) GetCurrentGame(roomId, uid uint) (game model.Game, err error) {
	room, e := Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game, err = this.GetGame(roomId, uid, room.Current)
	return
}

func (gameDao) Player(rid, uid uint) (player model.Player, err error) {
	ret:=Db().Where("uid=? and room_id=?", uid, rid).First(&player)
	if ret.Error !=nil {
		err = errors.New("查找数据库失败")
		glog.Errorln(ret.Error)
		return
	}
	if ret.RecordNotFound(){
		err = errors.New("用户未进入当前房间，如果已进入，可以尝试退出房间重新进入")
		return
	}
	return
}