package router

import (
	"encoding/json"
	"github.com/golang/glog"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqCreateClub, createClub)
	game.AddAuthHandler(game.ReqClubs, resClubs)
}

func resClubs(s *zero.Session, msg *zero.Message) {
	type clubV struct {
		Id      uint           `json:"id"`
		Score   enum.ScoreType `json:"score"`
		Pay     enum.PayType   `json:"pay"`
		Count   int            `json:"count"`
		Boss    string         `json:"boss"`
		BossUid uint           `json:"bossUid"`
	}

	resMsg := utils.Msg("")

	defer func() {
		resMsg.Send(game.ResClubs, s)
	}()

	p := game.GetPlayerFromSession(s)

	var clubsV []clubV
	for _, v := range srv.Club.MyClubs(uint(p.Uid)) {
		clubsV = append(clubsV, clubV{
			Id:      v.ID,
			Score:   v.Score,
			Pay:     v.Pay,
			Count:   v.Count,
			Boss:    v.BossNick,
			BossUid: v.Uid,
		})
	}
	resMsg  = utils.Msg("").AddData("clubs", clubsV)
}

func createClub(s *zero.Session, msg *zero.Message) {
	resMsg := utils.Msg("")
	defer func() {
		resMsg.Send(game.ResCreateClub, s)
	}()


	type reqForm struct {
		Players   int            `json:"players"`
		Score     enum.ScoreType `json:"score"`
		Pay       enum.PayType   `json:"pay"`
		Count     int            `json:"count"`
		StartType enum.StartType `json:"start"`
		Times     int            `json:"times"`
	}

	var form reqForm
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

	var club model.Club
	p := game.GetPlayerFromSession(s)
	club.Uid = uint(p.Uid)

	if ok := utils.Copy(form, &club); !ok {
		resMsg = utils.Msg("茶楼信息赋值失败，请联系管理员").Code(-8)
		return
	}

	// 如果是老板支付，就默认需要审核才能进入俱乐部
	if club.Pay == enum.PayBoss {
		club.Check = true
	} else if club.Pay == enum.PayAA {
		club.Check = false
	}

	// 填充昵称
	u,e:=dao.User.Get(uint(p.Uid))
	if e!= nil {
		glog.Errorln(e)
		resMsg = utils.Msg(e.Error()).Code(-7)
		return
	}
	club.BossNick = u.Nick

	if err := srv.Club.Create(&club); err != nil {
		resMsg = utils.Msg(err.Error()).Code(-9)
		return
	}

	err = srv.Club.Join(club.ID, club.Uid)
	if err != nil {
		resMsg = utils.Msg(err.Error()).Code(-10)
		return
	}

	resMsg = utils.Msg("").AddData("clubId", club.ID)
}