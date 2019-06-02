package event

import "testing"

func TestSend(t *testing.T) {
	Send(101010,"UserJoinRoom",101012,101013,101014)
}

func TestGet(t *testing.T) {
	Get(101010)
}

func TestDoArg(t *testing.T) {
	str:=doArgs([]interface{}{1,"sss",3,2.3})
	print(str)
}