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
	game.AddAuthHandler(game.ReqCreateClub, createClub)
	game.AddAuthHandler(game.ReqEditClubRoom, editClubRoom)
	game.AddAuthHandler(game.ReqJoinClub, reqJoinClub)
	game.AddAuthHandler(game.ReqClubs, reqClubs)
	game.AddAuthHandler(game.ReqEditClub, reqEditClub)
	game.AddAuthHandler(game.ReqClub, reqClub)
	game.AddAuthHandler(game.ReqClubUsers, reqClubUsers)
	game.AddAuthHandler(game.ReqDelClub, reqDelClub)
	game.AddAuthHandler(game.ReqEditClubUser, reqEditClubUser)
	game.AddAuthHandler(game.ReqCreateClubRoom, reqCreateClubRoom)
	game.AddAuthHandler(game.ReqExitClub, reqExitClub)
	game.AddAuthHandler(game.ReqClubRoomUsers, reqClubRoomUsers)
	game.AddAuthHandler(game.ReqClubRooms, reqClubRooms)
}

func reqClubRooms(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId uint `json:"clubId"`
	}
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResClubRooms, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	var rvs []domain.ResRoomV
	var rooms []model.Room
	ret := dao.Db().Where(model.Room{ClubId: data.ClubId}).Find(&rooms)
	if ret.RowsAffected == 0 {
		res.AddData("rooms", rvs)
		return
	}

	for _, v := range rooms {
		var rv domain.ResRoomV
		if !utils.Copy(v, &rv) {
			res = utils.Msg("复制房间信息出错，请联系管理员").Code(-1)
			return
		}
		rvs = append(rvs, rv)
	}
	res.AddData("rooms", rvs)
}

func reqClubRoomUsers(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		RoomId uint `json:"roomId"`
	}
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResClubRoomUsers, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	room, e := dao.Room.Get(data.RoomId)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	if room.TableId == 0 {
		res = utils.Msg("该房间不是茶楼的房间").Code(-1)
		return
	}

	type ClubRoomUser struct {
		Uid    uint   `json:"uid"`    // 用户编号
		Nick   string `json:"nick"`   // 昵称
		Avatar string `json:"avatar"` // 用户头像
		DeskId int    `json:"deskId"` // 座位号
	}
	players := dao.Room.PlayersSitDown(data.RoomId)
	var crus []ClubRoomUser
	for _, v := range players {
		var cru ClubRoomUser
		utils.Copy(v, &cru)
		crus = append(crus, cru)
	}
	res.AddData("users", crus).AddData("roomId", room.ID).AddData("clubId", room.ClubId).AddData("tableId", room.TableId)
}

// 退出茶楼，把用户从茶楼在线列表中删除，无须返回成功与否
func reqExitClub(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId uint `json:"clubId"`
	}
	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResExitClub, s)
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
		return
	}
	// 如果是暂时离开茶楼，就简单的从在线列表删除该玩家即可
	if data.ClubId == 0 {
		game.ClubPlayers.Del(p.Uid)
		res = nil
		return
	}

	// 如果是退出茶楼，就从茶楼玩家中移除
	e = dao.Club.DelClubUser(data.ClubId, p.Uid)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	res.AddData("clubId", data.ClubId).AddData("uid", p.Uid)
}

func reqCreateClubRoom(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId  uint `json:"clubId"`
		TableId int  `json:"tableId"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResCreateClubRoom, s)
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
		return
	}

	// 该用户是否是当前茶楼用户，并且检测有没有被封
	user, e := dao.Club.GetUser(data.ClubId, p.Uid)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	if user.Status == enum.ClubUserDisable {
		res = utils.Msg("您已被茶楼管理员冻结，请联系茶楼管理员！").Code(-1)
		return
	} else if user.Status == enum.ClubUserWait {
		res = utils.Msg("您的账号正在等待管理员审核中，请联系管理员解除！").Code(-1)
		return
	}

	club, e := dao.Club.Get(data.ClubId)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	// 如果tableId指定的位置已经存在房间
	var r model.Room
	ret := dao.Db().Where(&model.Room{ClubId: data.ClubId, TableId: data.TableId}).First(&r)
	if !ret.RecordNotFound() {
		// 如果房里有人，直接返回房间号
		if len(dao.Room.PlayersSitDown(r.ID)) > 0 {
			res = utils.Msg("").AddData("clubId", r.ClubId).AddData("roomId", r.ID).AddData("uid", p.Uid)
			return
		}
		// 否则删除房间重新创建
		srv.Room.Delete(r.ID, r.Uid)
	}

	var room model.Room

	// 获取特殊房间的信息
	var cr model.ClubRoom
	ret = dao.Db().Model(model.ClubRoom{}).Where("club_id=? and table_id=?", data.ClubId, data.TableId).First(&cr)
	if ret.RecordNotFound() {
		if ok := utils.Copy(club, &room); !ok {
			res = utils.Msg("房间信息赋值失败，请联系管理员").Code(-1)
			return
		}
	} else {
		if ok := utils.Copy(cr, &room); !ok {
			res = utils.Msg("房间信息赋值失败，请联系管理员").Code(-1)
			return
		}
	}

	room.ID = 0
	room.TableId = data.TableId
	room.ClubId = club.ID
	room.Uid = club.Uid
	room.BossNick = club.BossNick
	room.Pay = club.Pay // 统一支付方式

	if err := srv.Room.Create(&room); err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	err = srv.Room.Join(room.ID, room.Uid)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	res = utils.Msg("").AddData("clubId", room.ClubId).AddData("roomId", room.ID).AddData("uid", p.Uid)
	// 通知当前房间所属茶楼的所有正在茶楼的玩家，有新房间创建了
	game.NotifyClubPlayers(game.ResCreateClubRoom, room.ID, utils.Msg("").AddData("uid", p.Uid))
}

func reqEditClubUser(s *zero.Session, msg *zero.Message) {

	type reqData struct {
		ClubId uint   `json:"clubId"`
		Uid    uint   `json:"uid"`
		Action string `json:"action"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResEditClubUser, s)
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
		return
	}

	// 编辑会员状态：设为管理(admin) 取消管理(-admin)  冻结(disable) 取消冻结(-disable) 设为代付(pay) 取消代付(-pay) 审核通过用户(add)  移除用户(-add)
	action := data.Action

	glog.V(3).Infoln("编辑用户：", p.Nick, "\t", action)

	isAdmin := srv.Club.IsAdmin(p.Uid, data.ClubId)
	isBoss := srv.Club.IsBoss(p.Uid, data.ClubId)
	// 只有管理员或创建者可以操作
	if !isAdmin && !isBoss {
		res = utils.Msg("您不是管理员或老板，无法操作！").Code(-1)
		return
	}

	// 自己不能编辑自己
	if p.Uid == data.Uid {
		res = utils.Msg("您不能对自己进行操作！").Code(-1)
		return
	}

	err = nil

	switch action {
	case "admin":
		// 只有老板可以设置管理员
		if !isBoss {
			res = utils.Msg("您不是老板，无法设置管理员！").Code(-1)
			return
		}
		err = srv.Club.SetAdmin(data.ClubId, data.Uid, true)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "-admin":
		// 只有老板可以取消管理员
		if !isBoss {
			res = utils.Msg("您不是老板，无法取消管理员！").Code(-1)
			return
		}
		err = srv.Club.SetAdmin(data.ClubId, data.Uid, false)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "disable":
		// 管理员 不能冻结管理员或老板
		if isAdmin && (srv.Club.IsBoss(data.Uid, data.ClubId) || srv.Club.IsAdmin(data.Uid, data.ClubId)) {
			res = utils.Msg("管理员无法冻结其他管理员和老板").Code(-1)
			return
		}
		err = srv.Club.SetDisable(data.ClubId, data.Uid, true)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "-disable":
		// 管理员 不能接触冻结管理员或老板
		if isAdmin && (srv.Club.IsBoss(data.Uid, data.ClubId) || srv.Club.IsAdmin(data.Uid, data.ClubId)) {
			res = utils.Msg("管理员无法接触冻结管理员和老板").Code(-1)
			return
		}
		err = srv.Club.SetDisable(data.ClubId, data.Uid, false)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "pay":
		if !isBoss {
			res = utils.Msg("您不是老板，无法设置代付！").Code(-1)
			return
		}
		err = srv.Club.SetPay(data.ClubId, data.Uid, true)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "-pay":
		if !isBoss {
			res = utils.Msg("您不是老板，无法取消代付！").Code(-1)
			return
		}
		err = srv.Club.SetPay(data.ClubId, data.Uid, false)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "add":
		// 审核通过，就是设置为普通用户，跟取消冻结操作一样
		err = srv.Club.SetDisable(data.ClubId, data.Uid, false)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	case "-add":
		// 管理员 不能移除管理员或老板
		if isAdmin && (srv.Club.IsBoss(data.Uid, data.ClubId) || srv.Club.IsAdmin(data.Uid, data.ClubId)) {
			res = utils.Msg("管理员无法移除其他管理员和老板").Code(-1)
			return
		}
		err = srv.Club.RemoveClubUser(data.ClubId, data.Uid)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}
	default:
		res = utils.Msg("不支持这个操作:" + action).Code(-1)
		return
	}
	res.AddData("clubId", data.ClubId)
}

func reqDelClub(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		ClubId uint `json:"clubId"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.BroadcastDelClub, s)
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
		return
	}

	err = srv.Club.DelClub(data.ClubId, p.Uid)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	ret := dao.Db().Unscoped().Where("club_id=?", data.ClubId).Delete(model.ClubRoom{})
	if ret.Error != nil {
		glog.Error(ret.Error)
		res = utils.Msg("删除茶楼房间相关信息失败").Code(-1)
		return
	}
	res = nil
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

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	// 只能看到自己加入的茶楼的用户列表
	if !srv.Club.IsClubUser(uint(p.Uid), data.ClubId) {
		res = utils.Msg("你不属于该茶楼，无法查看该茶楼用户列表").Code(-1)
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

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	err = srv.Club.Join(data.ClubId, uint(p.Uid))
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	res.AddData("clubId", data.ClubId).AddData("uid", p.Uid)
	game.ClubPlayers.Add(data.ClubId, p.Uid, s) // 添加当前玩家到茶楼在线列表
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

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	if err := srv.Club.UpdateInfo(data.ClubId, uint(p.Uid), data.Check, data.Close, data.Name, data.RollText, data.Notice); err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	res.AddData("clubId", data.ClubId)
}

func reqClub(s *zero.Session, msg *zero.Message) {
	type clubV struct {
		Id        uint           `json:"id" xml:"ID"` // 在返回茶楼房间的时候，座位房间桌子编号
		Name      string         `json:"name"`        // 茶楼名称
		Check     bool           `json:"check"`       // 是否审查
		Notice    string         `json:"notice"`      // 公告
		RollText  string         `json:"rollText"`    // 茶楼大厅滚动文字
		Score     enum.ScoreType `json:"score"`       // 底分 以竖线分割的底分方式
		Players   int            `json:"players"`     // 玩家个数
		Count     int            `json:"count"`       // 局数
		StartType enum.StartType `json:"startType"`   // 游戏开始方式 只支持1 首位开始
		Pay       enum.PayType   `json:"pay"`         // 付款方式 0 茶楼老板付 1 AA
		Times     enum.TimesType `json:"times"`       // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
		Special   int            `json:"special"`     // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
		King      enum.KingType  `json:"king"`        // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
		Uid       uint           `json:"uid"`         // 老板
		Close     bool           `json:"close"`       // 是否打烊
		PayerUid  uint           `json:"payerUid"`    // 代付用户id
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

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

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

	// 获取特殊编辑过的桌子
	var tables []clubV
	var crs []model.ClubRoom
	ret := dao.Db().Where("club_id=?", data.ClubId).Find(&crs)
	if ret.Error != nil {
		glog.Error(ret.Error)
		res = utils.Msg("获取桌子信息出错").Code(-1)
		return
	}
	for _, v := range crs {
		var tv clubV
		if !utils.Copy(v, &tv) {
			res = utils.Msg("内容转换出错").Code(-1)
			return
		}
		tv.Id = uint(v.TableId) // 在这里id用来保存table编号
		tables = append(tables, tv)
	}

	res = utils.Msg("").AddData("club", cv).AddData("tables", tables)
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

	res := utils.Msg("")

	defer func() {
		res.Send(game.ResClubs, s)
	}()

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

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
	res = utils.Msg("").AddData("clubs", clubsV)
}

func createClub(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		res.Send(game.ResCreateClub, s)
	}()

	type reqForm struct {
		Players   int            `json:"players"`
		Score     enum.ScoreType `json:"score"`
		Pay       enum.PayType   `json:"pay"`
		Tui       bool           `json:"tui"`
		Count     int            `json:"count"`
		StartType enum.StartType `json:"start"`
		Times     int            `json:"times"`
	}

	var form reqForm
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

	var club model.Club
	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	club.Uid = uint(p.Uid)

	if ok := utils.Copy(form, &club); !ok {
		res = utils.Msg("茶楼信息赋值失败，请联系管理员").Code(-8)
		return
	}

	// 如果是老板支付，就默认需要审核才能进入茶楼
	if club.Pay == enum.PayBoss {
		club.Check = true
	} else if club.Pay == enum.PayAA {
		club.Check = false
	}

	// 填充昵称
	u, e := dao.User.Get(uint(p.Uid))
	if e != nil {
		glog.Errorln(e)
		res = utils.Msg(e.Error()).Code(-7)
		return
	}
	club.BossNick = u.Nick

	if err := srv.Club.Create(&club); err != nil {
		res = utils.Msg(err.Error()).Code(-9)
		return
	}

	err = srv.Club.Join(club.ID, club.Uid)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-10)
		return
	}

	res = utils.Msg("").AddData("clubId", club.ID)
}

func editClubRoom(s *zero.Session, msg *zero.Message) {
	res := utils.Msg("")
	defer func() {
		res.Send(game.ResEditClubRoom, s)
	}()

	type reqForm struct {
		ClubId    uint           `json:"clubId"`
		TableId   int            `json:"tableId"`
		Players   int            `json:"players"`
		Score     enum.ScoreType `json:"score"`
		Pay       enum.PayType   `json:"pay"`
		Count     int            `json:"count"`
		StartType enum.StartType `json:"start"`
		Times     int            `json:"times"`
		Tui       bool           `json:"tui"`
	}

	var form reqForm
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

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	club, err := dao.Club.Get(form.ClubId)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	// 不是管理员或者茶楼老板，无法操作
	if !srv.Club.IsBoss(p.Uid, club.ID) && !srv.Club.IsAdmin(p.Uid, club.ID) {
		res = utils.Msg("只有茶楼老板或管理员可以编辑房间信息").Code(-1)
		return
	}

	// 如果茶楼房间信息存在，就更新，不存在就添加
	var n int
	ret := dao.Db().Model(model.ClubRoom{}).Where("club_id=? and table_id=?", form.ClubId, form.TableId).Count(&n)
	if ret.Error != nil {
		glog.Error(ret.Error)
		res = utils.Msg("获取茶楼房间信息出错").Code(-1)
		return
	}

	var clubRoom model.ClubRoom
	if ok := utils.Copy(form, &clubRoom); !ok {
		res = utils.Msg("茶楼信息赋值失败，请联系管理员").Code(-8)
		return
	}

	if n == 0 { // 不存在，则添加
		ret = dao.Db().Create(&clubRoom)
		if ret.Error != nil {
			glog.Error(ret.Error)
			res = utils.Msg("添加茶楼房间信息出错").Code(-1)
			return
		}
	} else {
		ret = dao.Db().Model(model.ClubRoom{}).Where("club_id=? and table_id=?", form.ClubId, form.TableId).Update(
			map[string]interface{}{"players": clubRoom.Players, "count": clubRoom.Count, "score": clubRoom.Score, "tui": clubRoom.Tui, "times": clubRoom.Times},
		)
		if ret.Error != nil {
			glog.Error(ret.Error)
			res = utils.Msg("更新茶楼房间信息出错").Code(-1)
			return
		}
	}
	res.AddData("clubId", form.ClubId)
	return
}
