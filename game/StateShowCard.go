package game

import (
	"github.com/golang/glog"
	"github.com/noxue/utils/fsm"
	"qipai/dao"
	"qipai/utils"
	"time"
	"utils/argsUtil"
)

func StateShowCard(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != ShowCardAction {
		return
	}

	var roomId uint
	e := argsUtil.NewArgsUtil(args...).To(&roomId)
	if e != nil {
		glog.Errorln(e)
		return
	}

	gs, err := dao.Game.GetCurrentGames(roomId)
	if err != nil {
		glog.Error(err)
		return
	}
	for _, g := range gs {
		pp := GetPlayer(int(g.PlayerId))
		if pp == nil {
			glog.V(3).Infoln("用户", g.PlayerId, "不在线，无法通知他看牌")
			continue
		}

		e := utils.Msg("").AddData("game", g).Send(BroadcastShowCard, pp.Session)
		if e != nil {
			glog.Error(e)
		}
	}

	g1, e := Games.Get(roomId)
	if e!=nil{
		glog.Error(e)
		return
	}

	nextState = CompareCardState

	go func() {
		time.Sleep(time.Second*1) // 等待一秒后通知看牌
		g1.CompareCard()
	}()

	return
}