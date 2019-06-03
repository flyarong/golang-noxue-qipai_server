package game

// 用户相关
const (
	NoPermission    int32 = iota + 100
	ReqLogin         // 登录请求
	ResLogin         // 响应登录结果
	ReqReg           // 用户注册
	ResReg           // 响应注册结果
	ReqReset         // 重置密码
	ResReset         // 响应重置密码结果
	ReqBind          // 账号绑定
	ResBind          // 响应绑定结果
	ReqGetUserInfo   // 获取用户信息
	ResGetUserInfo   // 响应用户信息
	ReqCode          // 请求手机验证码
	ResCode          // 返回验证码发送结果
	ReqLoginByToken  // 通过token登录
	ResLoginByToken
)

// 房间相关
const (
	// 创建房间
	ReqCreateRoom int32 = iota + 201
	ResCreateRoom
	// 房间列表
	ReqGetRoomList
	ResGetRoomList
	// 房间信息
	ReqRoom
	ResRoom
	// 进入房间
	ReqJoinRoom
	ResJoinRoom  // 此处返回当前房间的所有玩家

	// 广播进入房间
	BroadcastJoinRoom

	// 坐下
	ReqSit
	ResSit

	// 离开房间
	ReqLeaveRoom
	ResLeaveRoom
	// 解散房间
	ReqDeleteRoom
	ResDeleteRoom
)

// 游戏相关
const (
	// 开始游戏
	ReqGameStart int32 = 301
	ResGameStart int32 = 302
	// 发牌，一张一张发
	PutCard int32 = 303

	// 获取指定用户的纸牌
	ReqUserCards int32 = 304
	ResUserCards int32 = 305

	// 抢庄
	ReqTimes       int32 = 306
	BroadcastTimes int32 = 307
	// 广播谁是庄家
	BroadcastBanker int32 = 308

	// 下注
	ReqSetScore int32 = 309
	// 广播下注的大小
	BroadcastSetScore = 310
)
