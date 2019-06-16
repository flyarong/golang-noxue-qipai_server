package game

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
	"qipai/utils"
	"time"
	"zero"
)

// 发送信息给指定房间的所有坐下的玩家
func SendToAllPlayers(msg *utils.Message, msgId int32, roomId uint) {
	ps := dao.Room.PlayersSitDown(roomId)
	for _, p := range ps {
		pp := GetPlayer(p.Uid)
		if pp == nil {
			continue
		}
		err := msg.Send(msgId, pp.Session)
		if err != nil {
			glog.Error(err)
		}
	}
	return
}

// 发送消息
func SendMessage(msg *utils.Message, msgId int32, s *zero.Session) (err error) {
	message := zero.NewMessage(msgId, msg.ToBytes())
	if s == nil {
		glog.Warningln("session为nil指针，发送的消息编号为是：", msgId)
		return
	}
	err = s.GetConn().SendMessage(message)
	return
}

// 发牌
func DealCards(roomId uint) (err error) {

	// 获取房间信息
	var room model.Room
	room, err = dao.Room.Get(roomId)

	if room.Status == enum.GameOver {
		err = errors.New("游戏已经结束")
		return
	}


	// 更新当前局数
	dao.Db().Model(&room).Update(&model.Room{Current: room.Current + 1})

	// 获取玩家信息
	players := dao.Room.PlayersSitDown(roomId)
	if len(players) < 2 {
		err = errors.New("少于2个玩家，无法开始")
		return
	}

	var cards []int
	for i := 0; i < 52; i++ {
		cards = append(cards, i)
	}
	rand.Seed(time.Now().Unix())
	for _, p := range players {

		var game model.Game
		game.RoomId = roomId
		game.Current = room.Current
		game.DeskId = p.DeskId
		game.PlayerId = p.Uid

		for j := 0; j < 5; j++ {
			n := 0
			if room.King == enum.KingNone {
				n = rand.Intn(len(cards) - 2)
			} else {
				n = rand.Intn(len(cards))
			}
			game.Cards += fmt.Sprintf("|%v", cards[n])
			cards = append(cards[:n], cards[n+1:]...)
		}
		game.Cards = game.Cards[1:]
		dao.Db().Save(&game)
	}

	return
}

// 根据房间号和uid获取玩家信息
func GetRoomPlayer(roomId, uid uint) (player model.Player, err error) {
	res := dao.Db().Where(&model.Player{RoomId: roomId, Uid: uid}).First(&player)
	if res.Error != nil {
		glog.Errorln(res.Error)
		err = errors.New("从数据库查询玩家信息发生错误")
		return
	}
	if res.RecordNotFound() {
		err = errors.New(fmt.Sprintf("房间[%v]不存在[%v]用户", roomId, uid))
		glog.Error(err)
	}
	return
}
