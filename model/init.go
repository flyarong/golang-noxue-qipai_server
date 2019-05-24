package model

import (
	"github.com/jinzhu/gorm"
	"qipai/dao"
)

var Online online

// 测试表，用于判断是否设置过 AUTO_INCREMENT
type test struct {
	gorm.Model
}

func init() {
	Online.init()

	dao.Db.AutoMigrate(
		&Auth{},
		&User{},
		&Room{},
		&Club{},
		&ClubRoom{},
		&ClubUser{},
		&Player{},
		&Event{},
		&Game{},
	)

	// 第一次创建，到此处还没有test表，才执行下面操作
	if !dao.Db.HasTable(&test{}) {
		dao.Db.Exec("alter table rooms AUTO_INCREMENT = 101010")
		dao.Db.Exec("alter table clubs AUTO_INCREMENT = 101010")
		dao.Db.Exec("alter table users AUTO_INCREMENT = 100000")
	}

	dao.Db.AutoMigrate(&test{})
}
