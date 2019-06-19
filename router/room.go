package router

import (
	"encoding/json"
	"github.com/golang/glog"
	"qipai/dao"
	"qipai/domain"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqCreateRoom, createRoom)
	game.AddAuthHandler(game.ReqRoomList, roomList)
	game.AddAuthHandler(game.ReqRoom, room)         // 请求房间信息
	game.AddAuthHandler(game.ReqJoinRoom, joinRoom) // 请求加入房间
	game.AddAuthHandler(game.ReqSit, sit)
	game.AddAuthHandler(game.ReqLeaveRoom, leaveRoom)
	game.AddAuthHandler(game.ReqDeleteRoom, deleteRoom)
}

func deleteRoom(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	res = nil
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResDeleteRoom, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}

	err = srv.Room.Delete(data.Id, uint(p.Uid))
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
}

func leaveRoom(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResLeaveRoom, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}

	err = srv.Room.Exit(data.Id, uint(p.Uid))
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	res = nil
}

func sit(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResSit, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}

	roomId, deskId, e := srv.Room.SitDown(data.Id, uint(p.Uid))
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1).AddData("roomId", roomId)
		return
	}
	res = utils.Msg("").AddData("deskId", deskId)

	// 获取当前房间所有玩家
	type playerV struct {
		Uid        uint `json:"uid"`        // 用户编号
		DeskId     int  `json:"deskId"`     // 座位号
		TotalScore int  `json:"totalScore"` // 玩家总分
	}

	players := dao.Room.PlayersSitDown(data.Id)
	var pvs []playerV
	for _, v := range players {
		var pv playerV
		if !utils.Copy(v, &pv) {
			res = utils.Msg("玩家数组赋值出错，请联系管理员").Code(-1)
			return
		}
		pvs = append(pvs, pv)
	}

	// 通知房间中其他坐下的玩家，我坐下了
	for _, v := range pvs {
		// 不用通知自己
		if v.Uid == uint(p.Uid) {
			continue
		}
		otherPlayer := game.GetPlayer(v.Uid)
		if otherPlayer == nil {
			glog.Errorln("通知其他用户有用户坐下失败")
			continue
		}
		utils.Msg("").
			AddData("roomId", data.Id).
			AddData("uid", p.Uid).
			AddData("deskId", deskId).Send(game.BroadcastSitRoom, otherPlayer.Session)
	}

	res.AddData("uid", p.Uid).AddData("players", pvs)
}

func joinRoom(s *zero.Session, msg *zero.Message) {

	type reqData struct {
		RoomId uint `json:"roomId"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResJoinRoom, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}

	err = srv.Room.Join(data.RoomId, uint(p.Uid), p.Nick)
	if err != nil {
		if err.Error() == "该房间不存在，或已解散" {
			res = nil
			utils.Msg("房间超过10分钟未开始或已经结束，自动解散").AddData("roomId", data.RoomId).Send(game.ResDeleteRoom, s)
			return
		}
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	// 获取当前房间所有玩家
	type playerV struct {
		Uid    uint `json:"uid"`    // 用户编号
		DeskId int  `json:"deskId"` // 座位号
	}

	players := dao.Room.PlayersSitDown(data.RoomId)
	var pvs []playerV
	for _, v := range players {
		var pv playerV
		if !utils.Copy(v, &pv) {
			res = utils.Msg("玩家数组赋值出错，请联系管理员").Code(-1)
			return
		}
		pvs = append(pvs, pv)
	}
	res.AddData("players", pvs)
}

func room(s *zero.Session, msg *zero.Message) {

	type reqRoom struct {
		Id uint `json:"id"`
	}

	type roomV struct {
		ID        uint           `json:"id"`
		Score     enum.ScoreType `json:"score"`     // 底分类型
		Pay       enum.PayType   `json:"pay"`       // 支付方式
		Current   int            `json:"current"`   // 当前第几局
		Count     int            `json:"count"`     // 总共可以玩几局
		Uid       uint           `json:"uid"`       // 房主用户编号
		StartType enum.StartType            `json:"startType"` // 游戏开始方式
		Players   int            `json:"players"`   // 玩家个数
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResRoom, s)
	}()

	var data reqRoom
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	room, err := dao.Room.Get(data.Id)
	if err != nil {
		if err.Error() == "该房间不存在，或游戏已结束" {
			res = nil
			utils.Msg("房间超过10分钟未开始或已经结束，自动解散").AddData("id", data.Id).Send(game.ResDeleteRoom, s)
			return
		}
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	var rv roomV
	if !utils.Copy(room, &rv) {
		res = utils.Msg("复制房间信息出错，请联系管理员").Code(-1)
		return
	}

	res = utils.Msg("").AddData("room", rv)
}

func createRoom(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		res.Send(game.ResCreateRoom, s)
	}()

	var form domain.ReqCreateRoom
	err := json.Unmarshal(msg.GetData(), &form)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	// 限制只能 10  20 30 局
	if form.Count != 10 && form.Count != 20 && form.Count != 30 {
		res = utils.Msg("局数[count]只能是10/20/30").Code(-2)
		return
	}

	// 限制游戏开始方式
	if form.StartType != 0 && form.StartType != 1 {
		res = utils.Msg("开始方式[start]只能是0或1").Code(-3)
		return
	}

	// 限制支付模式
	if form.Pay != 0 && form.Pay != 1 {
		res = utils.Msg("支付方式[pay]只能是0或1").Code(-4)
		return
	}

	// 限制翻倍规则
	if form.Times < 0 || form.Times > 4 {
		res = utils.Msg("翻倍规则[times]取值不合法，只能在0-4之间").Code(-7)
		return
	}

	// 底分取值不合法
	if form.Score < 0 || form.Score > 5 {
		res = utils.Msg("底分类型取值只能在0-5之间").Code(-7)
		return
	}

	var room model.Room
	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}
	room.Uid = uint(p.Uid)

	if ok := utils.Copy(form, &room); !ok {
		res = utils.Msg("房间信息赋值失败，请联系管理员").Code(-8)
		return
	}

	if err := srv.Room.Create(&room); err != nil {
		res = utils.Msg(err.Error()).Code(-9)
		return
	}

	err = srv.Room.Join(room.ID, room.Uid, p.Nick)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-10)
		return
	}

	res = utils.Msg("创建成功").AddData("roomId", room.ID)
}

func roomList(s *zero.Session, msg *zero.Message) {

	type roomV struct {
		ID      uint           `json:"id"`
		Score   enum.ScoreType `json:"score"`   // 底分类型
		Pay     enum.PayType   `json:"pay"`     // 支付方式
		Current int            `json:"current"` // 当前第几局
		Count   int            `json:"count"`   // 总共可以玩几局
		Uid     uint           `json:"uid"`     // 房主用户编号
		Players int            `json:"players"` // 玩家个数
	}

	res := utils.Msg("")

	defer func() {
		res.Send(game.ResRoomList, s)
	}()

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
	}

	rooms := dao.Room.MyRooms(uint(p.Uid))
	var roomsV []roomV
	for _, v := range rooms {
		var r roomV
		if !utils.Copy(v, &r) {
			res = utils.Msg("内容转换出错").Code(-1)
			return
		}
		roomsV = append(roomsV, r)
	}
	res = utils.Msg("获取房间列表成功").AddData("rooms", roomsV)
}
