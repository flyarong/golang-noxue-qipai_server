package game

import (
	"github.com/golang/glog"
	"log"
	"qipai/dao"
	"qipai/utils"
	"zero"
)

type handler func(s *zero.Session, msg *zero.Message)

type handlerWrap struct {
	needAuth bool // 记录 是否需要授权后才能执行handler
	handler  handler
}

// 保存所有消息的处理函数
var handlers map[int32]handlerWrap = make(map[int32]handlerWrap)

// 添加消息处理函数
func AddHandler(msgID int32, handler handler) {
	// 如果已存在，直接退出
	if _, ok := handlers[msgID]; ok {
		glog.Fatalln(msgID, "消息处理函数已存在，请勿重复添加")
	}
	handlers[msgID] = handlerWrap{handler: handler}
}

// 添加需要授权后才可以访问的处理函数
func AddAuthHandler(msgID int32, handler handler) {
	// 如果已存在，直接退出
	if _, ok := handlers[msgID]; ok {
		glog.Fatalln(msgID, "消息处理函数已存在，请勿重复添加")
	}
	handlers[msgID] = handlerWrap{handler: handler, needAuth: true}
}

func HandleMessage(s *zero.Session, msg *zero.Message) {

	msgID := msg.GetID()

	handler, ok := handlers[msgID]
	if !ok {
		glog.Warning("存在未处理的消息，编号为：", msgID)
		return
	}
	// 需要登录 并且没登录，就提示错误
	if handler.needAuth && !IsLogin(s) {
		utils.Msg("").Code(-1).Send(NoPermission, s)
		return
	}
	handler.handler(s, msg)
}

// HandleDisconnect 处理网络断线
func HandleDisconnect(s *zero.Session, err error) {
	glog.V(2).Infoln(s.GetConn().GetName() + " 掉线")
	// 如果玩家已登录，保存掉线玩家
	if IsLogin(s){
		p,e:=GetPlayerFromSession(s)
		if e!=nil{
			glog.Errorln(e)
			err = e
			return
		}
		RemovePlayer(p.Uid)
		dao.Db().Save(p)
	}
}

// HandleConnect 处理网络连接
func HandleConnect(s *zero.Session) {
	log.Println(s.GetConn().GetName() + " 连接")
}
