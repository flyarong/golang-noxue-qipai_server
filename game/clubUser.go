package game

import (
	"github.com/golang/glog"
	"sync"
	"zero"
)

// map[clubId][uid]*clubPlayer
type clubPlayers map[uint]map[uint]*clubPlayer

var lock sync.Mutex
var ClubPlayers clubPlayers

type clubPlayer struct {
	Session *zero.Session
}

func init() {
	ClubPlayers = make(map[uint]map[uint]*clubPlayer)
}

// 对指定clubId的用户做指定的操作，用于通知同一个俱乐部中的在线用户
func (me clubPlayers)Do(clubId uint,callback func(s *zero.Session)){
	lock.Lock()
	defer lock.Unlock()
	for k,v:=range me{
		if k!=clubId{
			continue
		}
		for _,v1:=range v{
			callback(v1.Session)
		}
		break
	}
}

func (me clubPlayers) Add(clubId, uid uint, session *zero.Session) {
	lock.Lock()
	defer lock.Unlock()

	glog.V(3).Infoln(uid, "进入", clubId, "茶楼")

	_, ok := me[clubId]
	if !ok {
		me[clubId] = make(map[uint]*clubPlayer)
	}
	if _, ok := me[clubId][uid]; ok {
		me[clubId][uid].Session = session
		return
	}
	me[clubId][uid] = &clubPlayer{
		Session: session,
	}
}

func (me clubPlayers) Del(uid uint) {
	lock.Lock()
	defer lock.Unlock()
	for k, v := range ClubPlayers {
		for k1, _ := range v {
			if k1 == uid {
				delete(ClubPlayers[k], k1)
				glog.V(3).Infoln(uid, "离开", k, "茶楼")
				break
			}
		}
	}
}
