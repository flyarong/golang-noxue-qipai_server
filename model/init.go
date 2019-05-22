package model

import "qipai/dao"

var Online online

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
	)

	dao.Db.Exec("alter table rooms AUTO_INCREMENT = 101010")
	dao.Db.Exec("alter table clubs AUTO_INCREMENT = 101010")
	dao.Db.Exec("alter table users AUTO_INCREMENT = 100000")
}
