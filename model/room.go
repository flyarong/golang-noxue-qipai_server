package model

import (
	"github.com/jinzhu/gorm"
	"qipai/enum"
	"time"
)

type Club struct {
	Name     string // 俱乐部名称
	Check    bool   // 是否审查
	Notice   string // 公告
	RollText string // 俱乐部大厅滚动文字
	Close    bool   // 是否打烊
	PayerUid uint   // 代付用户id
	ClubRoomBase
}

// 提取club和room共有的字段
type ClubRoomBase struct {
	gorm.Model
	Score     enum.ScoreType // 底分 以竖线分割的底分方式
	Players   int            // 玩家个数
	Count     int            // 总局数
	StartType enum.StartType // 游戏开始方式
	Pay       enum.PayType   // 付款方式 0 俱乐部老板付 1 AA
	Times     enum.TimesType // 翻倍规则，预先固定的几个选择，比如：牛牛x3  牛九x2
	Special   int            // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
	King      enum.KingType  // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
	Uid       uint           // 老板
}

type Room struct {
	ClubRoomBase
	Current int             // 当前第几局
	Status  enum.GameStatus // 0 未开始，1 游戏中， 2 已结束
	ClubId  uint            // 属于哪个俱乐部
}

// 俱乐部和房间的关系
type ClubRoom struct {
	Cid uint // 俱乐部编号
	Rid uint // 房间编号
}

// 记录俱乐部的用户
type ClubUser struct {
	gorm.Model
	Uid    uint              // 用户编号
	ClubId uint              // 俱乐部编号
	Status enum.ClubUserType // 0 等待审核，1 正式用户， 2 冻结用户
	Admin  bool              // 是否是管理员 true 是管理员
}

// 记录房间中的用户
type Player struct {
	gorm.Model
	Uid      uint                     // 用户编号
	Nick     string                   // 昵称
	DeskId   int                      // 座位号
	RoomId   uint                     // 房间编号
	JoinedAt *time.Time `sql:"index"` // 加入时间
}

type Game struct {
	gorm.Model
	Banker   bool   // 是否是庄家 true表示是庄家
	PlayerId uint   // 玩家编号
	RoomId   uint   // 房间编号
	DeskId   int    // 座位号
	Times    int    // 下注倍数
	Special  int    // 特殊牌型加倍
	Score    int    // 输赢积分，通过底分*庄家倍数*特殊牌型加倍 计算
	Cards    string // 用户所拥有的牌
	Current  int    // 这是第几局
	Auto     bool   // 是否自动托管
}

type Event struct {
	gorm.Model
	Uid  uint `sql:"index"`
	Name string
	Args string
}
