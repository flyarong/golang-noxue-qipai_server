package enum

type UserType int

const (
	Mobile UserType = 1
	WeChat UserType = 2
)

// 王癞类型
type KingType int

const (
	KingNone KingType = 0 // 无王癞
	KingLast KingType = 1 // 经典王癞（王牌只会出现在最后一张）
	KingAll  KingType = 2 // 疯狂王癞(王牌会出现在任何一张)
)

