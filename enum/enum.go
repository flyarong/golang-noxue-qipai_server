package enum

type UserType int

const (
	MobilePass UserType = 1 // 手机 密码登录
	MobileCode UserType = 2 // 手机 验证码登录
	WeChat     UserType = 3 // 微信登录
)

// 王癞类型
type KingType int

const (
	KingNone KingType = 0 // 无王癞
	KingLast KingType = 1 // 经典王癞（王牌只会出现在最后一张）
	KingAll  KingType = 2 // 疯狂王癞(王牌会出现在任何一张)
)

// 游戏开始方式
type StartType int

const (
	StartBoss  StartType = 0 // 房主开始
	StartFirst StartType = 1 // 首位开始
)

// 付款方式
type PayType int

const (
	PayBoss PayType = 0
	PayAA   PayType = 1
)

// 用户游戏状态
type GameStatus int

const (
	GameReady   GameStatus = 0
	GamePlaying GameStatus = 1
	GameOver    GameStatus = 2
)

// 俱乐部用户类型
type ClubUserType int

const (
	ClubUserWait    ClubUserType = 0 // 等待加入
	ClubUserVip     ClubUserType = 1 // 正式用户
	ClubUserDisable ClubUserType = 2 // 冻结的用户
)

// 翻倍规则
type TimesType int

const (
	TimesType1 TimesType = 0 // 牛一~牛牛 分别对应 1~10倍
	TimesType2 TimesType = 1 // 牛牛x5 牛九x4 牛八x3 牛七x2
	TimesType3 TimesType = 2 // 牛牛x3 牛九x2 牛八x2 牛七x2
	TimesType4 TimesType = 3 // 牛牛x3 牛九x2牛八x2 牛七x1
	TimesType5 TimesType = 4 // 牛牛x4 牛九x3 牛八x2 牛七x2
)

// 底分类型
type ScoreType int

const (
	ScoreType1 ScoreType = 0
	ScoreType2 ScoreType = 1
	ScoreType3 ScoreType = 2
	ScoreType4 ScoreType = 3
	ScoreType5 ScoreType = 4
	ScoreType6 ScoreType = 5
)
