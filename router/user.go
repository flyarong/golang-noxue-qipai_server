package router

import (
	"encoding/json"
	"qipai/game"
	"qipai/srv"
	"qipai/utils"
	"time"
	"zero"
)

func init() {
	game.AddAuthHandler(game.ReqUserInfo, userInfo)
}

func userInfo(s *zero.Session, msg *zero.Message) {
	type userV struct {
		ID        uint      `json:"id"`
		Nick      string    `gorm:"size:20" json:"nick"`
		Avatar    string    `gorm:"size:120" json:"avatar"`
		Ip        string    `gorm:"size:20" json:"ip"`
		Address   string    `gorm:"size:50" json:"address"`
		Card      int       `json:"card"`
		CreatedAt time.Time `json:"createdAt"`
	}

	type reqData struct {
		Id uint `json:"id"`
	}

	res := utils.Msg("")
	defer func() {
		if res == nil {
			return
		}
		res.Send(game.ResUserInfo, s)
	}()

	var data reqData
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	if data.Id == 0{
		p := game.GetPlayerFromSession(s)
		data.Id = uint(p.Uid)
	}

	user, err := srv.User.GetInfo(data.Id)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}
	var uv userV
	utils.Copy(user, &uv)

	res = utils.Msg("").AddData("user", uv)
}
