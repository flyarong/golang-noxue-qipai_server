package game

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"

)

func StateGameOver(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != GameOverAction {
		return
	}
	var roomId uint
	e := argsUtil.NewArgsUtil(args...).To(&roomId)
	if e != nil {
		glog.Errorln(e)
		return
	}
	fmt.Println("游戏结束！！！")

	nextState = GameDeletedState
	e = Games.GameOver(roomId)
	if e != nil {
		glog.Errorln(e)
		return
	}
	return
}
