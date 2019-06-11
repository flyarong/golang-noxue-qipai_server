package srv

import (
	"errors"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
)

var Game gameSrv

type gameSrv struct {
}

func (this *gameSrv) Start(roomId, uid uint) (err error) {

	// 查找房间
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	players := dao.Room.PlayersSitDown(roomId)
	if len(players) < 2 {
		err = errors.New("少于2个玩家，无法开始")
		return
	}

	// 根据游戏开始方式，判断uid指定的用户是否可以开始游戏
	if room.StartType == enum.StartBoss { // 老板才能开始
		// 判断是否是boss
		if room.Uid != uid {
			err = errors.New("该房间只有房主可以开始游戏")
			return
		}
	} else if room.StartType == enum.StartFirst { // 首位开始
		// 获取首位玩家
		var p model.Player
		res := dao.Db().Where("room_id=? and desk_id > 0", roomId).Order("joined_at asc").First(&p)
		if res.Error != nil || res.RecordNotFound() {
			err = errors.New("该房间还没有人，看到这个错误请联系管理员")
			return
		}
		if p.Uid != uid {
			err = errors.New("您不是第一个进入房间的玩家，无权开始游戏")
			return
		}
	}
	room.Status = enum.GamePlaying

	dao.Db().Model(&room).Update("status", enum.GamePlaying)

	g, e := game.Games.NewGame(roomId)
	if e != nil {
		err = e
		return
	}
	g.Start()

	return
}

// 设置抢庄倍数
func (this *gameSrv) SetTimes(roomId, uid uint, times int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	if times < 0 || times > 4 {
		err = errors.New("抢庄倍数错误，只能是0-4倍")
		return
	}

	g, e := dao.Game.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经抢庄，直接返回
	if g.Times >= 0 {
		return
	}

	game1, e := game.Games.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game1.SetTimes(uid, times, false)

	return
}

// 下注
func (this *gameSrv) SetScore(roomId, uid uint, score int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
	s := ss[room.Score]

	if s[0] != score && s[1] != score {
		err = errors.New("积分不合法")
		return
	}

	g, e := dao.Game.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经下注，直接返回
	if g.Score != 0 {
		return
	}

	g1, e := game.Games.Get(roomId)
	if e != nil {
		err = e
		return
	}
	g1.SetScore(uid, score, false)
	return
}
