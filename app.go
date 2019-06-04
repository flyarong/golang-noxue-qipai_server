package main

import (
	"flag"
	"github.com/golang/glog"
	"qipai/game"
	_ "qipai/router"
	"zero"
)

var debug bool

func init() {
	flag.BoolVar(&debug, "d", false, "是否开启调试")
	flag.Parse()
}

func main() {

	host := ":8899"
	ss, err := zero.NewSocketService(host)
	if err != nil {
		glog.Fatal(err)
	}

	//ss.SetHeartBeat(5*time.Second, 10*time.Second)

	ss.RegMessageHandler(game.HandleMessage)
	ss.RegConnectHandler(game.HandleConnect)
	ss.RegDisconnectHandler(game.HandleDisconnect)

	glog.Infoln("服务器启动成功,监听地址" + host)

	ss.Serv()
}
