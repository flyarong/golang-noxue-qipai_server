package router

import (
	"encoding/json"
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
}

func joinRoom(s *zero.Session, msg *zero.Message) {

	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.ToSend(game.ResJoinRoom, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)

	err = srv.Room.Join(data.Id, uint(p.Uid), p.Nick)
	if err != nil {
		if err.Error() == "该房间不存在，或已解散" {
			res = nil
			utils.Msg("房间超过10分钟未开始或已经结束，自动解散").AddData("id", data.Id).ToSend(game.ResDeleteRoom, s)
			return
		}
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
}

func room(s *zero.Session, msg *zero.Message) {

	type reqRoom struct {
		Id uint `json:"id"`
	}

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
		if res == nil {
			return
		}
		res.ToSend(game.ResRoom, s)
	}()

	var data reqRoom
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	room, err := srv.Room.Get(data.Id)
	if err != nil {
		if err.Error() == "该房间不存在，或游戏已结束" {
			res = nil
			utils.Msg("房间超过10分钟未开始或已经结束，自动解散").AddData("id", data.Id).ToSend(game.ResDeleteRoom, s)
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
	resMsg := utils.Msg("")
	defer func() {
		resMsg.ToSend(game.ResCreateRoom, s)
	}()

	var form domain.ReqCreateRoom
	err := json.Unmarshal(msg.GetData(), &form)
	if err != nil {
		resMsg = utils.Msg(err.Error()).Code(-1)
		return
	}

	// 限制只能 10  20 30 局
	if form.Count != 10 && form.Count != 20 && form.Count != 30 {
		resMsg = utils.Msg("局数[count]只能是10/20/30").Code(-2)
		return
	}

	// 限制游戏开始方式
	if form.StartType != 0 && form.StartType != 1 {
		resMsg = utils.Msg("开始方式[start]只能是0或1").Code(-3)
		return
	}

	// 限制支付模式
	if form.Pay != 0 && form.Pay != 1 {
		resMsg = utils.Msg("支付方式[pay]只能是0或1").Code(-4)
		return
	}

	// 限制翻倍规则
	if form.Times < 0 || form.Times > 4 {
		resMsg = utils.Msg("翻倍规则[times]取值不合法，只能在0-4之间").Code(-7)
		return
	}

	// 底分取值不合法
	if form.Score < 0 || form.Score > 5 {
		resMsg = utils.Msg("底分类型取值只能在0-5之间").Code(-7)
		return
	}

	var room model.Room
	p := game.GetPlayerFromSession(s)
	room.Uid = uint(p.Uid)

	if ok := utils.Copy(form, &room); !ok {
		resMsg = utils.Msg("房间信息赋值失败，请联系管理员").Code(-8)
		return
	}

	if err := srv.Room.Create(&room); err != nil {
		resMsg = utils.Msg(err.Error()).Code(-9)
		return
	}

	err = srv.Room.Join(room.ID, room.Uid, p.Nick)
	if err != nil {
		resMsg = utils.Msg(err.Error()).Code(-10)
		return
	}

	resMsg = utils.Msg("创建成功").AddData("id", room.ID)
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

	resMsg := utils.Msg("")

	defer func() {
		resMsg.ToSend(game.ResRoomList, s)
	}()

	p := game.GetPlayerFromSession(s)

	rooms := srv.Room.MyRooms(uint(p.Uid))
	var roomsV []roomV
	for _, v := range rooms {
		var r roomV
		if !utils.Copy(v, &r) {
			resMsg = utils.Msg("内容转换出错").Code(-1)
			return
		}
		roomsV = append(roomsV, r)
	}
	resMsg = utils.Msg("获取房间列表成功").AddData("rooms", roomsV)
}
