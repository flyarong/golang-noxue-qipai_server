package game

import (
	"errors"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/noxue/utils/fsm"
	"qipai/dao"
	"qipai/game/card"
	"qipai/model"
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
	var banker model.Game  // 庄家
	var games []model.Game // 闲家
	for _, v := range gs {
		if v.Banker {
			banker = v
			continue
		}
		games = append(games, v)
	}

	// 庄家和每个闲家比较
	for _, v := range games {
		result, e := card.CmpCards(banker.Cards, v.Cards)
		if e != nil {
			err = e
			return
		}
		win := -1           // 记录庄家是否胜利，最终记录积分正负
		winnerCardType := 0 // 记录赢家牌型
		if result >= 0 {
			winnerCardType = banker.CardType
			win = 1
		} else if result < 0 {
			winnerCardType = v.CardType
		}

		bankerTimes := banker.Times
		//庄家没抢庄，防止*0,
		if bankerTimes == 0 {
			bankerTimes = 1
		}
		// 牌型倍数 * 闲家下注倍数 * 庄家抢庄倍数
		totalScore := getTimes(winnerCardType) * v.Score * bankerTimes

		// 更新闲家积分
		ret := dao.Db().Model(&v).Update("total_score", totalScore*win*-1)
		if ret.Error != nil {
			glog.Error(ret.Error)
			err = errors.New("更新闲家积分出错")
			return
		}

		// 更新闲家总积分到玩家数据表
		ret = dao.Db().Model(&model.Player{}).Where("uid=? and room_id=?", v.PlayerId, v.RoomId).Update("total_score", gorm.Expr("total_score + ?", totalScore*win*-1))
		if ret.Error != nil {
			glog.Error(ret.Error)
			err = errors.New("更新闲家总积分出错")
			return
		}

		// 更新庄家积分
		ret = dao.Db().Model(&banker).Update("total_score", gorm.Expr("total_score + ?", totalScore*win))
		if ret.Error != nil {
			glog.Error(ret.Error)
			err = errors.New("更新庄总家积分出错")
			return
		}

		// 更新庄家总积分到玩家数据表
		ret = dao.Db().Model(&model.Player{}).Where("uid=? and room_id=?", banker.PlayerId, banker.RoomId).Update("total_score", gorm.Expr("total_score + ?", totalScore*win))
		if ret.Error != nil {
			glog.Error(ret.Error)
			err = errors.New("更新庄家总积分出错")
			return
		}
	}
	return
}

func getTimes(cardType int) (times int) {
	switch cardType {
	case card.DouniuType_Niu7, card.DouniuType_Niu8, card.DouniuType_Niu9:
		times = 2
		break
	case card.DouniuType_Niuniu:
		times = 3
		break
	case card.DouniuType_Wuhua:
		times = 5
		break
	case card.DouniuType_Zhadan:
		times = 8
		break
	case card.DouniuType_Wuxiao:
		times = 10
		break
	default:
		times = 1
	}
	return
}
