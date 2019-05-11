package srv

import (
	"errors"
	"math/rand"
	"qipai/dao"
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
			time.Sleep(time.Minute * 10)
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

func (roomSrv) Delete(roomId uint) {
	dao.Db.Where("id=? and status=0", roomId).Delete(&model.Room{})
}

func (roomSrv) MyRooms(uid uint) (rooms []model.Room) {
	// select r.* from rooms r join  players p on p.room_id=r.id where p.uid=100000;
	dao.Db.Raw("select r.* from rooms r join  players p on p.room_id=r.id where p.uid=?", uid).Scan(&rooms)
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

func (this *roomSrv) SitDown(rid, uid uint) (deskId int,err error) {
	var room model.Room
	room, err = this.Get(rid)
	if err != nil {
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
	var player model.Player
	dao.Db.Where("uid=? and room_id=?", uid, rid).First(&player)
	if player.ID == 0 {
		err = errors.New("用户未进入房间，如果已进入，可以尝试退出房间重新进入")
		return
	}
	player.DeskId = deskId
	dao.Db.Save(&player)

	return
}

func (this *roomSrv) Join(rid, uid uint, nick string) (err error) {

	// 检查房间号是否存在
	if !this.RoomExists(rid) {
		err = errors.New("该房间不存在")
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

func (this *roomSrv) Players(roomId uint) (players []model.Player) {
	dao.Db.Where(&model.Player{RoomId: roomId}).Find(&players)
	return
}
