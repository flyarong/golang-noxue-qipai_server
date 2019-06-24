package srv

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
	"qipai/utils"
	"time"
)

var Room roomSrv

type roomSrv struct {
}

func deleteAllInvalidRooms() {
	var rooms []model.Room
	res := dao.Db().Find(&rooms)
	if res.Error != nil {
		glog.Errorln(res.Error)
		return
	}
	for _, room := range rooms {
		// 超过10分钟，游戏没开始的房间，并且不是俱乐部房间，自动解散
		if isRoomExpired(&room) {
			deletePlayersInRoom(room.ID)
			dao.Db().Delete(&room)
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
func isRoomExpired(room *model.Room) bool {
	// 超过10分钟，游戏没开始的房间，并且不是俱乐部房间，自动解散
	return (time.Now().Sub(room.CreatedAt) >= time.Minute*10 && room.ClubId == 0 && room.Status == enum.GameReady)
}

// 删除过期的房间，并通知客户端
func deleteExpiredRoom(roomId uint) (err error) {
	var room model.Room
	res := dao.Db().Find(&room, roomId)
	if res.Error != nil {
		glog.Errorln(res.Error)
		return
	}
	if res.RecordNotFound() {
		err = errors.New("没有找到房间")
		return
	}
	if isRoomExpired(&room) {
		deletePlayersInRoom(room.ID)
		dao.Room.Delete(room.ID)
	}
	return
}

// 删除房间里面的用户并通知他们
func deletePlayersInRoom(roomId uint) (err error) {
	var ps []model.Player
	res := dao.Db().Where(&model.Player{RoomId: roomId}).Find(&ps)
	if res.Error != nil {
		err = errors.New(fmt.Sprint("查询房间对应的玩家失败，房间号：%d", roomId))
		return
	}
	for _, v := range ps {
		if p := game.GetPlayer(v.Uid); p != nil {
			// 通知在线的相关用户
			utils.Msg("").AddData("roomId", roomId).Send(game.ResDeleteRoom, p.Session)
		}
		dao.Db().Unscoped().Delete(&v)
	}
	dao.Db().Delete(model.Player{}, "room_id=?", roomId)
	return
}

func (this *roomSrv) Create(room *model.Room) (err error) {
	res := dao.Db().Save(room)
	if res.Error != nil || res.RowsAffected == 0 {
		err = errors.New("房间添加失败，请联系管理员")
		return
	}

	go func() {
		time.Sleep(time.Minute * 10)
		deleteAllInvalidRooms()
	}()

	return
}

/**
删除房间，并通知房间内的人
*/
func (this *roomSrv) Delete(roomId, uid uint) (err error) {

	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	if room.Uid != uid {
		err = errors.New("您不是该房间的房主，无权解散")
		return
	}

	if room.Status != 0 {
		err = errors.New("游戏已开始，无法删除房间")
		return
	}

	err = dao.Room.Delete(roomId)
	if err != nil {
		return
	}

	err = deletePlayersInRoom(roomId)
	return
}

func (this *roomSrv) SitDown(rid, uid uint) (roomId uint, deskId int, err error) {
	var room model.Room
	room, err = dao.Room.Get(rid)
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
	player, err = dao.Game.Player(rid, uid)
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
	dao.Db().Model(&model.Player{}).Where("desk_id>0 and room_id=?", rid).Count(&n)
	if n >= room.Players {
		err = errors.New("当前房间已坐满")
		return
	}

	// 删除已经有人的座位
	players := dao.Game.Players(rid)
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
	dao.Db().Model(&player).Update(&model.Player{DeskId:deskId,JoinedAt:&t})

	return
}

func (this *roomSrv) Join(rid, uid uint) (err error) {

	room, e := dao.Room.Get(rid)
	if e != nil {
		glog.Error(e)
		err = errors.New("该房间不存在，或已解散")
		return
	}

	// 正在游戏的房间无法进入
	if room.Status == enum.GamePlaying {
		err = errors.New("该房间正在游戏中，无法进入!")
		return
	}

	// 检测是不是退出后重新进入的玩家
	players := dao.Room.PlayersSitDown(rid)
	ok := false
	for _, v := range players {
		if v.Uid == uid && v.DeskId > 0 {
			ok = true
		}
	}
	// 游戏中无法加入,防止别人扫描哪些房间存在，游戏中的房间和不存在的提示信息一样
	if room.Status == enum.GamePlaying && !ok {
		err = errors.New("该房间不存在")
		return
	}

	if dao.Room.IsRoomPlayer(rid, uid) {
		return
	}

	user, e := dao.User.Get(uid)
	if e != nil {
		err = e
		return
	}
	ru := model.Player{
		Uid:    uid,
		RoomId: rid,
		Nick:   user.Nick,
		Avatar: user.Avatar,
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
	defer func() {
		if err == nil {
			this.SendExit(rid, uid)
			if ret := dao.Db().Model(model.Player{}).Where("uid=? and room_id=?", uid, rid).Update(map[string]interface{}{"desk_id": 0, "joined_at": nil}); ret.Error != nil {
				glog.Error(ret.Error)
				return
			}
		}
	}()

	// 游戏开始后无法退出
	room, e := dao.Room.Get(rid)
	if e != nil {
		err = e
		return
	}
	if room.Status == enum.GamePlaying {
		err = errors.New("游戏中，无法退出")
		return
	}

	return
}

func (this *roomSrv) SendExit(rid, uid uint) {
	// 通知其他客户端玩家，我退出了
	game.SendToAllPlayers(utils.Msg("").AddData("roomId", rid).AddData("uid", uid), game.ResLeaveRoom, rid)
}
