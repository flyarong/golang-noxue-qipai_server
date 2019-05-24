package model

import (
	"log"
	"sync"
	"time"
)

type online struct {
	lock sync.Mutex
	info map[uint]int64
}

func (this *online) init() {
	this.info = make(map[uint]int64)
	go func() {
		for {
			time.Sleep(time.Second * 10)
			this.lock.Lock()
			for i, v := range this.info {
				// 正常状态30秒就会更新一次 超过40秒都没有更新，表示不在线了
				if time.Now().Unix()-v > 40 {
					delete(this.info, i)
				}
				log.Println("uid:", i)
			}
			this.lock.Unlock()
		}
	}()
}

func (this *online) Get(uid uint) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.info[uid]
	return ok
}

func (this *online) SetOnline(uid uint) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.info[uid] = time.Now().Unix()
}

func (this *online) SetOffline(uid uint) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.info[uid] = 0
}
