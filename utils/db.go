package utils

import (
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"qipai/config"
	"time"
)

var Lv *lv

type lv struct {
	db *leveldb.DB
}

func initLv() () {
	db, err := leveldb.OpenFile(config.Config.Lvdb, nil)
	if err != nil {
		log.Panic(err.Error())
	}
	Lv = &lv{
		db: db,
	}
}

func (this *lv) Get(key string) string {
	data, _ := this.db.Get([]byte(key), nil)
	return string(data)
}

func (this *lv) Put(key, value string) {
	this.db.Put([]byte(key), []byte(value), nil)
}

func (this *lv) PutEx(key, value string, expired time.Duration) {
	this.Put(key, value)
	go func() {
		time.Sleep(expired)
		this.Del(key)
	}()
}

func (this *lv) Del(key string) {
	this.db.Delete([]byte(key), nil)
}

func (this *lv) Close() {
	this.db.Close()
}
