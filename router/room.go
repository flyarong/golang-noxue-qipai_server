package router

import (
	"encoding/json"
	"qipai/domain"
	"qipai/game"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"strconv"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqCreateRoom, createRoom)
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
	id, _ := strconv.Atoi(s.GetUserID())
	room.Uid = uint(id)

	if ok := utils.Copy(form, &room); !ok {
		resMsg = utils.Msg("房间信息赋值失败，请联系管理员").Code(-8)
		return
	}

	if err := srv.Room.Create(&room); err != nil {
		resMsg = utils.Msg(err.Error()).Code(-9)
		return
	}

	resMsg = utils.Msg("创建成功").AddData("id", room.ID)
}
