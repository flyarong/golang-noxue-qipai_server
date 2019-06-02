package event

import (
	"errors"
	"fmt"
	"qipai/dao"
	"qipai/model"
	"reflect"
	"strings"
)

type EventType string

const (
	RoomJoin      EventType = "RoomJoin"
	PlayerSitDown EventType = "PlayerSitDown"
	RoomExit      EventType = "RoomExit"
	RoomList      EventType = "RoomList"
	RoomCreate    EventType = "RoomCreate"
	RoomDelete    EventType = "RoomDelete"
	RoomStart     EventType = "RoomStart"
	GameBegin     EventType = "GameBegin"
)

type Event struct {
	Name EventType   `json:"name"`
	Args interface{} `json:"args"`
}

func Send(uid uint, eventName EventType, args ...interface{}) (err error) {

	var e model.Event
	e.Uid = uid
	e.Name = string(eventName)
	for _, a := range args {
		e.Args += doArgs(a)
	}
	if dao.Db().Save(&e).Error != nil {
		err = errors.New("添加事件失败")
	}
	return
}

func doArgs(arg interface{}) (string) {
	kind := reflect.TypeOf(arg).Kind()
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array {
		str := ""
		for _, v := range arg.([]interface{}) {
			str += fmt.Sprintf("\t%v", v)
		}
		return str;
	}
	return fmt.Sprintf("\t%v", arg)
}

func Get(uid uint) (events []Event, ok bool) {
	var e model.Event
	dao.Db().Where("uid=?", uid).First(&e)
	if e.ID == 0 {
		return
	}
	var ev Event
	ev.Name = EventType(e.Name)
	if len(e.Args) > 0 {
		ev.Args = strings.Split(e.Args, "\t")[1:]
	}
	events = append(events, ev)

	dao.Db().Delete(&model.Event{}, e.ID)
	ok = true
	return
}
