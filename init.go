package main

import (
	"math/rand"
	"time"
)

func init() {
	// 指定随机数种子
	rand.Seed(time.Now().Unix())
}