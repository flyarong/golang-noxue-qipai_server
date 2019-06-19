package game

import (
	"github.com/golang/glog"
	"qipai/dao"
	"zero"
)

var players map[uint]*Player

func init() {
	players = make(map[uint]*Player)

	// 生成存储 玩家状态的表
	dao.Db().AutoMigrate(&Player{})
}

type Player struct {
	Id      uint          `json:"-" gorm:"primary_key"`
	Uid     uint           `json:"id" sql:"index"`
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
func RemovePlayer(id uint) {
	if _, ok := players[id]; ok {
		delete(players, id)
	}
}

// 获取保持网络通讯的玩家信息
func GetPlayer(uid uint) *Player {
	if v, ok := players[uid]; ok {
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
func GetPlayerFromSession(s *zero.Session) (player *Player, err error) {
	var ok bool
	player, ok = s.GetSetting("user").(*Player)
	if !ok {
		glog.Errorln("在线用户信息转换失败")
		return
	}
	return
}

func IsLogin(s *zero.Session) bool {
	_, ok := s.GetSetting("user").(*Player)
	return ok
}
