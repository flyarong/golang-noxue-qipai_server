package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qipai/enum"
	"qipai/game"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
)

func user(){
	r := R.Group("/users")
	r.POST("/login", userLoginFunc)
	ar := r.Group("")
	ar.Use(middleware.JWTAuth())
	ar.POST("/notice", postNoticeFunc)
	ar.POST("/rollText", postRollText)
	ar.POST("/shareText", postShareText)
}

func postRollText(c *gin.Context) {
	type ReqForm struct {
		RollText     string        `form:"rollText" json:"rollText" binding:"required"`
	}
	var form ReqForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
		return
	}
	err :=utils.Lv.Put("user_rollText", form.RollText)
	if err!=nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
	}
	players := game.GetPlayerList()
	for _,v:=range players{
		utils.Msg("").AddData("rollText", form.RollText).Send(game.ResRollText,v.Session)
	}
	c.JSON(http.StatusOK, utils.Msg("发布滚动字幕成功").GetData())
}

func postShareText(c *gin.Context) {
	type ReqForm struct {
		ShareText     string        `form:"shareText" json:"shareText" binding:"required"`
	}
	var form ReqForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
		return
	}
	err :=utils.Lv.Put("user_shareText", form.ShareText)
	if err!=nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
	}
	c.JSON(http.StatusOK, utils.Msg("更新分享内容成功").GetData())
}

func postNoticeFunc(c *gin.Context) {
	type ReqForm struct {
		Notice     string        `form:"notice" json:"notice" binding:"required"`
	}
	var form ReqForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
		return
	}
	err :=utils.Lv.Put("user_notice", form.Notice)
	if err!=nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
	}
	players := game.GetPlayerList()
	for _,v:=range players{
		utils.Msg("").AddData("notice", form.Notice).Send(game.ResNotice,v.Session)
	}
	c.JSON(http.StatusOK, utils.Msg("发布通知成功").GetData())
}

func userLoginFunc(c *gin.Context) {
	type LoginForm struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
	}

	var login LoginForm
	if err := c.ShouldBind(&login); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1).GetData())
		return
	}

	token,user, err := srv.User.Login(&model.Auth{UserType: login.UserType, Name: login.Name, Pass: login.Pass})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()).Code(-1).GetData())
		return
	}
	c.JSON(http.StatusOK, utils.Msg("登录成功").AddData("token", token).AddData("user",user).GetData())
}

