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
	ReqUserInfo      // 获取用户信息
	ResUserInfo      // 响应用户信息
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
	ReqRoomList
	ResRoomList
	// 房间信息
	ReqRoom
	ResRoom
	// 进入房间
	ReqJoinRoom
	ResJoinRoom  // 此处返回当前房间的所有玩家

	// 广播进入房间
	BroadcastJoinRoom
	// 广播坐下
	BroadcastSitRoom

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
	ReqGameStart int32 = iota + 301
	ResGameStart
	ReqSetTimes  // 抢庄
	BroadcastTimes
	BroadcastBanker       // 广播谁是庄家
	ReqSetScore           // 下注
	BroadcastScore        // 广播下注的大小
	BroadcastShowCard     // 广播看牌，下注完毕，可以看自己的牌
	BroadcastCompareCard  // 比牌，返回所有人牌型及大小输赢积分，前端展示比牌结果
	BroadcastGameOver     // 游戏结束
	ReqGameResult         // 请求游戏战绩
	ResGameResult         // 返回游戏战绩
)

// 俱乐部相关
const (
	ReqCreateClub = iota + 401 // 创建俱乐部
	ResCreateClub

	ReqClub // 获取指定俱乐部信息
	ResClub

	ReqClubs // 列出加入的俱乐部
	ResClubs

	ReqExitClub // 请求退出俱乐部
	ResExitClub

	ReqDelClub // 解散俱乐部
	BroadcastDelClub

	ReqEditClub // 修改俱乐部名称和公告
	BroadcastEditClub

	ReqJoinClub // 加入俱乐部
	BroadcastJoinClub

	ReqClubUsers // users会员列表
	ResClubUsers

	ReqEditClubUser // 编辑会员状态：action 设为管理(admin) 取消管理(-admin)  冻结(disable) 取消冻结(-disable) 设为代付(pay) 取消代付(-pay) 审核通过用户(add)  移除用户(-add)
	ResEditClubUser

	ReqCreateClubRoom // 创建俱乐部房间
	ResCreateClubRoom
)
