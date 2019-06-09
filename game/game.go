package game

import (
	"errors"
	"github.com/noxue/utils/fsm"
	"sync"
)

const (
	ReadyState                = iota + 1 // 准备中
	SelectBankerState                    // 抢庄中
	SetScoreState                        // 下注中
	ShowCardState                        // 看牌比牌中
	CompareCardState                     // 比牌中
	GameOverState                        // 结束
	GameDeletedState                     // 游戏已删除状态
)

const (
	StartAction       = iota + 1
	SetTimesAction     // 抢庄
	SetScoreAction     // 下注，下注完毕，自动把牌算好
	ShowCardAction     // 看牌
	CompareCardAction  // 比牌
	GameOverAction     // 结束游戏
)

type gamesType struct {
	Games map[uint]*Game
	lock  sync.Mutex
}

type Game struct {
	lock         sync.Mutex
	Fsm          *fsm.FSM
	RoomId       uint
	AutoPlayers  map[uint]bool // 记录玩家是否托管
	OnlinePlayer map[uint]bool // 记录玩家是否在线
}

var Games *gamesType

func init() {
	Games = &gamesType{
		Games: map[uint]*Game{},
	}
}

func (me *gamesType) NewGame(roomId uint) (game *Game, err error) {
	me.lock.Lock()
	defer me.lock.Unlock()

	game = &Game{
		Fsm:    fsm.New(StartAction),
		RoomId: roomId,
	}
	_, ok := me.Games[roomId]
	if ok {
		err = errors.New("该房间已创建游戏")
		return
	}
	// 添加不同状态调用的函数
	game.Fsm.AddState(ReadyState, StateReady)
	game.Fsm.AddState(SelectBankerState, StateSelectBanker)
	game.Fsm.AddState(SetScoreState, StateSetScore)
	game.Fsm.AddState(ShowCardState, StateShowCard)
	game.Fsm.AddState(CompareCardState, StateCompareCard)
	game.Fsm.AddState(GameOverState, StateGameOver)

	// 保存到map中统一管理
	me.Games[roomId] = game
	return
}

func (me *gamesType) Get(roomId uint)(g *Game, err error){
	me.lock.Lock()
	defer me.lock.Unlock()
	g, ok := me.Games[roomId]
	if !ok {
		err = errors.New("该房间未开始游戏")
		return
	}
	return
}

func (me *gamesType) GameOver(roomId uint) (err error) {
	me.lock.Lock()
	defer me.lock.Unlock()
	_, ok := me.Games[roomId]
	if ok {
		err = errors.New("该房间没有创建游戏，无须结束")
		return
	}
	delete(me.Games, roomId)
	return
}

func (me *Game) Start() {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(StartAction, me.RoomId)
}

func (me *Game) SetTimes(uid uint, times int, auto bool) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(SetTimesAction, me.RoomId, uid, times, auto)
}

func (me *Game) SetScore(uid uint, score int) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(SetScoreAction, me.RoomId, uid, score)
}

func (me *Game) ShowCard(uid uint) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(ShowCardAction, me.RoomId, uid)
}

// 比牌
func (me *Game) CompareCard() {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(CompareCardAction, me.RoomId)
}

// 游戏结束
func (me *Game) GameOver() {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.Fsm.Do(GameOverAction, me.RoomId)
}
