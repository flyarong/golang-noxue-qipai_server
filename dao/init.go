package dao

import (
	"log"
	"qipai/config"
	"qipai/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func Db() *gorm.DB {
	return db.New()
}

// 测试表，用于判断是否设置过 AUTO_INCREMENT
type test struct {
	gorm.Model
}

func init() {
	var err error
	db, err = gorm.Open("mysql", config.Config.Db.Url)
	if err != nil {
		log.Panicln(err.Error())
	}

	db = db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;")

	db.LogMode(config.Config.Debug)

	Db().AutoMigrate(
		&model.Auth{},
		&model.User{},
		&model.Room{},
		&model.Club{},
		&model.ClubRoom{},
		&model.ClubUser{},
		&model.Player{},
		&model.Event{},
		&model.Game{},
	)

	// 第一次创建，到此处还没有test表，才执行下面操作
	if !Db().HasTable(&test{}) {
		Db().Exec("alter table rooms AUTO_INCREMENT = 101010")
		Db().Exec("alter table clubs AUTO_INCREMENT = 101010")
		Db().Exec("alter table users AUTO_INCREMENT = 368751")
	}

	Db().AutoMigrate(&test{})
}
