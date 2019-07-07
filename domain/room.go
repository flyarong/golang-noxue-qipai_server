package domain

import "qipai/enum"

type ReqCreateRoom struct {
	Players   int            `json:"players"`
	Score     enum.ScoreType `json:"score"`
	Pay       enum.PayType   `json:"pay"`
	Count     int            `json:"count"`
	StartType enum.StartType `json:"start"`
	Times     int            `json:"times"`
}

type ResRoomV struct {
	ID        uint            `json:"id"`
	Score     enum.ScoreType  `json:"score"`     // 底分类型
	Pay       enum.PayType    `json:"pay"`       // 支付方式
	Current   int             `json:"current"`   // 当前第几局
	Count     int             `json:"count"`     // 总共可以玩几局
	Uid       uint            `json:"uid"`       // 房主用户编号
	StartType enum.StartType  `json:"startType"` // 游戏开始方式
	Players   int             `json:"players"`   // 玩家个数
	ClubId    uint            `json:"clubId"`    // 属于哪个俱乐部
	TableId   int             `json:"tableId"`   // 俱乐部第几桌
	Status    enum.GameStatus `json:"status"`
	Times     int             `json:"times"` // 翻倍方式
}
