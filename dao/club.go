package dao

import (
	"errors"
	"github.com/golang/glog"
	"qipai/model"
)

var Club clubDao

type clubDao struct {
}

func(clubDao)Get(clubId uint)(club model.Club, err error){
	ret:=Db().First(&club,clubId)
	if ret.Error!=nil {
		err =  errors.New("查询俱乐部数据出错")
		glog.Errorln(ret.Error)
		return
	}

	if ret.RecordNotFound(){
		err = errors.New("该俱乐部不存在")
		return
	}
	return
}