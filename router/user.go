package router

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"qipai/enum"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
)

func user() {
	r := R.Group("/users")
	r.POST("/login", userLoginFunc)
	r.POST("/token/refresh", userTokenRefreshFunc)
	r.POST("", userRegFunc)

	ar := r.Group("")
	ar.Use(middleware.JWTAuth())
	ar.POST("/bind", userBindFunc)
	ar.GET("", userInfoFunc)

}

func userLoginFunc(c *gin.Context) {
	type LoginForm struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
	}

	var login LoginForm
	if err := c.ShouldBind(&login); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	token, err := srv.User.Login(&model.Auth{UserType: login.UserType, Name: login.Name, Pass: login.Pass})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.Msg().Msg("登录成功").AddData("token", token))

}

func userTokenRefreshFunc(c *gin.Context) {
	token, err := srv.User.TokenRefresh(c.GetHeader("token"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func userRegFunc(c *gin.Context) {

	type RegForm struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Nick     string        `form:"nick" json:"nick" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
		Code     string        `form:"code" json:"code" binding:"required"`
	}

	var reg RegForm
	if err := c.ShouldBind(&reg); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	if reg.UserType != enum.Mobile {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("目前仅支持手机注册"))
		return
	}

	// 检查手机验证码，无论对错都删除验证码，防止暴力破解
	code :=utils.Lv.Get("code_"+reg.Name)
	utils.Lv.Del("code_"+reg.Name)
	if code != reg.Code {
		c.JSON(http.StatusBadRequest, utils.Msg().Code(-1).Msg("手机验证码错误"))
		return
	}

	user := model.User{Nick: reg.Nick, Ip: c.ClientIP(), Address: utils.GetAddress(c.ClientIP()),
		Auths: []model.Auth{{Name: reg.Name, UserType: reg.UserType, Pass: reg.Pass, Verified: true}}}
	err := srv.User.Register(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg().Code(-1).Msg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.Msg().Msg("注册成功"))
}

func userBindFunc(c *gin.Context) {
	type RegForm struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
	}

	user := c.MustGet("user").(*utils.UserInfo)

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func userInfoFunc(c *gin.Context) {
	info := c.MustGet("user").(*utils.UserInfo)
	user, err := srv.User.GetInfo(info.Uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
