package srv

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
)

var Game gameSrv

type gameSrv struct {
}

func (this *gameSrv) Start(roomId, uid uint) (err error) {

	// 查找房间
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	players := dao.Room.PlayersSitDown(roomId)
	if len(players) < 2 {
		err = errors.New("少于2个玩家，无法开始")
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
		p, err = dao.Game.FirstPlayer(roomId)
		if err != nil {
			return
		}
		if p.Uid != uid {
			err = errors.New("您不是第一个进入房间的玩家，无权开始游戏")
			return
		}
	}
	room.Status = enum.GamePlaying

	// 扣除房卡
	if e = TakeUserCard(roomId, uid); e != nil {
		err = e
		return
	}

	// 更新游戏状态
	dao.Db().Model(&model.Room{}).Where("id=?", roomId).Update("status", enum.GamePlaying)

	g, e := game.Games.NewGame(roomId)
	if e != nil {
		err = e
		glog.Error(e)
		return
	}

	dao.Db().Model(&model.Room{}).Where("id=?", roomId).Update("status", enum.GamePlaying)

	g.Start()

	return
}

func TakeUserCard(roomId, uid uint) (err error) {

	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		glog.Error(e)
		return
	}

	var payUid uint = 0 // 0表示AA支付，大于0 表示具体uid指定的用户

	// 如果clubId大于0，表示该房间是属于某个俱乐部，那么就查询该俱乐部是否有代付
	if room.ClubId > 0 {
		club, e := dao.Club.Get(room.ClubId)
		if e != nil {
			glog.Errorln(e)
			err = e
			return
		}

		// 俱乐部如果有代付，就是代付优先
		if club.PayerUid > 0 {
			payUid = club.PayerUid
		} else if club.Pay == enum.PayBoss { // 老板支付
			payUid = club.Uid
		}
	} else if room.Pay == enum.PayBoss { // 如果不是俱乐部房间，又是老板支付，那就是房主支付
		payUid = room.Uid
	}

	//计算开始游戏需要多少房卡
	card := 0
	if payUid > 0 {
		card = (room.Count / 10) * (room.Players / 2)
	} else {
		card = room.Count / 10
	}

	// 如果是老板支付
	if payUid > 0 {
		payer, e := dao.User.Get(payUid)
		if e != nil {
			glog.Error(e)
			err = e
			return
		}
		// 如果老板房卡不足，返回错误提示
		if payer.Card < card {
			err = errors.New(fmt.Sprintf("老板房卡不足%d个", card))
			return
		}

		// 从老板账号扣款
		err = dao.User.TakeCard(payUid, card)
		if err != nil {
			glog.Errorln(err)
			return
		}
		return
	}

	// 以下是AA付款逻辑
	players := dao.Room.PlayersSitDown(roomId)
	// 坐下的玩家是否都有足够的卡
	var uids []uint
	for _, v := range players {
		uids = append(uids, v.Uid)
	}

	var users []model.User
	ret := dao.Db().Where("id in (?)", uids).Find(&users)
	if ret.Error != nil {
		err = errors.New("查询玩家列表数据出错")
		glog.Errorln(ret.Error)
		return
	}
	if ret.RecordNotFound() {
		err = errors.New("玩家列表为空")
		glog.Errorln(err)
		return
	}

	// 循环检查每个玩家是否都有足够的卡，只要有一个没有，就无法开始
	for _, v := range users {
		if v.Card < card {
			err = errors.New(fmt.Sprintf("玩家["+v.Nick+"]的钻石不足%d个", card))
			return
		}
	}

	// 每个玩家都有足够钻石的情况下，付款
	for _, v := range users {
		_ = dao.User.TakeCard(v.ID, card)
	}
	return
}

// 设置抢庄倍数
func (this *gameSrv) SetTimes(roomId, uid uint, times int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}
	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	if times < 0 || times > 4 {
		err = errors.New("抢庄倍数错误，只能是0-4倍")
		return
	}

	g, e := dao.Game.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经抢庄，直接返回
	if g.Times >= 0 {
		return
	}

	game1, e := game.Games.Get(roomId)
	if e != nil {
		err = e
		return
	}
	game1.SetTimes(uid, times, false)

	return
}

// 下注
func (this *gameSrv) SetScore(roomId, uid uint, score int) (err error) {
	room, e := dao.Room.Get(roomId)
	if e != nil {
		err = e
		return
	}

	if room.Status != enum.GamePlaying {
		err = errors.New("游戏未开始，无法下注")
		return
	}

	ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
	s := ss[room.Score]

	if s[0] != score && s[1] != score {
		err = errors.New("积分不合法")
		return
	}

	g, e := dao.Game.GetCurrentGame(roomId, uid)
	if e != nil {
		err = e
		return
	}

	// 已经下注，直接返回
	if g.Score != 0 {
		return
	}

	g1, e := game.Games.Get(roomId)
	if e != nil {
		err = e
		return
	}
	g1.SetScore(uid, score, false)
	return
}
