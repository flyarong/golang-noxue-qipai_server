package game

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
)

func StateSetScore(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != SetScoreAction {
		return
	}

	var roomId uint
	var uid uint
	var score int

	e := argsUtil.NewArgsUtil(args...).To(&roomId, &uid, &score)
	if e != nil {
		glog.Errorln(e)
		return
	}

	fmt.Println(roomId, "用户：", uid, " 下注，", score, "分")
	n++
	if n == 5 {
		fmt.Println("下注完毕，看牌中...")
		nextState = ShowCardState
	}

	return
}

