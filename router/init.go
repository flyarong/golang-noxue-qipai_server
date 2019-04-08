package router

import (
	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func init() {
	Router = gin.Default()
	user()
}
