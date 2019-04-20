package model

import (
	"github.com/jinzhu/gorm"
	"qipai/enum"
)

type Club struct {
	Name   string // 俱乐部名称
	Check  bool   // 是否审查
	Notice string // 公告
	Room
}

type Room struct {
	gorm.Model
	Score     enum.ScoreType  // 底分 以竖线分割的底分方式 如 10|20
	Players   int             // 玩家个数
	Count     int             // 局数
	Current   int             // 当前第几局
	StartType enum.StartType  // 游戏开始方式
	Pay       enum.PayType    // 付款方式 0 房主或俱乐部老板付 1 AA
	Times     enum.TimesType  // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
	Special   int             // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
	King      enum.KingType   // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
	Uid       uint            // 老板
	Status    enum.GameStatus // 0 未开始，1 游戏中， 2 中场休息， 3 已结束
}

// 记录俱乐部的用户
type ClubUser struct {
	gorm.Model
	Uid    uint              // 用户编号
	ClubId uint              // 俱乐部编号
	Status enum.ClubUserType // 0 等待审核，1 正式用户， 2 冻结用户
	Payer  bool              // 是否是代付 true 是代付
	Admin  bool              // 是否是管理员 true 是管理员
}

// 每局游戏数据
type Game struct {
	gorm.Model
	RoomId  uint
	Count   int    // 第几局
	Cards   string // 已竖线分割的牌型字符串
	Order   int    // 座位编号
	Banker  bool   // 是否是庄家 true表示是庄家
	Times   int    // 倍数
	Base    int    // 底分
	Special int    // 特殊牌型加倍
	Score   int    // 输赢积分，通过底分*庄家倍数*特殊牌型加倍 计算
}
