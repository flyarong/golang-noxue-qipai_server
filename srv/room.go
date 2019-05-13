package srv

import (
	"errors"
	"math/rand"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
	"time"
)

var Room roomSrv

type roomSrv struct {
}

func (this *roomSrv) Create(room *model.Room, delete bool) (err error) {
	dao.Db.Save(room)
	if room.ID == 0 {
		err = errors.New("房间添加失败，请联系管理员")
		return
	}

	// 如果需要删除，十分钟后，未开始游戏，房间自动删除
	if delete {
		go func() {
			time.Sleep(time.Second * 10)
			this.Delete(room.ID)
		}()
	}
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

func (this *roomSrv) SitDown(rid, uid uint) (deskId int, err error) {
	var room model.Room
	room, err = this.Get(rid)
	if err != nil {
		return
	}

	// 判断是否已在其他房间坐下
	var p model.Player
	dao.Db.Where("desk_id<>0 and uid=? and room_id<>?", uid, rid).First(&p)
	if p.ID > 0 {
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
			err=nil
		}
		return
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
	return
}

func (this *roomSrv) Players(roomId uint) (players []model.Player) {
	dao.Db.Where(&model.Player{RoomId: roomId}).Find(&players)
	return
}
