package main

import (
	"flag"
	"qipai/controller"
	"qipai/game"
	_ "qipai/router"
	"time"
	"zero"

	"github.com/golang/glog"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "d", false, "是否开启调试")
	flag.Parse()
}

func main() {
	go func() {
		err := controller.R.Run(":9988")
		if err != nil {
			glog.Fatalln(err)
		}
	}()

	host := ":8899"
	ss, err := zero.NewSocketService(host)
	if err != nil {
		glog.Fatal(err)
	}

	ss.SetHeartBeat(5*time.Second, 120*time.Second)

	ss.RegMessageHandler(game.HandleMessage)
	ss.RegConnectHandler(game.HandleConnect)
	ss.RegDisconnectHandler(game.HandleDisconnect)

	glog.Infoln("服务器启动成功,监听地址" + host)

	ss.Serv()
}
