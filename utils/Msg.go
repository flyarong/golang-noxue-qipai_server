package utils

type msg struct {
	ErrorCode int                    `json:"code"`
	Message   string                 `json:"msg"`
	Data      map[string]interface{} `json:"data"`
}

func Msg(msgStr string) *msg {
	return &msg{
		Message: msgStr,
		Data:    map[string]interface{}{},
	}
}

func (this *msg) Code(code int) *msg {
	this.ErrorCode = code
	return this
}

func (this *msg) AddData(key string, data interface{}) *msg {
	this.Data[key] = data
	return this
}
