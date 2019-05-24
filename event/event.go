package event

import (
	"fmt"
	"qipai/dao"
	"qipai/model"
	"strings"
)

type EventType string

const (
	RoomJoin   EventType = "RoomJoin"
	RoomExit   EventType = "RoomExit"
	RoomList   EventType = "RoomList"
	RoomCreate EventType = "RoomCreate"
	RoomDelete EventType = "RoomDelete"
	RoomStart  EventType = "RoomStart"
	GameBegin  EventType = "GameBegin"
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
		e.Args += fmt.Sprintf("\t%v", a)
	}
	dao.Db.Save(&e)
	return
}

func Get(uid uint) (events []Event, ok bool) {
	var e model.Event
	dao.Db.Where("uid=?", uid).First(&e)
	if e.ID == 0 {
		return
	}
	var ev Event
	ev.Name = EventType(e.Name)
	if len(e.Args)>0{
		ev.Args = strings.Split(e.Args, "\t")[1:]
	}
	events = append(events, ev)

	dao.Db.Delete(&model.Event{}, e.ID)
	ok = true
	return
}
