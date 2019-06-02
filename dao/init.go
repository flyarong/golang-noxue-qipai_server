package dao

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"log"
	"qipai/config"
)

var db *gorm.DB

func Db() *gorm.DB{
	return db.New()
}

func init() {
	var err error
	db, err = gorm.Open("mysql", config.Config.Db.Url)
	if err != nil {
		log.Panicln(err.Error())
	}

	db = db.Set("gorm:table_options", "ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;")
	//Db.LogMode(true)
}
