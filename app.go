package main

import (
	"log"
	"qipai/router"
)

func main() {
	err := router.R.Run(":80")
	if err != nil {
		log.Panicln(err)
	}
}
