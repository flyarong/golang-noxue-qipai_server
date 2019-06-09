package dao

import (
	"errors"
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
