package srv

func init() {
	initUser()
	// 开线程定时删除应该解散的房间
	deleteAllInvalidRooms()
}

