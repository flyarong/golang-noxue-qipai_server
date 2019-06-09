package game

import (
	"fmt"
	"github.com/noxue/utils/fsm"
)

func StateShowCard(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != ShowCardAction {
		return
	}

	fmt.Println("看牌完毕!!")
	nextState = GameOverState

	return
}