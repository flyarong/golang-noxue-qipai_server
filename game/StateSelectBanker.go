package game

import (
	"errors"
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
	"math/rand"
	"qipai/dao"
	"qipai/model"
	"qipai/utils"
	"time"
)

var n = 0

func StateSelectBanker(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	switch action {
	case SetTimesAction:
		var roomId uint
		var uid uint
		var times int
		var auto bool

		res := utils.Msg("")
		defer func() {
			if res == nil {
				res = utils.Msg("").AddData("game", &model.Game{
					PlayerId: uid,
					Times:    times,
					RoomId:   roomId,
					Auto:     auto,
				})
				SendToAllPlayers(res, BroadcastTimes, roomId)
				return
			}
			p := GetPlayer(int(uid))
			if p == nil {
				glog.V(1).Infoln("发送抢庄信息失败，玩家：", uid, "不在线")
				return
			}
			res.Send(BroadcastTimes, p.Session)
		}()
		e := argsUtil.NewArgsUtil(args...).To(&roomId, &uid, &times, &auto)
		if e != nil {
			glog.Errorln(e)
			return
		}

		room, e := dao.Room.Get(roomId)
		if e != nil {
			res = utils.Msg(e.Error()).Code(-1)
			return
		}
		// 抢庄
		if dao.Db().Model(&model.Game{}).Where("player_id=? and times = -1", uid).Update(map[string]interface{}{"times": times, "auto": auto}).Error != nil {
			res = utils.Msg("更新下注信息失败").Code(-1)
			return
		}

		// 判断是否都抢庄
		games, err := dao.Game.GetGames(roomId, room.Current)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}

		all := true // 是否全部都抢庄了
		for _, v := range games {
			// 如果还有没抢庄的，直接返回，通知所有人该用户的抢庄倍数
			if v.Times == -1 {
				glog.V(3).Infoln(roomId, "房间：", uid, " 抢庄，", times, "倍。是否自动抢庄：", auto)
				all = false
				break
			}
		}

		// 全部抢庄完毕，选择庄家，并通知所有人
		if all {
			// 进入闲家下注状态
			nextState = SetScoreState

			bankerUid, err := selectBanker(roomId)
			if err != nil {
				res = utils.Msg(err.Error()).Code(-1)
				return
			}
			go func() {
				time.Sleep(time.Second) // 等待一秒后通知谁是庄家
				res = utils.Msg("").AddData("game", &model.Game{
					PlayerId: bankerUid,
					Banker:   true,
				})
				SendToAllPlayers(res, BroadcastBanker, roomId)
			}()
		}

		res = nil // 抢庄倍数设置成功，res设置为nil，将在defer函数中通知所有人
		return
	}
	return
}

// 选择庄家
func selectBanker(roomId uint) (uid uint, err error) {
	games, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}
	if len(games) == 0 {
		err = errors.New("当前房间没有玩家")
		return
	}

	eq := true // 记录是否全部相等
	var game = games[0]
	// 选择下注最大的
	for _, g := range games {
		if g.Times != game.Times {
			eq = false
			if g.Times > game.Times {
				game = g
			}
		}
	}
	// 如果都一样大，就随机选一个
	if eq {
		rand.Seed(time.Now().Unix())
		game = games[rand.Intn(len(games))]
	}

	uid = game.PlayerId
	// 更新
	res := dao.Db().Model(&game).Update("banker", true)
	if res.Error != nil {
		err = errors.New("选定庄家出错")
		return
	}
	if res.RowsAffected == 0 {
		err = errors.New("更新庄家信息出错")
		return
	}
	return
}
