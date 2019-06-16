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
	game.AddAuthHandler(game.ReqJoinClub, reqJoinClub)
	game.AddAuthHandler(game.ReqClubs, reqClubs)
	game.AddAuthHandler(game.ReqEditClub, reqEditClub)
	game.AddAuthHandler(game.ReqClub, reqClub)
	game.AddAuthHandler(game.ReqClubUsers, reqClubUsers)
	game.AddAuthHandler(game.ReqDelClub, reqDelClub)
}

func reqDelClub(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId uint `json:"clubId"`
	}

	res := utils.Msg("")
	defer func() {
		res.Send(game.ResClubUsers, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)
	club,e:=dao.Club.Get(data.ClubId)
	if e!=nil{
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	if club.Uid != p.Uid  {
		res = utils.Msg("您不是茶楼老板，无法解散茶楼!")
		return
	}

	e=dao.Club.Del(data.ClubId)
	if e!=nil{
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
}

func reqClubUsers(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId uint `json:"clubId"`
	}

	res := utils.Msg("")
	defer func() {
		res.Send(game.ResClubUsers, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)
	// 只能看到自己加入的俱乐部的用户列表
	if !srv.Club.IsClubUser(uint(p.Uid), data.ClubId) {
		res = utils.Msg("你不属于该俱乐部，无法查看该俱乐部用户列表").Code(-1)
		return
	}

	users := srv.Club.Users(data.ClubId)

	res.AddData("users", users)
}

func reqJoinClub(s *zero.Session, msg *zero.Message) {

	type reqData struct {
		ClubId uint `json:"clubId"`
	}

	res := utils.Msg("")
	defer func() {
		res.Send(game.BroadcastJoinClub, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)

	err = srv.Club.Join(data.ClubId, uint(p.Uid))
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	res.AddData("clubId", data.ClubId).AddData("uid", p.Uid)
}

func reqEditClub(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId   uint   `json:"clubId"`
		Check    bool   `json:"check"`
		Close    bool   `json:"close"`
		Name     string `json:"name"`
		RollText string `json:"rollText"`
		Notice   string `json:"notice"`
	}

	res := utils.Msg("")
	defer func() {
		res.Send(game.BroadcastEditClub, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)

	if err := srv.Club.UpdateInfo(data.ClubId, uint(p.Uid), data.Check, data.Close, data.Name, data.RollText, data.Notice); err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	res.AddData("clubId", data.ClubId)
}

func reqClub(s *zero.Session, msg *zero.Message) {
	type clubV struct {
		Id        uint           `json:"id" xml:"ID"`
		Name      string         `json:"name"`      // 俱乐部名称
		Check     bool           `json:"check"`     // 是否审查
		Notice    string         `json:"notice"`    // 公告
		RollText  string         `json:"rollText"`  // 俱乐部大厅滚动文字
		Score     enum.ScoreType `json:"score"`     // 底分 以竖线分割的底分方式
		Players   int            `json:"players"`   // 玩家个数
		Count     int            `json:"count"`     // 局数
		StartType enum.StartType `json:"startType"` // 游戏开始方式 只支持1 首位开始
		Pay       enum.PayType   `json:"pay"`       // 付款方式 0 俱乐部老板付 1 AA
		Times     enum.TimesType `json:"times"`     // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
		Special   int            `json:"special"`   // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
		King      enum.KingType  `json:"king"`      // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
		Uid       uint           `json:"uid"`       // 老板
		Close     bool           `json:"close"`     // 是否打烊
		PayerUid  uint           `json:"payerUid"`  // 代付用户id
		BossNick  string         `json:"boss"`
	}

	type reqData struct {
		ClubId uint `json:"clubId"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResClub, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p := game.GetPlayerFromSession(s)

	club, err := srv.Club.GetClub(uint(p.Uid), data.ClubId)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	var cv clubV

	if !utils.Copy(club, &cv) {
		res = utils.Msg("内容转换出错").Code(-1)
		return
	}

	res = utils.Msg("").AddData("club", cv)
}

func reqClubs(s *zero.Session, msg *zero.Message) {
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
	resMsg = utils.Msg("").AddData("clubs", clubsV)
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
	u, e := dao.User.Get(uint(p.Uid))
	if e != nil {
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
