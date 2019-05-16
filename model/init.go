package model

import "qipai/dao"

var Online online

func init() {
	Online.init()

	//defer dao.Db.Close()
	dao.Db.AutoMigrate(
		&Auth{},
		&User{},
		&Room{},
		&Club{},
		&ClubRoom{},
		&ClubUser{},
		&Game{},
		&Player{},
	)
	dao.Db.Exec("alter table rooms AUTO_INCREMENT = 101010");
	dao.Db.Exec("alter table clubs AUTO_INCREMENT = 101010");
	dao.Db.Exec("alter table users AUTO_INCREMENT = 100000");
}
