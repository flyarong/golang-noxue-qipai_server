package game

import (
	"github.com/golang/glog"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
	"qipai/utils"
	"time"
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
	g1, e := Games.Get(roomId)
	if e != nil {
		glog.Error(e)
		return
	}

	room, e := dao.Room.Get(roomId)
	if e != nil {
		glog.Error(e)
		return
	}

	// 如果当前已经是最大局数，就设置为结束状态
	if room.Status == enum.GamePlaying && room.Current >= room.Count {
		dao.Db().Model(&room).Update(&model.Room{Status: enum.GameOver})
	}

	go func() {
		time.Sleep(time.Second * 2)

		room, err := dao.Room.Get(roomId)
		if err != nil {
			glog.Error(err)
			return
		}
		if room.Status == enum.GameOver {
			err := Games.GameOver(roomId)
			if err != nil {
				glog.Error(err)
				return
			}

			// 删除房间
			err = dao.Room.Delete(room.ID)
			if err != nil {
				glog.Error(err)
				return
			}

			nextState = GameDeletedState
			SendToAllPlayers(utils.Msg("").AddData("roomId", roomId), BroadcastGameOver, roomId)

			// 删除房间里面的人
			ret := dao.Db().Delete(model.Player{}, "room_id=?", roomId)
			if ret.Error != nil {
				glog.Error(ret.Error)
				return
			}
		} else {
			g1.Start()
		}
	}()

	nextState = ReadyState
	return
}
