package controller

import (
	"github.com/gin-gonic/gin"
)

var R *gin.Engine

func init() {
	R = gin.Default()
	R.Static("/static","./static")
	user()
}

