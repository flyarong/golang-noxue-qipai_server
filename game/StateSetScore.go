package game

import (
	"errors"
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
	"qipai/dao"
	"qipai/game/card"
	"qipai/model"
	"qipai/utils"
	"time"
)

func StateSetScore(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != SetScoreAction {
		return
	}

	var roomId uint
	var uid uint
	var score int
	var auto bool

	res := utils.Msg("")
	alreadySet := false // 是否已经下注
	defer func() {
		if alreadySet { // 已经下注就直接退出，不用再次通知
			return
		}
		if res == nil {
			res = utils.Msg("").AddData("game", &model.Game{
				PlayerId: uid,
				Score:    score,
				RoomId:   roomId,
				Auto:     auto,
			})
			SendToAllPlayers(res, BroadcastScore, roomId)
			return
		}
		p := GetPlayer(int(uid))
		if p == nil {
			glog.V(1).Infoln("玩家：", uid, "不在线，发送下注信息失败")
			return
		}
		res.Send(BroadcastScore, p.Session)
	}()

	e := argsUtil.NewArgsUtil(args...).To(&roomId, &uid, &score, &auto)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		glog.Errorln(e)
		return
	}

	room, e := dao.Room.Get(roomId)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		glog.Errorln(e)
		return
	}

	ret := dao.Db().Model(&model.Game{}).Where("room_id=? and player_id=? and current=? and banker=0", roomId, uid, room.Current).Update(map[string]interface{}{"score": score, "auto": auto})
	if ret.Error != nil {
		res = utils.Msg(ret.Error.Error()).Code(-1)
		return
	}

	if ret.RowsAffected == 0 {
		alreadySet = true
	}

	games, err := dao.Game.GetGames(roomId, room.Current)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	all := true // 是否全部都抢庄了
	for _, v := range games {
		if v.Banker {
			continue
		}
		// 如果还有没下注的，直接返回，通知所有人该用户的抢庄倍数
		if v.Score == 0 {
			all = false
			break
		}
	}

	if all {
		nextState = ShowCardState

		// 计算牌型
		e:=calcCard(roomId)
		if e!=nil{
			res = utils.Msg(e.Error()).Code(-1)
			return
		}
		g1, e := Games.Get(roomId)
		if e!=nil{
			res = utils.Msg(e.Error()).Code(-1)
			return
		}

		go func() {
			time.Sleep(time.Millisecond*500) // 等待0.5秒后通知看牌
			g1.ShowCard()
		}()
	}

	glog.V(3).Infoln(roomId, "房间：", uid, " 下注，", score, "分。是否自动下注：", auto)

	res = nil
	return
}


func calcCard(roomId uint)(err error){
	games, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}
	if len(games) == 0 {
		err = errors.New("当前房间没有玩家")
		return
	}

	for _, g := range games {
		paixing, cardStr, e := card.GetPaixing(g.Cards)
		if e != nil {
			err = e
			return
		}
		if dao.Db().Model(&g).Update(map[string]interface{}{"card_type": paixing, "cards": cardStr}).Error != nil {
			err = errors.New("更新牌型失败")
			return
		}
	}
	return
}