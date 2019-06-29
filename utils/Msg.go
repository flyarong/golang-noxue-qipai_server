package utils

import (
	"encoding/json"
	"github.com/golang/glog"
	"zero"
)

type Message struct {
	data map[string]interface{} `json:"data"`
}

func Msg(msgStr string) *Message {
	m := &Message{
		data: make(map[string]interface{}),
	}
	m.data["code"] = 0
	if msgStr != "" {
		m.data["msg"] = msgStr
	}
	return m
}

func (this *Message) Code(code int) *Message {
	this.data["code"] = code
	return this
}

func (this *Message) AddData(key string, data interface{}) *Message {
	this.data[key] = data
	return this
}

func (this *Message) ToJson() string {
	jsonData, _ := json.Marshal(this.data)
	return string(jsonData)
}

func (this *Message) ToBytes() []byte {
	jsonData, _ := json.Marshal(this.data)
	return jsonData
}

func (this *Message) GetData() map[string]interface{}{
	return this.data
}

// 发送到客户端
func (this *Message) Send(msgID int32, s *zero.Session) (err error) {
	message := zero.NewMessage(msgID, this.ToBytes())
	if s == nil {
		glog.Warningln("session为nil指针，发送的消息编号为是：", msgID)
		return
	}
	err = s.GetConn().SendMessage(message)
	return
}
