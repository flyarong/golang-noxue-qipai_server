package srv

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"qipai/dao"
	"qipai/enum"
	"qipai/event"
	"qipai/model"
	"sync"
	"time"
)

var Room roomSrv

type roomSrv struct {
}

func deleteRoom() {
	type result struct {
		ID  int
		Uid int
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)
			var res []result
			dao.Db.Raw("select id,uid  from rooms  where deleted_at is null and club_id<>1 and status=0 and now()-created_at>1000").Scan(&res)
			if len(res) > 0 {
				var ids []int
				for _, v := range res {
					ids = append(ids, v.ID)
					event.Send(uint(v.Uid), event.RoomDelete, v.ID)
				}
				dao.Db.Unscoped().Where("id in (?)", ids).Delete(model.Room{})
				dao.Db.Where("room_id in (?)", ids).Delete(&model.Player{})
			}
		}
	}()
}

func (this *roomSrv) Create(room *model.Room) (err error) {
	dao.Db.Save(room)
	if room.ID == 0 {
		err = errors.New("房间添加失败，请联系管理员")
		return
	}
	event.Send(room.Uid, event.RoomCreate, room.ID)
	return
}

func (roomSrv) Get(roomId uint) (room model.Room, err error) {
	dao.Db.First(&room, roomId)
	if room.ID == 0 {
		err = errors.New("该房间不存在，或已解散")
		return
	}
	return
}

func (roomSrv) Delete(roomId uint) (err error) {
	// 删除房间信息
	dao.Db.Where("id=? and status=0", roomId).Delete(&model.Room{})

	// 获取玩家
	var ps []model.Player
	dao.Db.Where("room_id=?", roomId).Find(&ps)
	for _, p := range ps {
		// 只通知在线的玩家
		if model.Online.Get(p.Uid) {
			event.Send(p.Uid, "RoomDelete", p.RoomId)
		}
	}

	// 删除玩家
	dao.Db.Where("room_id=?", roomId).Delete(&model.Player{})
	return
}

func (roomSrv) MyRooms(uid uint) (rooms []model.Room) {
	// select r.* from rooms r join  players p on p.room_id=r.id where p.uid=100000;
	dao.Db.Raw("select r.* from rooms r join  players p on p.room_id=r.id where r.`deleted_at` IS NULL and p.uid=?", uid).Scan(&rooms)
	return
}

func (roomSrv) IsRoomPlayer(rid, uid uint) bool {
	var n int
	dao.Db.Model(&model.Player{}).Where(&model.Player{Uid: uid, RoomId: rid}).Count(&n)
	return n > 0
}

func (roomSrv) RoomExists(roomId uint) bool {
	var n int
	dao.Db.Model(&model.Room{}).Where("id=?", roomId).Count(&n)
	return n > 0
}

func (roomSrv) Player(rid, uid uint) (player model.Player, err error) {
	dao.Db.Where("uid=? and room_id=?", uid, rid).First(&player)
	if player.ID == 0 {
		err = errors.New("用户未进入当前房间，如果已进入，可以尝试退出房间重新进入")
		return
	}
	return
}

func (this *roomSrv) SitDown(rid, uid uint) (roomId uint, deskId int, err error) {
	var room model.Room
	room, err = this.Get(rid)
	if err != nil {
		return
	}
	roomId = rid
	// 判断是否已在其他房间坐下
	var p model.Player
	dao.Db.Where("desk_id<>0 and uid=? and room_id<>?", uid, rid).First(&p)
	if p.ID > 0 {
		roomId = p.RoomId
		err = errors.New("您当前正在其他房间")
		return
	}

	// 获取当前玩家座位信息
	var player model.Player
	player, err = this.Player(rid, uid)
	if err != nil {
		return
	}

	// 如果已经坐下，直接返回
	if player.DeskId > 0 {
		deskId = player.DeskId
		return
	}

	// 是否坐满
	var n int
	dao.Db.Model(&model.Player{}).Where("desk_id>0 and room_id=?", rid).Count(&n)
	if n >= room.Players {
		err = errors.New("当前房间已坐满")
		return
	}

	// 删除已经有人的座位
	players := this.Players(rid)
	deskIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[:room.Players-1]
	for _, player := range players {
		for i, v := range deskIds {
			if player.DeskId == v {
				deskIds = append(deskIds[:i], deskIds[i+1:]...)
			}
		}
	}

	// 然后从中没人的座位中 随机选定座位号
	deskId = deskIds[rand.Intn(len(deskIds))]

	// 坐下
	player.DeskId = deskId

	t := time.Now()
	player.JoinedAt = &t
	dao.Db.Save(&player)

	ps := this.PlayersSitDown(rid)
	for _, p := range ps {
		if !model.Online.Get(p.Uid) {
			continue
		}
		event.Send(p.Uid, event.RoomJoin, rid, uid)
	}

	return
}

func (this *roomSrv) Join(rid, uid uint, nick string) (err error) {

	// 检查房间号是否存在
	if !this.RoomExists(rid) {
		err = errors.New("该房间不存在，或已解散")
		return
	}

	if this.IsRoomPlayer(rid, uid) {
		err = errors.New("该用户已经进入房间，不得重复进入")
		return
	}

	ru := model.Player{
		Uid:    uid,
		RoomId: rid,
		Nick:   nick,
	}

	dao.Db.Save(&ru)
	if ru.ID == 0 {
		err = errors.New("加入出错，请联系管理员")
		return
	}

	return
}

/*
退出房间
 */
func (this *roomSrv) Exit(rid, uid uint) (err error) {

	var player model.Player
	player, err = this.Player(rid, uid)
	if err != nil {
		// 房间被解散，也可以成功退出
		if err.Error() == "用户未进入当前房间，如果已进入，可以尝试退出房间重新进入" {
			err = nil
		}
		return
	}

	// 通知其他客户端玩家，我退出了
	ps := this.PlayersSitDown(rid)
	for _, p := range ps {
		event.Send(p.Uid, event.RoomExit, rid, uid)
		log.Println(p.Uid, "\t退出房间")
	}

	player.JoinedAt = nil // 加入时间设置为空
	player.DeskId = 0     // 释放座位号
	dao.Db.Save(&player)

	return
}

func (this *roomSrv) Start(roomId, uid uint) (err error) {

	// 查找房间
	room, e := this.Get(roomId)
	if e != nil {
		err = e
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
		dao.Db.Where("room_id=? and desk_id > 0").Order("joined_at asc").First(&p)
		if p.ID == 0 {
			err = errors.New("该房间还没有人，看到这个错误请联系管理员")
			return
		}
		if p.Uid != uid {
			err = errors.New("您不是第一个进入房间的玩家，无权开始游戏")
			return
		}
	}
	room.Status = enum.GamePlaying

	dao.Db.Save(&room)

	// 通知所有人游戏开始
	for _, p := range this.PlayersSitDown(roomId) {
		event.Send(p.Uid, event.RoomStart, roomId)
	}

	if err = this.GameBegin(roomId); err != nil {
		return
	}

	return
}

// 开始一局新游戏
func (this *roomSrv) GameBegin(roomId uint) (err error) {
	// 发牌
	if err = this.DealCards(roomId); err != nil {
		return
	}

	// 通知所有人游戏开局了
	for _, p := range this.PlayersSitDown(roomId) {
		event.Send(p.Uid, event.GameBegin, roomId)
	}

	room, err := this.Get(roomId)
	if err != nil {
		return
	}

	games, err := this.GetCurrentGames(roomId)
	if err != nil {
		return
	}

	// 加锁，保证线程执行顺序
	var lock sync.Mutex
	for _, game := range games {

		go func(g model.Game) {
			lock.Lock()

			waitTime := time.Second * 10

			// 如果不是第一局，看上一把是否有托管
			if g.Current != 1 {
				tGame, _ := this.GetGame(g.RoomId, g.PlayerId, g.Current-1)
				if tGame.Auto {
					waitTime = time.Second * 4
				}
			}
			lock.Unlock()
			time.Sleep(waitTime)
			lock.Lock()
			defer lock.Unlock()
			// 判断是否还是之前的那一局
			roomNow, err := this.Get(g.RoomId)
			if err != nil {
				return
			}
			// 不是之前那一局，或者游戏结束，就退出
			if room.Current != roomNow.Current || room.Status == enum.GameOver {
				return
			}

			// 如果还没有下注，进入托管
			game, e := this.GetCurrentGame(g.RoomId, g.PlayerId)
			if e != nil {
				err = e
				return
			}
			// 如果用户没有下注，就自动下注
			if game.Times == 0 {
				/*
				自动下注规则，第一局就挂机，那就按照最低倍数下注
				第二局及以后都按照上一局的倍数为准
				*/
				times := 0
				if game.Current == 1 {
					ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
					s := ss[room.Score]
					times = s[0]
				} else {
					tGame, _ := this.GetGame(game.RoomId, game.PlayerId, game.Current-1)
					times = tGame.Times
				}
				game.Times = times
				game.Auto = true // 设置为托管
				dao.Db.Save(&game)
				err = this.SendSetTimes(roomId, g.PlayerId, times)
			}
		}(game)

	}

	return
}

func (this *roomSrv) Players(roomId uint) (players []model.Player) {
	dao.Db.Where(&model.Player{RoomId: roomId}).Find(&players)
	return
}

// 房间中所有坐下的玩家
func (this *roomSrv) PlayersSitDown(roomId uint) (players []model.Player) {
	dao.Db.Where(&model.Player{RoomId: roomId}).Where("desk_id>0").Find(&players)
	return
}

func (this *roomSrv) SendGameOver(roomId uint) {
	for _, p := range this.Players(roomId) {
		event.Send(p.Uid, "GameOver", roomId)
	}
}
func (this *roomSrv) DealCards(roomId uint) (err error) {

	// 获取房间信息
	var room model.Room
	room, err = this.Get(roomId)

	if room.Status == enum.GameOver {
		this.SendGameOver(roomId)
		return
	}

	// 如果当前已经是最大局数，就不发牌了，提示gameover
	if room.Status == enum.GamePlaying && room.Current >= room.Count {
		room.Status = enum.GameOver
		//room.Current --
		t := time.Now()
		room.DeletedAt = &t // 房间解散
		dao.Db.Save(&room)
		this.SendGameOver(roomId)
		return
	}

	// 更新当前局数
	room.Current++
	dao.Db.Save(&room)

	// 获取玩家信息
	players := this.PlayersSitDown(roomId)
	if len(players) < 2 {
		err = errors.New("少于2个玩家，无法开始")
		return
	}

	var cards []int
	for i := 0; i < 54; i++ {
		cards = append(cards, i)
	}
	rand.Seed(time.Now().Unix())
	for _, p := range players {

		var game model.Game
		game.RoomId = roomId
		game.Current = room.Current
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
		dao.Db.Save(&game)
	}

	return
}

// 下注
func (this *roomSrv) SetScore(roomId, uid uint, times int) (err error) {
	room, e := this.Get(roomId)
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

	if s[0] != times && s[1] != times {
		err = errors.New("积分不合法")
		return
	}

	game, e := this.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	if game.Times != 0 {
		err = errors.New("您已经下过注，请勿重复下注")
		return
	}

	game.Times = times
	game.Auto = false // 设置为手动下注
	dao.Db.Save(&game)

	err = this.SendSetTimes(roomId, uid, times)

	return
}

func (this *roomSrv) SendSetTimes(roomId, uid uint, times int) (err error) {
	// 通知所有人有人下注
	allSetTimes := true
	ps := this.Players(roomId)
	for _, p := range ps {
		event.Send(p.Uid, "SetTimes", roomId, uid, times)
	}

	games, e := this.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}

	for _, game := range games {
		// 如果有人没下注，表示不是所有人都下注了
		if game.Times == 0 {
			allSetTimes = false
		}
	}

	// 如果每个人都下注了，通知玩家全下了注
	if allSetTimes {
		for _, p := range ps {
			event.Send(p.Uid, "SetTimesAll", roomId)
		}
		go func() {
			time.Sleep(time.Second * 2)
			this.GameBegin(roomId)
		}()
	}
	return
}

func (this *roomSrv) GetCurrentGame(roomId, uid uint) (game model.Game, err error) {
	room, e := this.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game, err = this.GetGame(roomId, uid, room.Current)
	return
}

func (this *roomSrv) GetGame(roomId, uid uint, current int) (game model.Game, err error) {
	dao.Db.Where(&model.Game{RoomId: roomId, PlayerId: uid, Current: current}).First(&game)
	if game.ID == 0 {
		err = errors.New("获取游戏数据失败")
		return
	}
	return
}

func (this *roomSrv) GetCurrentGames(roomId uint) (game []model.Game, err error) {
	room, e := this.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game, err = this.GetGames(roomId, room.Current)
	return
}

func (this *roomSrv) GetGames(roomId uint, current int) (game []model.Game, err error) {
	dao.Db.Where(&model.Game{RoomId: roomId, Current: current}).Find(&game)
	return
}


