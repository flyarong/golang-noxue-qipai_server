package utils

import (
	"github.com/golang/glog"
	alisms "github.com/noxue/alisms"
	"qipai/config"
)

func SendSms(tplCode, mobile, code string) bool {

	if config.Config.Debug {
		glog.Infoln("发送的验证码为：",code)
		return true
	}

	userInput := &alisms.UserParams{
		AccessKeyId:  config.Config.Sms.Key,
		AppSecret:    config.Config.Sms.Secret,
		PhoneNumbers: mobile,
		SignName:     config.Config.Sms.Sign,
		TemplateCode: tplCode,
		// 模板变量赋值，一定是json格式，注意转义
		TemplateParam: "{\"code\": \"" + code + "\"}",
	}
	ok, msg, err := alisms.SendMessage(userInput)
	if !ok {
		// 根据业务进行错误处理
		glog.Errorln(msg, err)
	}
	return ok
}

func SendSmsRegCode(mobile, code string) bool {
	return SendSms("SMS_155370082", mobile, code)
}

func SendSmsResetCode(mobile, code string) bool {
	return SendSms("SMS_155370082", mobile, code)
}

func SendSmsLoginCode(mobile, code string) bool {
	return SendSms("SMS_155370082", mobile, code)
}
