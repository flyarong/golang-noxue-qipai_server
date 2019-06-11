package game

import (
	"github.com/golang/glog"
	"github.com/noxue/utils/fsm"
	"qipai/dao"
	"qipai/utils"
	"time"
	"utils/argsUtil"
)

func StateCompareCard(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != CompareCardAction {
		return
	}

	var roomId uint
	e := argsUtil.NewArgsUtil(args...).To(&roomId)
	if e != nil {
		glog.Errorln(e)
		return
	}

	e = cmpCard(roomId)
	if e != nil {
		glog.Errorln(e)
		return
	}

	// 查询计算后的结果，通知大家
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

		e := utils.Msg("").AddData("games", gs).Send(BroadcastCompareCard, pp.Session)
		if e != nil {
			glog.Error(e)
		}
	}

	g1, e := Games.Get(roomId)
	if e != nil {
		glog.Error(e)
		return
	}

	go func() {
		time.Sleep(time.Second * 2) // 游戏结束
		g1.GameOver()
	}()

	nextState = GameOverState
	return
}

func cmpCard(roomId uint) (err error) {
	gs, err := dao.Game.GetCurrentGames(roomId)
	if err != nil {
		glog.Error(err)
		return
	}
	gs = gs
	return
}
