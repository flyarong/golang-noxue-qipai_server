package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qipai/enum"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"strconv"
	"strings"
	"time"
)

func room() {
	r := R.Group("/rooms")
	r.Use(middleware.JWTAuth())

	// 创建房间
	r.POST("", roomCreateFunc)
	// 房间列表
	r.GET("", roomsFunc)
	// 房间信息
	r.GET("/:rid", roomInfoFunc)
	// 进入房间
	r.POST("/:rid/players", roomsJoinFunc)
	// 当前房间所有玩家信息
	r.GET("/:rid/players", roomsPlayersFunc)
	// 获取单个玩家的信息
	r.GET("/:rid/player", roomsPlayerFunc)
	// 获取单个玩家的信息
	r.GET("/:rid/player/:uid", roomsPlayerFunc)
	// 坐下
	r.PUT("/:rid/players/sit", roomsSitFunc)
	// 离开房间
	r.PUT("/:rid/players/exit", roomsExitFunc)

	// 开始游戏
	r.PUT("/:rid/start", roomsStartFunc)

	// 获取当前玩家的纸牌
	r.GET("/:rid/cards", roomsCardsFunc)
	// 获取指定用户的纸牌
	r.GET("/:rid/cards/:uid", roomsCardsFunc)

	// 下注
	r.PUT("/:rid/score/:score", roomsSetScore)

	// 解散房间
	r.DELETE("/:rid", roomsDeleteFunc)
}

func roomCreateFunc(c *gin.Context) {
	type roomForm struct {
		Score     enum.ScoreType `form:"score" json:"score"`                    // 底分方式
		Players   int            `form:"players" json:"players"`                // 玩家个数
		Count     int            `form:"count" json:"count" binding:"required"` // 局数
		StartType enum.StartType `form:"start" json:"start"`                    // 0 房主开始 1 首位开始
		Pay       enum.PayType   `form:"pay" json:"pay"`                        // 0 房主  1 AA
		King      enum.KingType  `form:"king" json:"king"`                      // 王癞 0 无王癞  1 经典王癞 2 疯狂王癞
		Special   int            `form:"special" json:"special"`                // 特殊牌型,二进制位表示特殊牌型翻倍规则，一共7类特殊牌型，用最低的7位二进制表示，1表示选中0表示没选中。
		Times     enum.TimesType `form:"times" json:"times"`                    // 翻倍规则
	}

	var form roomForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	// 限制只能 10  20 30 局
	if form.Count != 10 && form.Count != 20 && form.Count != 30 {
		c.JSON(http.StatusBadRequest, utils.Msg("局数[count]只能是10/20/30").Code(-2))
		return
	}

	// 限制游戏开始方式
	if form.StartType != 0 && form.StartType != 1 {
		c.JSON(http.StatusBadRequest, utils.Msg("开始方式[start]只能是0或1").Code(-3))
		return
	}

	// 限制支付模式
	if form.Pay != 0 && form.Pay != 1 {
		c.JSON(http.StatusBadRequest, utils.Msg("支付方式[pay]只能是0或1").Code(-4))
		return
	}
	// 限制王癞模式
	if form.King != 0 && form.King != 1 && form.King != 2 {
		c.JSON(http.StatusBadRequest, utils.Msg("王癞模式[king]只能是0/1/2").Code(-5))
		return
	}

	// 限制特殊牌型 全部选中状态为7位2进制都是1，最大为1111111==127
	if form.Special > 127 || form.Special < 0 {
		c.JSON(http.StatusBadRequest, utils.Msg("特殊牌型取值不合法").Code(-6))
		return
	}

	// 限制翻倍规则
	if form.Times < 0 || form.Times > 4 {
		c.JSON(http.StatusBadRequest, utils.Msg("翻倍规则[times]取值不合法，只能在0-4之间").Code(-7))
		return
	}

	// 底分取值不合法
	if form.Score < 0 || form.Score > 5 {
		c.JSON(http.StatusBadRequest, utils.Msg("底分类型取值只能在0-5之间").Code(-7))
		return
	}

	info := c.MustGet("user").(*utils.UserInfo)

	var room model.Room
	room.Uid = info.Uid

	if ok := utils.Copy(form, &room); !ok {
		c.JSON(http.StatusInternalServerError, utils.Msg("房间信息赋值失败，请联系管理员").Code(-8))
		return
	}

	if err := srv.Room.Create(&room); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-9))
		return
	}

	c.JSON(http.StatusOK, utils.Msg("创建成功").AddData("id", room.ID))
}

func roomsFunc(c *gin.Context) {
	type roomV struct {
		ID      uint           `json:"id"`
		Score   enum.ScoreType `json:"score"`   // 底分类型
		Pay     enum.PayType   `json:"pay"`     // 支付方式
		Current int            `json:"current"` // 当前第几局
		Count   int            `json:"count"`   // 总共可以玩几局
		Uid     uint           `json:"uid"`     // 房主用户编号
		Players int            `json:"players"` // 玩家个数
	}

	info := c.MustGet("user").(*utils.UserInfo)

	rooms := srv.Room.MyRooms(info.Uid)
	var roomsV []roomV
	for _, v := range rooms {
		var r roomV
		if !utils.Copy(v, &r) {
			c.JSON(http.StatusBadRequest, utils.Msg("内容转换出错").Code(-1))
			return
		}
		roomsV = append(roomsV, r)
	}
	c.JSON(http.StatusOK, utils.Msg("获取房间列表成功").AddData("rooms", roomsV))
}

func roomsJoinFunc(c *gin.Context) {
	info := c.MustGet("user").(*utils.UserInfo)

	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	err = srv.Room.Join(uint(rid), info.Uid, info.Nick)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("加入成功"))
}

func roomsPlayersFunc(c *gin.Context) {

	type playerV struct {
		Uid     uint   `json:"uid"`      // 用户编号
		Nick    string `json:"nick"`     // 昵称
		DeskId  int    `json:"desk_id"`  // 座位号
		IsReady bool   `json:"is_ready"` // 是否已准备
	}

	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}

	players := srv.Room.PlayersSitDown(uint(rid))
	var pvs []playerV
	for _, v := range players {
		var pv playerV
		if !utils.Copy(v, &pv) {
			c.JSON(http.StatusInternalServerError, utils.Msg("玩家数组赋值出错，请联系管理员").Code(-1))
			return
		}
		pvs = append(pvs, pv)
	}
	c.JSON(http.StatusOK, utils.Msg("获取玩家列表成功").AddData("players", pvs))
}

func roomsPlayerFunc(c *gin.Context) {

	type playerV struct {
		Uid      uint       `json:"uid"`       // 用户编号
		Nick     string     `json:"nick"`      // 昵称
		DeskId   int        `json:"desk_id"`   // 座位号
		RoomId   uint       `json:"room_id"`   // 房间编号
		Cards    string     `json:"cards"`     // 用户所拥有的牌
		JoinedAt *time.Time `json:"joined_at"` // 加入时间
	}

	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)
	uid := int(info.Uid)
	uidStr := c.Param("uid")
	if uidStr != "" {
		uid, err = strconv.Atoi(uidStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
			return
		}
	}

	p, err := srv.Room.Player(uint(rid), uint(uid))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
	var pv playerV
	if !utils.Copy(p, &pv) {
		c.JSON(http.StatusInternalServerError, utils.Msg("玩家信息拷贝出错").Code(-1))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("获取玩家信息成功").AddData("player", pv))
}

func roomsSitFunc(c *gin.Context) {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)

	roomId, deskId, e := srv.Room.SitDown(uint(rid), info.Uid)
	if e != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(e.Error()).Code(-1).AddData("room_id", roomId))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("坐下成功").AddData("desk_id", deskId))
}

func roomInfoFunc(c *gin.Context) {
	type roomV struct {
		ID      uint           `json:"id"`
		Score   enum.ScoreType `json:"score"`   // 底分类型
		Pay     enum.PayType   `json:"pay"`     // 支付方式
		Current int            `json:"current"` // 当前第几局
		Count   int            `json:"count"`   // 总共可以玩几局
		Uid     uint           `json:"uid"`     // 房主用户编号
		Players int            `json:"players"` // 玩家个数
	}

	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}

	room, err := srv.Room.Get(uint(rid))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
	var rv roomV
	if !utils.Copy(room, &rv) {
		c.JSON(http.StatusInternalServerError, utils.Msg("复制房间信息出错，请联系管理员").Code(-1))
		return
	}

	c.JSON(http.StatusOK, utils.Msg("获取房间信息成功").AddData("room", rv))
}

func roomsStartFunc(c *gin.Context) {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)
	err = srv.Room.Start(uint(rid), info.Uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("游戏已开始"))
}

func roomsExitFunc(c *gin.Context) {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)
	err = srv.Room.Exit(uint(rid), info.Uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("已离开房间"))
}

func roomsCardsFunc(c *gin.Context) {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)
	uid := info.Uid

	uidStr := c.Param("uid")
	if len(uidStr) > 0 {
		// 只有下注选定了庄家，才可以看别人的牌
		hasBanker := false
		gs, err := srv.Room.GetCurrentGames(uint(rid))
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
			return
		}
		for _, g := range gs {
			if g.Banker {
				hasBanker = true
				break
			}
		}
		// 没有庄家，提示错误
		if !hasBanker {
			c.JSON(http.StatusBadRequest, utils.Msg("选出庄家才可以查看其他玩家的牌").Code(-1))
			return
		}

		n, e := strconv.Atoi(uidStr)
		if e != nil {
			c.JSON(http.StatusBadRequest, utils.Msg("用户编号uid必须是数字").Code(-1))
			return
		}
		uid = uint(n)
	}

	game, err := srv.Room.GetCurrentGame(uint(rid), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}

	cards := game.Cards
	if game.Times == 0 {
		cs := strings.Split(cards, "|")
		cs = cs[:4]
		cards = strings.Join(cs, "|")
	}

	c.JSON(http.StatusOK, utils.Msg("获取纸牌成功").AddData("cards", cards))
}

func roomsSetScore(c *gin.Context) {
	rid, err := strconv.Atoi(c.Param("rid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
		return
	}
	score, err := strconv.Atoi(c.Param("score"))
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg("下注积分必须是数字").Code(-1))
		return
	}
	info := c.MustGet("user").(*utils.UserInfo)

	err = srv.Room.SetScore(uint(rid), info.Uid, score)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-1))
		return
	}
}

func roomsDeleteFunc(c *gin.Context) {
	//rid, err := strconv.Atoi(c.Param("rid"))
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, utils.Msg("房间编号必须是数字").Code(-1))
	//	return
	//}
	//info := c.MustGet("user").(*utils.UserInfo)
}
