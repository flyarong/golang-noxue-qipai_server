package router

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"qipai/enum"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
)

type RegForm struct {
	UserType enum.UserType `form:"type" json:"type" binding:"required"`
	Nick     string        `form:"nick" json:"nick" binding:"required"`
	Pass     string        `form:"pass" json:"pass" binding:"required"`
	Name     string        `form:"name" json:"name" binding:"required"`
}

func user() {

	R.GET("/user/reg", func(c *gin.Context) {
		address := utils.GetAddress(c.Query("ip"))
		c.JSON(http.StatusOK, gin.H{"address": address})
	})

	R.POST("/users", func(c *gin.Context) {
		var reg RegForm
		if err := c.ShouldBind(&reg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := model.User{Nick: reg.Nick, Ip: c.ClientIP(), Address: utils.GetAddress(c.ClientIP()),
			Auths: []model.Auth{{Name: reg.Name, UserType: reg.UserType, Pass: reg.Pass, Verified: true}}}
		err := srv.User.Register(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, reg)
	})
}
