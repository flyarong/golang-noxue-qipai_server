package main

import (
	"log"
	"qipai/router"
)

func main() {
	err := router.Router.Run(":80")
	if err != nil {
		log.Panicln(err)
	}
}
