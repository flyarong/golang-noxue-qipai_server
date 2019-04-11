package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"net/http"
	"qipai/utils"
	"regexp"
	"strconv"
	"time"
)

func common() {
	r := R.Group("")
	r.GET("/code", codeFunc)
}

func codeFunc(c *gin.Context) {
	phone := c.Query("phone")

	reg := regexp.MustCompile(`^1[34578]\d{9}$`)
	if !reg.MatchString(phone) {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("手机号格式不正确"))
		return
	}

	//if utils.Lv.Get("code_"+phone) != "" {
	//	c.JSON(http.StatusOK, utils.Msg().Msg("验证码已发送，请注意查收"))
	//	return
	//}

	code := rand.Intn(9000) + 1000

	err := utils.SendSmsRegCode(phone, strconv.Itoa(code))
	if !err {
		c.JSON(http.StatusBadRequest, utils.Msg().Msg("验证码已发送失败，请联系客服"))
		return
	}
	utils.Lv.PutEx("code_"+phone, fmt.Sprint(code), time.Minute*5)
	log.Println("手机验证码：", code)

	c.JSON(http.StatusOK, utils.Msg().Msg("获取成功，验证码5分钟内有效"))
}
