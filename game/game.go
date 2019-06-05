package game

import (
	"qipai/dao"
	"zero"
)

var players map[int]*Player

func init() {
	players = make(map[int]*Player)

	// 生成存储 玩家状态的表
	dao.Db().AutoMigrate(&Player{})
}

type Player struct {
	Id      uint          `json:"-" gorm:"primary_key"`
	Uid     int           `json:"id" sql:"index"`
	Nick    string        `json:"nick"`
	Session *zero.Session `json:"-" gorm:"-"`
}

func (Player) TableName() string {
	return "offline_players"
}

// 有玩家加入
func AddPlayer(s *zero.Session, p *Player) {

	var player Player
	res := dao.Db().Where("uid=?", p.Uid).First(&player)
	// 如果找到，表示只是掉线了，重新关联上之前的player
	if res.Error == nil && !res.RecordNotFound() {
		dao.Db().Delete(Player{}, "uid=?", p.Uid) // 用户重新上线，就从离线表中删除他
		s.SetSetting("user", &player)
		player.Session = s
		p = &player
	} else {
		p.Session = s
		p.Session.SetSetting("user", p)
	}
	players[p.Uid] = p
}

// 移除玩家
func RemovePlayer(id int) {
	if _, ok := players[id]; ok {
		delete(players, id)
	}
}

// 获取玩家
func GetPlayer(id int) *Player {
	if v, ok := players[id]; ok {
		return v
	}
	return nil
}

// 获取全部玩家
func GetPlayerList() []*Player {
	list := make([]*Player, 0)
	for _, p := range players {
		list = append(list, p)
	}
	return list
}

// 从session中获取玩家
func GetPlayerFromSession(s *zero.Session) *Player {
	p, _ := s.GetSetting("user").(*Player)
	return p
}

func IsLogin(s *zero.Session) bool {
	_, ok := s.GetSetting("user").(*Player)
	return ok
}
