package router

import (
	"encoding/json"
	"github.com/golang/glog"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/srv"
	"qipai/utils"
	"time"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqUserInfo, userInfo)
	game.AddHandler(game.ReqReset, reqReset)
	game.AddAuthHandler(game.ReqNotice, reqNotice)
	game.AddAuthHandler(game.ReqRollText, reqRollText)
	game.AddAuthHandler(game.ReqShareText, reqShareText)
	game.AddAuthHandler(game.ReqDefaultVoice, reqDefaultVoice)
}

func reqDefaultVoice(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		RoomId uint `json:"roomId"`
		DeskId int `json:"deskId"`
		VoiceId int `json:"voiceId"`
		Sex int `json:"sex"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.BroadcastDefaultVoice, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		glog.Error(err)
		res = utils.Msg("请求的数据格式不正确").Code(-1)
		return
	}

	players := dao.Room.PlayersSitDown(data.RoomId)
	// 通知房间中所有的玩家，我说话了
	for _, v := range players {
		otherPlayer := game.GetPlayer(v.Uid)
		if otherPlayer == nil {
			glog.Infoln("通知其他用户有用户坐下失败")
			continue
		}
		_=utils.Msg("").
			AddData("roomId", data.RoomId).
			AddData("deskId", data.DeskId).
			AddData("voiceId", data.VoiceId).
			AddData("sex", data.Sex).
			Send(game.BroadcastDefaultVoice, otherPlayer.Session)
	}
	res = nil
}

func reqRollText(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResRollText, s)
	}()

	rollText := utils.Lv.Get("user_rollText")
	res.AddData("rollText", rollText)
}

func reqShareText(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResShareText, s)
	}()

	shareText := utils.Lv.Get("user_shareText")
	res.AddData("shareText", shareText)
}

func reqNotice(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResNotice, s)
	}()

	notice := utils.Lv.Get("user_notice")
	res.AddData("notice", notice)
}

func reqReset(s *zero.Session, msg *zero.Message) {

	type ReqReset struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
		Code     string        `form:"code" json:"code" binding:"required"`
	}

	res := utils.Msg("")
	defer func() {
		res.Send(game.ResReset, s)
	}()
	var data ReqReset
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		return
	}
	if data.UserType != enum.MobilePass {
		res = utils.Msg("目前仅支持手机重置密码").Code(-1)
		return
	}

	// 检查手机验证码，无论对错都删除验证码，防止暴力破解
	code := utils.Lv.Get("code_" + data.Name)
	utils.Lv.Del("code_" + data.Name)
	if code != data.Code {
		res = utils.Msg("手机验证码错误").Code(-1)
		return
	}

	e := srv.User.ChangePass(data.UserType, data.Name, data.Pass)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	return
}

func userInfo(s *zero.Session, msg *zero.Message) {
	type userV struct {
		ID        uint      `json:"id"`
		Nick      string    `json:"nick"`
		Avatar    string    `json:"avatar"`
		Sex       int       `json:"sex"`
		Ip        string    `json:"ip"`
		Address   string    `json:"address"`
		Card      int       `json:"card"`
		CreatedAt time.Time `json:"createdAt"`
	}

	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResUserInfo, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	if data.Id == 0 {
		p, e := game.GetPlayerFromSession(s)
		if e != nil {
			glog.Error(e)
			res = utils.Msg(e.Error()).Code(-1)
		}
		data.Id = uint(p.Uid)
	}

	user, err := srv.User.GetInfo(data.Id)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	var uv userV
	utils.Copy(user, &uv)

	res = utils.Msg("").AddData("user", uv)
}
