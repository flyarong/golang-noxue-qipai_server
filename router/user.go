package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"net/http"
	"qipai/dao"
	"qipai/enum"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"time"
)

func user() {
	r := R.Group("/users")
	r.POST("/login", userLoginFunc)
	r.POST("/token/refresh", userTokenRefreshFunc)
	r.POST("", userRegFunc)
	r.PUT("/reset", userResetFunc)
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
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	token, err := srv.User.Login(&model.Auth{UserType: login.UserType, Name: login.Name, Pass: login.Pass})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()).Code(-1))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("登录成功").AddData("token", token))

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
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	if reg.UserType != enum.Mobile {
		c.JSON(http.StatusBadRequest, utils.Msg("目前仅支持手机注册").Code(-1))
		return
	}

	// 检查手机验证码，无论对错都删除验证码，防止暴力破解
	code := utils.Lv.Get("code_" + reg.Name)
	utils.Lv.Del("code_" + reg.Name)
	if code != reg.Code {
		c.JSON(http.StatusBadRequest, utils.Msg("手机验证码错误").Code(-1))
		return
	}

	user := model.User{Nick: reg.Nick, Ip: c.ClientIP(), Address: utils.GetAddress(c.ClientIP()),
		Auths: []model.Auth{{Name: reg.Name, UserType: reg.UserType, Pass: reg.Pass, Verified: true}}}
	user.Avatar = fmt.Sprintf("/avatar/Avatar%d.png", rand.Intn(199))
	err := srv.User.Register(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()).Code(-1))
		return
	}

	c.JSON(http.StatusOK, utils.Msg("注册成功"))
}

func userResetFunc(c *gin.Context) {

	type ResetForm struct {
		UserType enum.UserType `form:"type" json:"type" binding:"required"`
		Pass     string        `form:"pass" json:"pass" binding:"required"`
		Name     string        `form:"name" json:"name" binding:"required"`
		Code     string        `form:"code" json:"code" binding:"required"`
	}

	var reset ResetForm
	if err := c.ShouldBind(&reset); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	if reset.UserType != enum.Mobile {
		c.JSON(http.StatusBadRequest, utils.Msg("目前仅支持手机注册").Code(-1))
		return
	}

	// 检查手机验证码，无论对错都删除验证码，防止暴力破解
	code := utils.Lv.Get("code_" + reset.Name)
	utils.Lv.Del("code_" + reset.Name)
	if code != reset.Code {
		c.JSON(http.StatusBadRequest, utils.Msg("手机验证码错误").Code(-1))
		return
	}

	var a model.Auth
	dao.Db.Where(&model.Auth{UserType: reset.UserType, Name: reset.Name}).First(&a)
	if a.ID == 0 {
		c.JSON(http.StatusInternalServerError, utils.Msg("账号不存在").Code(-1))
		return
	}

	a.Pass = reset.Pass
	dao.Db.Save(&a)

	c.JSON(http.StatusOK, utils.Msg("密码修改成功"))
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
	type userV struct {
		ID        uint      `json:"id"`
		Nick      string    `gorm:"size:20" json:"nick"`
		Avatar    string    `gorm:"size:120" json:"avatar"`
		Mobile    string    `gorm:"size:20" json:"mobile"`
		Ip        string    `gorm:"size:20" json:"ip"`
		Address   string    `gorm:"size:50" json:"address"`
		Card      int       `json:"card"`
		CreatedAt time.Time `json:"created_at"`
	}
	info := c.MustGet("user").(*utils.UserInfo)
	user, err := srv.User.GetInfo(info.Uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Msg(err.Error()).Code(-1))
		return
	}
	var uv userV
	utils.Copy(user,&uv)
	c.JSON(http.StatusOK, utils.Msg("获取用户信息成功").AddData("user", uv))
}
