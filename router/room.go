package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"qipai/enum"
	"qipai/middleware"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
)

func room() {
	r := R.Group("/rooms")
	r.Use(middleware.JWTAuth())

	// 创建房间
	r.POST("", roomCreateFunc)
	// 房间列表
	r.GET("", roomsFunc)
	// 解散房间
	// 进入房间
	// 选座位
	// 坐下
	// 离开房间
}

func roomCreateFunc(c *gin.Context) {
	type roomForm struct {
		Score     enum.ScoreType `form:"score" json:"score"`                    // 底分方式
		Players   int            `form:"players" json:"players"`                // 玩家个数
		Count     int            `form:"count" json:"count" binding:"required"` // 局数
		StartType enum.StartType `form:"start" json:"start"`                    // 0 第一个入场的开始  1 全准备好开始
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
		c.JSON(http.StatusBadRequest, utils.Msg("房间信息赋值失败，请联系管理员").Code(-8))
		return
	}

	if err := srv.Room.Create(&room); err != nil {
		c.JSON(http.StatusBadRequest, utils.Msg(err.Error()).Code(-9))
		return
	}
	c.JSON(http.StatusOK, utils.Msg("创建成功"))
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
