package game

import (
	"errors"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/noxue/utils/argsUtil"
	"github.com/noxue/utils/fsm"
	"math/rand"
	"qipai/config"
	"qipai/dao"
	"qipai/model"
	"qipai/utils"
	"strconv"
	"strings"
	"time"
)

var n = 0

func StateSelectBanker(action fsm.ActionType, args ...interface{}) (nextState fsm.StateType) {
	switch action {
	case SetTimesAction:
		var roomId uint
		var uid uint
		var times int
		var auto bool

		res := utils.Msg("")
		alreadySet := false // 是否已经抢庄
		defer func() {
			if alreadySet { // 已经抢庄就直接退出，不用再次通知
				return
			}
			if res == nil {
				res = utils.Msg("").AddData("game", &model.Game{
					PlayerId: uid,
					Times:    times,
					RoomId:   roomId,
					Auto:     auto,
				})
				SendToAllPlayers(res, BroadcastTimes, roomId)
				return
			}
			p := GetPlayer(uid)
			if p == nil {
				glog.V(1).Infoln("玩家：", uid, "不在线，发送抢庄信息失败")
				return
			}
			res.Send(BroadcastTimes, p.Session)
		}()
		e := argsUtil.NewArgsUtil(args...).To(&roomId, &uid, &times, &auto)
		if e != nil {
			glog.Errorln(e)
			return
		}

		room, e := dao.Room.Get(roomId)
		if e != nil {
			res = utils.Msg(e.Error()).Code(-1)
			return
		}
		// 抢庄
		ret := dao.Db().Model(&model.Game{}).Where("player_id=? and times = -1", uid).Update(map[string]interface{}{"times": times, "auto": auto})
		if ret.Error != nil {
			res = utils.Msg("更新下注信息失败").Code(-1)
			return
		}
		if ret.RowsAffected == 0 { // 如果没有更新到记录，那说明已经抢庄了
			alreadySet = true
		}
		// 判断是否都抢庄
		games, err := dao.Game.GetGames(roomId, room.Current)
		if err != nil {
			res = utils.Msg(err.Error()).Code(-1)
			return
		}

		all := true // 是否全部都抢庄了
		for _, v := range games {
			// 如果还有没抢庄的，直接返回，通知所有人该用户的抢庄倍数
			if v.Times == -1 {
				glog.V(3).Infoln(roomId, "房间：", uid, " 抢庄，", times, "倍。是否自动抢庄：", auto)
				all = false
				break
			}
		}

		// 全部抢庄完毕，选择庄家，并通知所有人
		if all {
			// 进入闲家下注状态
			nextState = SetScoreState

			bankerUid, err := selectBanker(roomId)
			if err != nil {
				res = utils.Msg(err.Error()).Code(-1)
				return
			}
			go func() {
				time.Sleep(time.Second) // 等待一秒后通知谁是庄家
				res = utils.Msg("").AddData("game", &model.Game{
					PlayerId: bankerUid,
					Banker:   true,
				})
				SendToAllPlayers(res, BroadcastBanker, roomId)
			}()

			// 闲家定时下注
			for _, v := range games {
				if v.PlayerId == bankerUid { // 庄家不用下注
					continue
				}

				go func(g model.Game) {

					g1, e := Games.Get(g.RoomId)
					if e != nil {
						glog.Error(e)
						return
					}
					auto, _ := g1.AutoPlayers[g.PlayerId]

					waitTime := time.Second * 6
					if config.Config.Debug {
						waitTime = time.Millisecond * 100;
					}
					if auto {
						waitTime = time.Second * 2
					}
					time.Sleep(waitTime)

					ss := [][]int{{1, 2}, {2, 4}, {3, 6}, {4, 8}, {5, 10}, {10, 20}}
					s := ss[room.Score]
					score := s[0]
					g1.SetScore(g.PlayerId, score, true)
				}(v)
			}
		}

		res = nil // 抢庄倍数设置成功，res设置为nil，将在defer函数中通知所有人
		return
	}
	return
}

// 选择庄家
func selectBanker(roomId uint) (uid uint, err error) {
	games, e := dao.Game.GetCurrentGames(roomId)
	if e != nil {
		err = e
		return
	}
	if len(games) == 0 {
		err = errors.New("当前房间没有玩家")
		return
	}

	eq := true // 记录是否全部相等
	var game = games[0]
	var game4 []model.Game
	var sGame model.Game // 特殊用户
	// 选择下注最大的
	for _, g := range games {
		if g.Times == 4 {
			if dao.User.IsSpecialUser(g.PlayerId) {
				sGame = g
				break
			}
			game4 = append(game4, g)
		}
		if g.Times != game.Times {
			eq = false
			if g.Times > game.Times {
				game = g
			}
		}
	}

	// 如果有多个4倍，就随机选一个
	rand.Seed(time.Now().Unix())
	if sGame.ID > 0 { // 如果有特殊用户，那么设置他为庄
		game = sGame
	} else if eq {
		game = games[rand.Intn(len(games))]
	} else if len(game4) > 2 {
		game = game4[rand.Intn(len(game4))]
	}

	uid = game.PlayerId
	// 更新
	var res *gorm.DB
	if sGame.ID > 0 { // 特殊用户随机生成牛七到牛牛的牌，并且一定抢到庄
		// 计算要排除的牌
		var removeCards []int
		for _, g := range games {
			if g.ID == sGame.ID {
				continue
			}
			css := strings.Split(g.Cards, "|")
			for _, g1 := range css {
				v, err := strconv.Atoi(g1)
				if err != nil {
					glog.Error("转换牌的点数失败", err)
				}
				removeCards = append(removeCards, v)
			}
		}
		cards := ""
		result := createCard(removeCards, rand.Intn(4)+7)
		for _, v := range result {
			cards += strconv.Itoa(v) + "|"
		}
		res = dao.Db().Model(&game).Updates(model.Game{Banker: true, Cards: cards[:len(cards)-1]})
	} else {
		res = dao.Db().Model(&game).Update("banker", true)
	}
	if res.Error != nil {
		err = errors.New("选定庄家出错")
		return
	}
	if res.RowsAffected == 0 {
		err = errors.New("更新庄家信息出错")
		return
	}


	if game.Current > 1 {
		err = sendTuiZhu(roomId, game.PlayerId)
	}
	return
}

func sendTuiZhu(roomId, bankerUid uint) (err error) {
	games, e := dao.Game.GetLastGames(roomId)
	if e != nil {
		err = e
		return
	}

	type tuiUser struct {
		Uid    uint `json:"uid"`    // 用户 id
		DeskId int  `json:"deskId"` // 座位号
		Score  int  `json:"score"`  //推注积分
	}

	var tuiUsers []tuiUser
	for _, v := range games {
		// 上一把是庄家 或者 输了 或者 当前把是庄  或者 上把推注 就不能再推注
		if v.Banker || v.TotalScore <= 0 || v.PlayerId == bankerUid || v.Tui {
			continue
		}
		tuiUsers = append(tuiUsers, tuiUser{
			Uid:    v.PlayerId,
			DeskId: v.DeskId,
			Score:  v.Score + v.TotalScore,
		})

		ret:=dao.Db().Model(model.Game{}).Where(model.Game{RoomId: roomId, PlayerId: v.PlayerId, Current: v.Current + 1}).Update(model.Game{Tui: true})
		if ret.Error!=nil{
			glog.Error(ret.Error)
		}
	}
	msg := utils.Msg("").AddData("roomId", roomId).AddData("users", tuiUsers)
	SendToAllPlayers(msg, BroadcastAllScore, roomId)
	return
}

// 生成指定牛牌
// cards 要排除的牌
// n 要生成的牛的点数，如牛七 n 就等于 7
func createCard(removeCards []int, n int) (result []int) {
	var cards [10][]int

	for i := 0; i < 52; i++ {
		// 排除已经发过的牌
		ok := func() bool {
			for _, v := range removeCards {
				if v == i {
					return true
				}
			}
			return false
		}()
		if ok {
			continue
		}
		c := i % 13
		if c > 9 {
			c = 9
		}
		cards[c] = append(cards[c], i)
	}

	// 选1张10点的牌
	var card int
	card, cards = randCard(cards, 10)
	result = append(result, card)

	// 要么3个10，要么其中两个牌组成10
	if rand.Intn(2) > 0 {
		card, cards = randCard(cards, 10)
		result = append(result, card)

		card, cards = randCard(cards, 10)
		result = append(result, card)
	} else {
	ReCreate:
		// 生成第一张牌
		n1 := rand.Intn(9)
		n2 := 9 - n1 - 1
		// 如果指定点数的牌是空的，那就重新生成
		if len(cards[n1]) == 0 || len(cards[n2]) == 0 {
			goto ReCreate
		}

		card, cards = randCard(cards, n1+1)
		result = append(result, card)

		card, cards = randCard(cards, n2+1)
		result = append(result, card)
	}

	// 生成两张牌牛牌
ReCreate2:
	// 生成第一张牌
	n1 := rand.Intn(n - 1)
	n2 := n - n1 - 2
	// 如果指定点数的牌是空的，那就重新生成
	if len(cards[n1]) == 0 || len(cards[n2]) == 0 {
		goto ReCreate2
	}

	card, cards = randCard(cards, n1+1)
	result = append(result, card)

	card, cards = randCard(cards, n2+1)
	result = append(result, card)

	return
}

func randCard(cards [10][]int, n int) (card int, returnCards [10][]int) {
	rand.Seed(time.Now().UnixNano())
	c := rand.Intn(len(cards[n-1]))
	card = cards[n-1][c]
	cards[n-1] = append(cards[n-1][:c], cards[n-1][c+1:]...)
	returnCards = cards
	return
}
