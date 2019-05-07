package srv

import (
	"errors"
	"qipai/dao"
	"qipai/model"
)

var Room roomSrv

type roomSrv struct {
}

func (roomSrv) Create(room *model.Room)(err error){
	dao.Db.Save(room)
	if room.ID == 0 {
		err = errors.New("房间添加失败，请联系管理员")
		return
	}
	return
}

func (roomSrv) MyRooms(uid uint)(rooms []model.Room){
	// 我创建的
	where:=&model.Room{}
	where.Uid = uid
	dao.Db.Where(where).Find(&rooms)
	return
}