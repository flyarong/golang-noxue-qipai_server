package game

import (
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
	"qipai/config"
	"qipai/dao"
	"qipai/model"
	"qipai/utils"
	"strings"
	"time"
)

func StateReady(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	switch action {
	case StartAction:
		var roomId uint
		e := argsUtil.NewArgsUtil(args...).To(&roomId)
		if e != nil {
			glog.Errorln(e)
			return
		}

		err := DealCards(roomId)
		if err != nil {
			glog.Error(err)
			SendToAllPlayers(utils.Msg(err.Error()).Code(-1), ResGameStart, roomId)
			return
		}

		room, err := dao.Room.Get(roomId)
		if err != nil {
			glog.Error(err)
			SendToAllPlayers(utils.Msg(err.Error()).Code(-1), ResGameStart, roomId)
			return
		}

		ps := dao.Room.PlayersSitDown(roomId)
		for _, p := range ps {

			// 如果超过一定时间还没有下注，就定时下注
			go func(player model.Player) {
				g, e := Games.Get(player.RoomId)
				if e != nil {
					glog.Error(e)
					return
				}

				auto, _ := g.AutoPlayers[player.Uid]
				var t time.Duration = 10 // 正常10秒钟
				if config.Config.Debug {
					t = 1;
				}
				if auto {
					t = 2 // 托管模式 2秒钟自动抢庄
				}
				time.Sleep(time.Second * t)
				g.SetTimes(player.Uid, 0, true) // 默认不抢庄
			}(p)

			pp := GetPlayer(int(p.Uid))
			if pp == nil {
				glog.Warningln("房间：", roomId, "中  玩家：", p.Uid, "不在线")
				continue
			}

			func() {
				res := utils.Msg("")
				defer func() {
					SendMessage(res, ResGameStart, pp.Session)
				}()

				// 获取指定玩家，当局游戏信息
				var gameInfo model.Game
				ret := dao.Db().Where(&model.Game{RoomId: roomId, PlayerId: uint(pp.Uid), Current: room.Current}).First(&gameInfo)
				if ret.Error != nil {
					res = utils.Msg(ret.Error.Error()).Code(-1)
				}
				cards := ""
				for i, v := range strings.Split(gameInfo.Cards, "|") {
					if i == 4 {
						break
					}
					cards += v + "|"
				}
				gameInfo.Cards = cards[:len(cards)-1]
				res = utils.Msg("").AddData("game", gameInfo)
			}()
		}
		nextState = SelectBankerState
	}
	return
}
