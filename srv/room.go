package srv

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"qipai/dao"
	"qipai/enum"
	"qipai/event"
	"qipai/game"
	"qipai/model"
	"qipai/srv/card"
	"qipai/utils"
	"sync"
	"time"
)

var Room roomSrv

type roomSrv struct {
}

func deleteAllInvalidRooms() {
	//type result struct {
	//	ID  int
	//	Uid int
	//}
	//
	//var res []result
	//dao.Db().Raw("select id,uid  from rooms  where deleted_at is null and club_id<>1 and status=0 and now()-created_at>=1000").Scan(&res)
	//if len(res) > 0 {
	//	var ids []int
	//	for _, v := range res {
	//		ids = append(ids, v.ID)
	//		if p := game.GetPlayer(v.Uid); p != nil {
	//			// 通知在线的相关用户
	//			utils.Msg("房间超过10分钟未开始，自动解散").AddData("id", v.ID).Send(game.ResDeleteRoom, p.Session)
	//		}
	//	}
	//	dao.Db().Unscoped().Where("id in (?)", ids).Delete(model.Room{})
	//	dao.Db().Where("room_id in (?)", ids).Delete(&model.Player{})
	//}

	var rooms []model.Room
	res := dao.Db().Find(&rooms)
	if res.Error != nil {
		glog.Errorln(res.Error)
		return
	}
	for _, room := range rooms {
		// 超过10分钟，游戏没开始的房间，并且不是俱乐部房间，自动解散
		if isRoomExpired(&room) {
			dao.Db().Unscoped().Delete(&room)
			sendRoomDelete(room.ID)
			continue // 继续下一个
		}

		// 主要是处理中途关闭程序，启动后部分房间未到10分钟但未开始的情况
		if time.Now().Sub(room.CreatedAt) < time.Minute*10 && room.ClubId == 0 && room.Status == enum.GameReady {
			go func(roomId uint) {
				// 等待剩余的时间
				time.Sleep(time.Minute*10 - time.Now().Sub(room.CreatedAt))
				deleteExpiredRoom(roomId)
			}(room.ID)
		}

	}
}

// 检查是否超时，超时返回true，表示可以删除了。
func isRoomExpired(room *model.Room) bool{
	// 超过10分钟，游戏没开始的房间，并且不是俱乐部房间，自动解散
	return (time.Now().Sub(room.CreatedAt) >= time.Minute*10 && room.ClubId == 0 && room.Status == enum.GameReady )
}

// 删除过期的房间，并通知客户端
func deleteExpiredRoom(roomId uint) (err error){
	var room model.Room
	res:=dao.Db().Find(&room,roomId)
	if res.Error != nil {
		glog.Errorln(res.Error)
		return
	}
	if res.RecordNotFound(){
		err = errors.New("没有找到房间")
		return
	}
	if isRoomExpired(&room) {
		dao.Db().Unscoped().Delete(&room)
		sendRoomDelete(room.ID)
	}
	return
}

func sendRoomDelete(roomId uint) (err error) {
	var ps []model.Player
	res := dao.Db().Where(&model.Player{RoomId: roomId}).Find(&ps)
	if res.Error != nil {
		err = errors.New(fmt.Sprint("查询房间对应的玩家失败，房间号：%d", roomId))
		return
	}
	for _, v := range ps {
		if p := game.GetPlayer(int(v.Uid)); p != nil {
			// 通知在线的相关用户
			utils.Msg("房间超过10分钟未开始或已经结束，自动解散").AddData("id", roomId).Send(game.ResDeleteRoom, p.Session)
		}
		dao.Db().Unscoped().Delete(&v)
	}
	return
}

func (this *roomSrv) Create(room *model.Room) (err error) {
	res := dao.Db().Save(room)
	if res.Error != nil || res.RowsAffected == 0 {
		err = errors.New("房间添加失败，请联系管理员")
		return
	}

	go func() {
		time.Sleep(time.Minute*10 + time.Second)
		deleteAllInvalidRooms()
	}()

	return
}

func (roomSrv) Get(roomId uint) (room model.Room, err error) {

	if dao.Db().First(&room, roomId).RecordNotFound() {
		err = errors.New("该房间不存在，或游戏已结束")
		return
	}
	return
}

func (roomSrv) Delete(roomId uint) (err error) {
	// 删除房间信息
	dao.Db().Where("id=? and status=0", roomId).Delete(&model.Room{})

	// 获取玩家
	var ps []model.Player
	dao.Db().Where("room_id=?", roomId).Find(&ps)
	for _, p := range ps {
		// 只通知在线的玩家
		if model.Online.Get(p.Uid) {
			event.Send(p.Uid, "RoomDelete", p.RoomId)
		}
	}

	// 删除玩家
	dao.Db().Where("room_id=?", roomId).Delete(&model.Player{})
	return
}

func (roomSrv) MyRooms(uid uint) (rooms []model.Room) {
	// select r.* from rooms r join  players p on p.room_id=r.id where p.uid=100000;
	dao.Db().Raw("select r.* from rooms r join  players p on p.room_id=r.id where r.`deleted_at` IS NULL and p.uid=?", uid).Scan(&rooms)
	return
}

func (roomSrv) IsRoomPlayer(rid, uid uint) bool {
	var n int
	dao.Db().Model(&model.Player{}).Where(&model.Player{Uid: uid, RoomId: rid}).Count(&n)
	return n > 0
}

func (roomSrv) RoomExists(roomId uint) bool {
	var n int
	dao.Db().Model(&model.Room{}).Where("id=?", roomId).Count(&n)
	return n > 0
}

func (roomSrv) Player(rid, uid uint) (player model.Player, err error) {
	dao.Db().Where("uid=? and room_id=?", uid, rid).First(&player)
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

	if !dao.Db().Where("desk_id<>0 and uid=? and room_id<>?", uid, rid).First(&p).RecordNotFound() {
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
		this.sendSitDown(roomId, uid)
		return
	}

	// 是否坐满
	var n int
	dao.Db().Model(&model.Player{}).Where("desk_id>0 and room_id=?", rid).Count(&n)
	if n >= room.Players {
		err = errors.New("当前房间已坐满")
		return
	}

	// 删除已经有人的座位
	players := this.Players(rid)
	deskIds := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[:room.Players]
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
	dao.Db().Save(&player)

	this.sendSitDown(roomId, uid)

	return
}

func (this *roomSrv) sendSitDown(rid, uid uint) {
	ps := this.PlayersSitDown(rid)
	for _, p := range ps {
		if !model.Online.Get(p.Uid) {
			continue
		}
		event.Send(p.Uid, event.PlayerSitDown, rid, uid)
	}
}

func (this *roomSrv) Join(rid, uid uint, nick string) (err error) {

	// 检查房间号是否存在
	if !this.RoomExists(rid) {
		err = errors.New("该房间不存在，或已解散")
		return
	}

	if this.IsRoomPlayer(rid, uid) {
		return
	}

	ru := model.Player{
		Uid:    uid,
		RoomId: rid,
		Nick:   nick,
	}

	dao.Db().Save(&ru)
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

	player, e := this.Player(rid, uid)
	if e != nil {
		// 房间被解散，也可以成功退出
		if e.Error() == "用户未进入当前房间，如果已进入，可以尝试退出房间重新进入" {
			return
		}
		err = e
		return
	}

	// 游戏开始后无法退出
	room, e := this.Get(rid)
	if e != nil {
		err = e
		return
	}
	if room.Status == enum.GamePlaying {
		err = errors.New("游戏中，无法退出")
		return
	}

	if dao.Db().Model(model.Player{}).Where("id=?", player.ID).Update(map[string]interface{}{"desk_id": 0, "joined_at": nil}).Error != nil {
		err = errors.New("更新退出房间数据失败")
	}

	this.SendExit(rid, uid)
	return
}

func (this *roomSrv) SendExit(rid, uid uint) {
	// 通知其他客户端玩家，我退出了
	ps := this.PlayersSitDown(rid)
	for _, p := range ps {
		pp:=game.GetPlayer(int(p.Uid))
		if pp == nil {
			continue
		}
		utils.Msg("").AddData("roomId",rid).AddData("uid",uid).Send(game.ResLeaveRoom, pp.Session)
	}
}

func (this *roomSrv) Start(roomId, uid uint) (err error) {

	// 查找房间
	room, e := this.Get(roomId)
	if e != nil {
		err = e
		return
	}

	players := this.PlayersSitDown(roomId)
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

	dao.Db().Save(&room)

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

			// 已经抢庄，直接返回
			if game.Times != -1 {
				return
			}

			// 不抢庄并设置为托管
			if dao.Db().Model(&game).Update(map[string]interface{}{"times": 0, "auto": true}).Error != nil {
				return
			}

			// 通知玩家，抢庄信息
			err = this.SendSetTimes(roomId, g.PlayerId, 0)

		}(game)

	}

	return
}

func (this *roomSrv) Players(roomId uint) (players []model.Player) {
	dao.Db().Where(&model.Player{RoomId: roomId}).Find(&players)
	return
}

// 房间中所有坐下的玩家
func (this *roomSrv) PlayersSitDown(roomId uint) (players []model.Player) {
	dao.Db().Where(&model.Player{RoomId: roomId}).Where("desk_id>0").Find(&players)
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
		err = errors.New("游戏已经结束")
		return
	}

	// 如果当前已经是最大局数，就不发牌了，提示gameover
	if room.Status == enum.GamePlaying && room.Current >= room.Count {
		room.Status = enum.GameOver
		dao.Db().Save(&room)
		err = errors.New("游戏已经结束")
		return
	}

	// 更新当前局数
	room.Current++
	dao.Db().Save(&room)

	// 获取玩家信息
	players := this.PlayersSitDown(roomId)
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

// 设置抢庄倍数
func (this *roomSrv) SetTimes(roomId, uid uint, times int) (err error) {
	room, e := this.Get(roomId)
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

	game, e := this.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经抢庄，直接返回
	if game.Times >= 0 {
		return
	}

	// 手动抢庄，并设置为非托管
	if dao.Db().Model(&game).Update(map[string]interface{}{"times": times, "auto": false}).Error != nil {
		err = errors.New("更新下注信息失败")
		return
	}

	err = this.SendSetTimes(roomId, uid, times)
	return
}

// 下注
func (this *roomSrv) SetScore(roomId, uid uint, score int) (err error) {
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

func (this *roomSrv) SendSetTimes(roomId, uid uint, times int) (err error) {

	// 通知所有人有人下注
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
		// 如果有人没抢庄
		if game.Times < 0 {
			return
		}
	}

	// 抢庄完毕，通知玩家谁是庄家
	bankerUid, e := this.selectBanker(roomId)
	if e != nil {
		err = e
		return
	}
	for _, p := range ps {
		event.Send(p.Uid, "SetBanker", roomId, bankerUid)
	}

	room, err := this.Get(roomId)
	if err != nil {
		return
	}

	// 加锁，保证线程执行顺序
	var lock sync.Mutex
	for _, game := range games {

		go func(g model.Game) {
			lock.Lock()

			waitTime := time.Second * 5

			// 如果不是第一局，看上一把是否有托管
			if g.Current != 1 {
				tGame, _ := this.GetGame(g.RoomId, g.PlayerId, g.Current-1)
				if tGame.Auto {
					waitTime = time.Second * 2
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

			game, e := this.GetCurrentGame(g.RoomId, g.PlayerId)
			if e != nil {
				err = e
				return
			}

			// 庄家不用下注
			if game.Banker {
				return
			}

			// 如果闲家已经下注，返回
			if game.Score > 0 {
				return
			}

			// 自动下注规则，第一局就挂机，那就按照最低倍数下注
			// 第二局及以后都按照上一局的倍数为准
			ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
			s := ss[room.Score]
			score := s[0]
			if game.Current > 1 {
				tGame, _ := this.GetGame(game.RoomId, game.PlayerId, game.Current-1)
				if tGame.Score > 0 {
					score = tGame.Score
				}
			}

			// 不抢庄并设置为托管
			if dao.Db().Model(&game).Update(map[string]interface{}{"score": score, "auto": true}).Error != nil {
				return
			}

			// 通知玩家，下注信息
			err = this.SendSetScore(roomId, g.PlayerId, score)

		}(game)

	}

	return
}

func (this *roomSrv) SendSetScore(roomId, uid uint, score int) (err error) {
	// 通知所有人有人下注
	ps := this.Players(roomId)
	for _, p := range ps {
		event.Send(p.Uid, "SetScore", roomId, uid, score)
	}

	games, e := this.GetCurrentGames(roomId)
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
	gs, e := this.GetCurrentGames(roomId)
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
		room, e := this.Get(roomId)
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
			this.GameBegin(roomId)
		}
	}()

	return
}

// 计算指定房间，当前牌局每个人的牌型并更新到数据库
func (this *roomSrv) checkPaixin(roomId uint) (err error) {
	games, e := this.GetCurrentGames(roomId)
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

// 选择庄家
func (this *roomSrv) selectBanker(roomId uint) (uid uint, err error) {
	games, e := this.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}
	if len(games) == 0 {
		err = errors.New("当前房间没有玩家")
		return
	}

	eq := true // 记录是否全部相等
	var game = games[0]
	// 选择下注最大的
	for _, g := range games {
		if g.Times != game.Times {
			eq = false
			if g.Times > game.Times {
				game = g
			}
		}
	}
	// 如果都一样大，就随机选一个
	if eq {
		rand.Seed(time.Now().Unix())
		game = games[rand.Intn(len(games))]
	}

	uid = game.PlayerId
	// 更新
	res := dao.Db().Model(&game).Update("banker", true)
	if res.Error != nil {
		err = errors.New("选定庄家出错")
		return
	}
	if res.RowsAffected == 0 {
		err = errors.New("更新庄家信息出错")
		return
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

	if dao.Db().Where(&model.Game{RoomId: roomId, PlayerId: uid, Current: current}).First(&game).RecordNotFound() {
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
	if dao.Db().Where(&model.Game{RoomId: roomId, Current: current}).Find(&game).Error != nil {
		err = errors.New("获取游戏信息失败")
		return
	}
	return
}
