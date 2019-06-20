package router

import (
	"encoding/json"
	"github.com/golang/glog"
	"qipai/dao"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"time"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqGameStart, gameStart)
	game.AddAuthHandler(game.ReqSetTimes, gameSetTimes)
	game.AddAuthHandler(game.ReqSetScore, gameSetScore)
	game.AddAuthHandler(game.ReqGameResult, reqGameResult)

}

func reqGameResult(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		Page int `json:"page"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResGameResult, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	// 根据page查找对应的玩家数据
	var players []model.Player
	ret := dao.Db().Unscoped().Where(&model.Player{Uid: p.Uid}).Order("deleted_at desc").Limit(1).Offset(data.Page - 1).Find(&players)
	if ret.RecordNotFound() {
		glog.Error(ret.Error)
		res = utils.Msg("您没有历史战绩!").Code(-1)
		return
	}

	if len(players)<1{
		res = utils.Msg("没有更多历史战绩!").Code(-1)
		return
	}

	var room model.Room
	ret = dao.Db().Unscoped().First(&room, players[0].RoomId)
	if ret.RecordNotFound() {
		glog.Error(ret.Error)
		res = utils.Msg("没有找到房间信息!").Code(-1)
		return
	}

	type roomV struct {
		ID        uint           `json:"id"`
		Players   int            `json:"players"`
		Score     enum.ScoreType `json:"score"`
		Pay       enum.PayType   `json:"pay"`
		Count     int            `json:"count"`
		StartType enum.StartType `json:"start"`
		Times     int            `json:"times"`
		CreatedAt time.Time      `json:"createdAt"`
	}

	var rv roomV
	ok := utils.Copy(room, &rv)
	if !ok {
		res = utils.Msg("复制房间信息失败!").Code(-1)
		return
	}

	// 查询玩家信息
	var ps []model.Player
	ret = dao.Db().Unscoped().Where(&model.Player{RoomId: room.ID}).Find(&ps)
	if ret.RecordNotFound() {
		glog.Error(ret.Error)
		res = utils.Msg("没有找到玩家的对局信息!").Code(-1)
		return
	}

	// 查找玩家具体信息
	var users []model.User
	var uids []uint
	for _, ps := range ps {
		uids = append(uids, ps.Uid)
	}
	ret = dao.Db().Where("id in (?)", uids).Find(&users)
	if ret.RecordNotFound() {
		glog.Error(ret.Error)
		res = utils.Msg("没有找到用用户信息!").Code(-1)
		return
	}

	type gameV struct {
		Uid        uint   `json:"uid"`
		Nick       string `json:"nick"`
		Avatar     string `json:"avatar"`
		Banks      int    `json:"banks"`      // 庄家次数
		TotalScore int    `json:"totalScore"` // 总分
	}

	var gvs []gameV
	for _, v := range users {
		gvs = append(gvs, gameV{
			Uid:    v.ID,
			Nick:   v.Nick,
			Avatar: v.Avatar,
		})
	}

	// 总分
	for _, v1 := range ps {
		for k, v2 := range gvs {
			if v1.Uid != v2.Uid {
				continue
			}
			gvs[k].TotalScore = v1.TotalScore
		}
	}

	// 查找玩家对应的游戏结果
	var gs []model.Game
	ret = dao.Db().Unscoped().Where(&model.Game{RoomId: room.ID}).Find(&gs)
	if ret.RecordNotFound() {
		glog.Error(ret.Error)
		res = utils.Msg("没有找到玩家的对局信息!").Code(-1)
		return
	}

	// 填充游戏对局信息到用户中
	for _, v := range gs {
		for k, v1 := range gvs {
			if v.PlayerId != v1.Uid {
				continue
			}

			// 累计做庄次数
			if v.Banker {
				gvs[k].Banks ++
			}

		}
	}

	res = utils.Msg("").
		AddData("room", rv).
		AddData("page", data.Page).
		AddData("games", gvs)
}

func gameSetScore(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		RoomId uint `json:"roomId"`
		Score  int  `json:"score"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.BroadcastScore, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	e = srv.Game.SetScore(data.RoomId, uint(p.Uid), data.Score)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	res = nil
}

func gameSetTimes(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		RoomId uint `json:"roomId"`
		Times  int  `json:"times"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.BroadcastTimes, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	e = srv.Game.SetTimes(data.RoomId, uint(p.Uid), data.Times)
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	res = nil
}

func gameStart(s *zero.Session, msg *zero.Message) {
	type reqData struct {
		Id uint `json:"roomId"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResGameStart, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	p, e := game.GetPlayerFromSession(s)
	if e != nil {
		glog.Error(e)
		res = utils.Msg(e.Error()).Code(-1)
		return
	}

	e = srv.Game.Start(data.Id, uint(p.Uid))
	if e != nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	res = nil
}
