package utils

import (
	"encoding/json"
	"zero"
)

type msg struct {
	data map[string]interface{} `json:"data"`
}

func Msg(msgStr string) *msg {
	m := &msg{
		data: make(map[string]interface{}),
	}
	m.data["code"] = 0
	if msgStr != "" {
		m.data["msg"] = msgStr
	}
	return m
}

func (this *msg) Code(code int) *msg {
	this.data["code"] = code
	return this
}

func (this *msg) AddData(key string, data interface{}) *msg {
	this.data[key] = data
	return this
}

func (this *msg) ToJson() string {
	jsonData, _ := json.Marshal(this.data)
	return string(jsonData)
}


func (this *msg) ToBytes() []byte {
	jsonData, _ := json.Marshal(this.data)
	return jsonData
}

// 发送到客户端
func (this *msg) ToSend(msgID int32,s *zero.Session)(err error){
	message := zero.NewMessage(msgID, this.ToBytes())
	err = s.GetConn().SendMessage(message)
	return
}
