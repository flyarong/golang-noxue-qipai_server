package event

import (
	"encoding/json"
	"errors"
	"log"
	"qipai/utils"
	"strconv"
	"sync"
)

type Event struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

var evtStr string = "Event"

var evtLock sync.Mutex

var events map[string]Event

func init() {
	events = make(map[string]Event)
}

func Send(uid uint, eventName string, args ...interface{}) (err error) {
	evtLock.Lock()
	defer evtLock.Unlock()

	key := evtStr + ":" + strconv.Itoa(int(uid))

	var evts []Event

	// 获取之前的事件
	v := utils.Lv.Get(key)
	if v != "" {
		err = json.Unmarshal([]byte(v), &evts)
		if err!=nil {
			log.Println(err)
			err = errors.New("解析事件信息出错")
			return
		}
	}

	// 把事件添加到列表
	var evt Event
	evt.Name = eventName
	var as []interface{}
	for _,v:=range args{
		as = append(as, v)
	}
	evt.Args =as
	evts = append(evts, evt)

	bs, e := json.Marshal(evts)
	if e != nil {
		err = e
		return
	}
	err = utils.Lv.Put(key, string(bs))
	log.Println(string(bs))
	return
}

func Get(uid uint) (evts []Event, err error) {
	evtLock.Lock()
	defer evtLock.Unlock()

	key := evtStr + ":" + strconv.Itoa(int(uid))

	v := utils.Lv.Get(key)
	if v == "" {
		return
	}
	log.Println(v)
	err = json.Unmarshal([]byte(v), &evts)
	if err != nil {
		log.Println(err)
		err = errors.New("事件解析出错")
		return
	}

	err = utils.Lv.Del(key)
	if err != nil {
		log.Println(err)
		err = errors.New("事件删除出错")
	}

	return
}
