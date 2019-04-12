package main

import (
	"log"
	"qipai/router"
)

func main() {
	router.R.Static("/static", "./static")
	err := router.R.Run(":80")
	if err != nil {
		log.Panicln(err)
	}
}
