package utils

import (
	alisms "github.com/noxue/alisms"
	"log"
	"qipai/config"
)

func SendSms(tplCode, mobile, code string) bool {

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
		log.Println(msg, err)
	}
	return ok
}

func SendSmsRegCode(mobile, code string) bool {
	return SendSms("SMS_155370082", mobile, code)
}
