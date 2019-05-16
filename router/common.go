package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"qipai/event"
	"qipai/middleware"
	"qipai/model"
	"qipai/utils"
	"regexp"
	"time"
)

func common() {
	r := R.Group("")
	ar := R.Group("")
	ar.Use(middleware.JWTAuth())

	r.GET("/code", codeFunc)

	// 获取用户事件
	ar.GET("/events", eventsFunc)

	// 发送事件
	r.POST("/events", sendEventFunc)

}

func codeFunc(c *gin.Context) {
	phone := c.Query("phone")

	reg := regexp.MustCompile(`^1[34578]\d{9}$`)
	if !reg.MatchString(phone) {
		c.JSON(http.StatusBadRequest, utils.Msg("手机号格式不正确").Code(-1))
		return
	}

	//if utils.Lv.Get("code_"+phone) != "" {
	//	c.JSON(http.StatusOK, utils.Msg().Msg("验证码已发送，请注意查收"))
	//	return
	//}

	code := rand.Intn(9000) + 1000

	//err := utils.SendSmsRegCode(phone, strconv.Itoa(code))
	//if !err {
	//	c.JSON(http.StatusBadRequest, utils.Msg().Msg("验证码已发送失败，请联系客服"))
	//	return
	//}
	utils.Lv.PutEx("code_"+phone, fmt.Sprint(code), time.Minute*5)
	log.Println("手机验证码：", code)

	c.JSON(http.StatusOK, utils.Msg("获取成功，验证码5分钟内有效"))
}

func eventsFunc(c *gin.Context) {
	info := c.MustGet("user").(*utils.UserInfo)

	// 更新用户在线状态
	model.Online.SetOnline(info.Uid)

	var evts []event.Event
	var err error
	for i := 0; i < 30; i++ {
		evts, err = event.Get(info.Uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()))
			return
		}
		if len(evts) > 0 {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	c.JSON(http.StatusOK, utils.Msg("获取事件成功").AddData("events", evts))
}

func sendEventFunc(c *gin.Context) {
	type Form struct {
		Uid   uint   `form:"uid" json:"uid" binding:"required"`
		Event string `form:"event" json:"event" binding:"required"`
		//Args  []interface{} `form:"args" json:"args"`
	}
	var form Form
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	err := event.Send(form.Uid, form.Event, 111, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("发送事件成功"))
}
