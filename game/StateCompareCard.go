package game

import "github.com/noxue/utils/fsm"

func StateCompareCard(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	if action != CompareCardAction {
		return
	}

	return
}
