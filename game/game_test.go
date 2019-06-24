package game

import (
	"testing"
)

func a() int {
	return 1
}

func b() int {
	return 2
}


func TestDoArgs(t *testing.T) {
	for{
		println(a(),b())
	}
}
