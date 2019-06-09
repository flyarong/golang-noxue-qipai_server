package router

import (
	"encoding/json"
	"qipai/game"
	"qipai/srv"
	"qipai/utils"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqGameStart, gameStart)
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

	p := game.GetPlayerFromSession(s)

	e:=srv.Game.Start(data.Id,uint(p.Uid))
	if e!=nil {
		res = utils.Msg(e.Error()).Code(-1)
		return
	}
	res = nil
}
