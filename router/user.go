package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"qipai/dao"
	"qipai/model"
	"qipai/utils"
)

func user() {

	Router.GET("/user/reg/:name", func(c *gin.Context) {
		user := model.User{Name: "admin", Ip: "127.0.0.222", Address: "中国贵州", Mobile: "13758277505"}
		fmt.Println("dao.Db",dao.Db)
		dao.Db.Create(&user)

		token,_:=utils.NewToken(1,c.Param("name"))
		c.JSON(200, token)
	})
}
