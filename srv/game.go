package srv

import (
	"errors"
	"fmt"
	"qipai/dao"
	"qipai/enum"
	"qipai/event"
	"qipai/game"
	"qipai/model"
	"qipai/srv/card"
	"time"
)

var Game gameSrv

type gameSrv struct {
}

func (this *gameSrv) Start(roomId, uid uint) (err error) {

	// 查找房间
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	players := dao.Room.PlayersSitDown(roomId)
	if len(players) < 2 {
		err = errors.New("少于2个玩家，无法开始")
		return
	}

	// 根据游戏开始方式，判断uid指定的用户是否可以开始游戏
	if room.StartType == enum.StartBoss { // 老板才能开始
		// 判断是否是boss
		if room.Uid != uid {
			err = errors.New("该房间只有房主可以开始游戏")
			return
		}
	} else if room.StartType == enum.StartFirst { // 首位开始
		// 获取首位玩家
		var p model.Player
		res := dao.Db().Where("room_id=? and desk_id > 0", roomId).Order("joined_at asc").First(&p)
		if res.Error != nil || res.RecordNotFound() {
			err = errors.New("该房间还没有人，看到这个错误请联系管理员")
			return
		}
		if p.Uid != uid {
			err = errors.New("您不是第一个进入房间的玩家，无权开始游戏")
			return
		}
	}
	room.Status = enum.GamePlaying

	dao.Db().Model(&room).Update("status", enum.GamePlaying)

	g,e:=game.Games.NewGame(roomId)
	if e!=nil {
		err = e
		return
	}
	g.Start()

	return
}

// 设置抢庄倍数
func (this *gameSrv) SetTimes(roomId, uid uint, times int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	if times < 0 || times > 4 {
		err = errors.New("抢庄倍数错误，只能是0-4倍")
		return
	}

	g, e := this.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经抢庄，直接返回
	if g.Times >= 0 {
		return
	}

	game.GetPlayer(int(uid))
	return
}

// 下注
func (this *gameSrv) SetScore(roomId, uid uint, score int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
	s := ss[room.Score]

	if s[0] != score && s[1] != score {
		err = errors.New("积分不合法")
		return
	}

	game, e := this.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经下注，直接返回
	if game.Score != 0 {
		return
	}

	if dao.Db().Model(&game).Update(map[string]interface{}{"score": score, "auto": false}).Error != nil {
		err = errors.New("更新下注信息失败")
		return
	}

	err = this.SendSetScore(roomId, uid, score)

	return
}
//
//func (this *gameSrv) SendSetTimes(roomId, uid uint, times int) (err error) {
//
//	// 通知所有人有人下注
//	ps := dao.Game.Players(roomId)
//	for _, p := range ps {
//		event.Send(p.Uid, "SetTimes", roomId, uid, times)
//	}
//
//	games, e := this.GetCurrentGames(roomId)
//	if e != nil {
//		err = e
//		return
//	}
//
//	for _, game := range games {
//		// 如果有人没抢庄
//		if game.Times < 0 {
//			return
//		}
//	}
//
//	// 抢庄完毕，通知玩家谁是庄家
//	bankerUid, e := this.selectBanker(roomId)
//	if e != nil {
//		err = e
//		return
//	}
//	for _, p := range ps {
//		event.Send(p.Uid, "SetBanker", roomId, bankerUid)
//	}
//
//	room, err := dao.Room.Get(roomId)
//	if err != nil {
//		return
//	}
//
//	// 加锁，保证线程执行顺序
//	var lock sync.Mutex
//	for _, game := range games {
//
//		go func(g model.Game) {
//			lock.Lock()
//
//			waitTime := time.Second * 5
//
//			// 如果不是第一局，看上一把是否有托管
//			if g.Current != 1 {
//				tGame, _ := this.GetGame(g.RoomId, g.PlayerId, g.Current-1)
//				if tGame.Auto {
//					waitTime = time.Second * 2
//				}
//			}
//			lock.Unlock()
//			time.Sleep(waitTime)
//			lock.Lock()
//			defer lock.Unlock()
//			// 判断是否还是之前的那一局
//			roomNow, err := dao.Room.Get(g.RoomId)
//			if err != nil {
//				return
//			}
//			// 不是之前那一局，或者游戏结束，就退出
//			if room.Current != roomNow.Current || room.Status == enum.GameOver {
//				return
//			}
//
//			game, e := this.GetCurrentGame(g.RoomId, g.PlayerId)
//			if e != nil {
//				err = e
//				return
//			}
//
//			// 庄家不用下注
//			if game.Banker {
//				return
//			}
//
//			// 如果闲家已经下注，返回
//			if game.Score > 0 {
//				return
//			}
//
//			// 自动下注规则，第一局就挂机，那就按照最低倍数下注
//			// 第二局及以后都按照上一局的倍数为准
//			ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
//			s := ss[room.Score]
//			score := s[0]
//			if game.Current > 1 {
//				tGame, _ := this.GetGame(game.RoomId, game.PlayerId, game.Current-1)
//				if tGame.Score > 0 {
//					score = tGame.Score
//				}
//			}
//
//			// 不抢庄并设置为托管
//			if dao.Db().Model(&game).Update(map[string]interface{}{"score": score, "auto": true}).Error != nil {
//				return
//			}
//
//			// 通知玩家，下注信息
//			err = this.SendSetScore(roomId, g.PlayerId, score)
//
//		}(game)
//
//	}
//
//	return
//}

func (this *gameSrv) SendSetScore(roomId, uid uint, score int) (err error) {
	// 通知所有人有人下注
	ps := dao.Game.Players(roomId)
	for _, p := range ps {
		event.Send(p.Uid, "SetScore", roomId, uid, score)
	}

	games, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}

	for _, game := range games {
		// 有闲家没下注，返回
		if !game.Banker && game.Score == 0 {
			return
		}
	}

	// 计算牌型
	err = this.checkPaixin(roomId)
	if err != nil {
		return
	}

	// 获取牌型字符串
	gs, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}

	pxs := ""
	for _, v := range gs {
		pxs += fmt.Sprintf("%v,%v,%v;", v.PlayerId, v.CardType, v.Cards)
	}

	for _, p := range ps {
		event.Send(p.Uid, "SetScoreAll", roomId)
		event.Send(p.Uid, "CardTypes", roomId, pxs[:len(pxs)-1])
	}
	go func() {
		time.Sleep(time.Second * 5)
		room, e := dao.Room.Get(roomId)
		if e != nil {
			return
		}

		if room.Current == room.Count {
			go func(rid uint) {
				time.Sleep(time.Second * 5)
				dao.Db().Delete(&model.Room{}, rid)
				// 把玩家从房间删除
				dao.Db().Where("room_id=?", rid).Delete(model.Player{})
				this.SendGameOver(rid)
			}(roomId)
		}
		if room.Status == enum.GamePlaying {
			//this.GameBegin(roomId)
		}
	}()

	return
}

// 计算指定房间，当前牌局每个人的牌型并更新到数据库
func (this *gameSrv) checkPaixin(roomId uint) (err error) {
	games, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}
	if len(games) == 0 {
		err = errors.New("当前房间没有玩家")
		return
	}

	for _, g := range games {
		paixing, cardStr, e := card.GetPaixing(g.Cards)
		if e != nil {
			err = e
			return
		}
		if dao.Db().Model(&g).Update(map[string]interface{}{"card_type": paixing, "cards": cardStr}).Error != nil {
			err = errors.New("更新牌型失败")
			return
		}
	}
	return
}


func (this *gameSrv) GetCurrentGame(roomId, uid uint) (game model.Game, err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game, err = this.GetGame(roomId, uid, room.Current)
	return
}

func (this *gameSrv) GetGame(roomId, uid uint, current int) (game model.Game, err error) {

	if dao.Db().Where(&model.Game{RoomId: roomId, PlayerId: uid, Current: current}).First(&game).RecordNotFound() {
		err = errors.New("获取游戏数据失败")
		return
	}
	return
}


func (this *gameSrv) SendGameOver(roomId uint) {
	for _, p := range dao.Game.Players(roomId) {
		event.Send(p.Uid, "GameOver", roomId)
	}
}


func (roomSrv) Player(rid, uid uint) (player model.Player, err error) {
	dao.Db().Where("uid=? and room_id=?", uid, rid).First(&player)
	if player.ID == 0 {
		err = errors.New("用户未进入当前房间，如果已进入，可以尝试退出房间重新进入")
		return
	}
	return
}