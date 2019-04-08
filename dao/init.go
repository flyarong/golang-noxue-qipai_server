package dao

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"qipai/config"
	"qipai/model"
)

var Db *gorm.DB

func init() {
	fmt.Println("config.Config.Db.Url", config.Config.Db.Url)
	var err error
	Db, err = gorm.Open("mysql", config.Config.Db.Url)
	if err != nil {
		log.Panicln(err.Error())
	}

	Db = Db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;")
	Db.LogMode(true)
	//defer Db.Close()
	Db.AutoMigrate(&model.Auth{}, &model.User{})

}
