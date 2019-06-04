package router

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"math/rand"
	"qipai/domain"
	"qipai/enum"
	"qipai/game"
	"qipai/model"
	"qipai/srv"
	"qipai/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zero"
)

func init() {
	game.AddHandler(game.ReqLogin, login)
	game.AddHandler(game.ReqReg, reg)
	game.AddHandler(game.ReqCode, code)
	game.AddHandler(game.ReqLoginByToken, loginByToken)
}

func login(s *zero.Session, msg *zero.Message) {

	resLogin := utils.Msg("")
	defer func() {
		resLogin.ToSend(game.ResLogin, s)
	}()

	var login domain.ReqLogin
	err := json.Unmarshal(msg.GetData(), &login)
	if err != nil {
		resLogin = utils.Msg(err.Error()).Code(-1)
		return
	}

	token, user, err := srv.User.Login(&model.Auth{UserType: login.Type, Name: login.Name, Pass: login.Pass})
	if err != nil {
		resLogin = utils.Msg(err.Error()).Code(-1)
		return
	}

	game.AddPlayer(s,&game.Player{
		Uid:  int(user.ID),
		Nick: user.Nick,
	})

	resLogin.AddData("token", token)
}

func reg(s *zero.Session, msg *zero.Message) {

	resReg := utils.Msg("")
	defer func() {
		resReg.ToSend(game.ResReg, s)
	}()
	var reg domain.ReqReg
	err := json.Unmarshal(msg.GetData(), &reg)
	if err != nil {
		return
	}
	if reg.UserType != enum.MobilePass {
		resReg = utils.Msg("目前仅支持手机注册").Code(-1)
		return
	}
	glog.V(3).Infoln("注册昵称：", reg.Nick)
	// 检查手机验证码，无论对错都删除验证码，防止暴力破解
	code := utils.Lv.Get("code_" + reg.Name)
	utils.Lv.Del("code_" + reg.Name)
	if code != reg.Code {
		resReg = utils.Msg("手机验证码错误").Code(-1)
		return
	}

	user := model.User{Nick: reg.Nick, Ip: s.GetConn().GetName(), Address: utils.GetAddress(strings.Split(s.GetConn().GetName(), ":")[0]),
		Auths: []model.Auth{{Name: reg.Name, UserType: reg.UserType, Pass: reg.Pass, Verified: true}}}
	user.Avatar = fmt.Sprintf("/avatar/Avatar%d.png", rand.Intn(199))
	err = srv.User.Register(&user)
	if err != nil {
		resReg = utils.Msg(err.Error()).Code(-1)
		return
	}
	resReg = utils.Msg("")
}

type smsCodeType int

const (
	smsCodeReg   smsCodeType = 1
	smsCodeLogin smsCodeType = 2
	smsCodeReset smsCodeType = 3
)

// 手机注册验证码
func code(s *zero.Session, msg *zero.Message) {

	resCode := utils.Msg("")
	defer func() {
		resCode.ToSend(game.ResCode, s)
	}()
	var reqCode domain.ReqCode
	err := json.Unmarshal(msg.GetData(), &reqCode)
	if err != nil {
		resCode = utils.Msg(err.Error()).Code(-1)
		return
	}

	reg := regexp.MustCompile(`^1[34578]\d{9}$`)
	if !reg.MatchString(reqCode.Phone) {
		resCode = utils.Msg("手机号格式不正确").Code(-1)
		return
	}

	if utils.Lv.Get("code_"+reqCode.Phone) != "" {
		resCode = utils.Msg("验证码已发送，请注意查收")
		return
	}

	code := rand.Intn(9000) + 1000

	ok := utils.SendSmsRegCode(reqCode.Phone, strconv.Itoa(code))
	if !ok {
		resCode = utils.Msg("验证码已发送失败，请联系客服").Code(-1)
		return
	}
	utils.Lv.PutEx("code_"+reqCode.Phone, fmt.Sprint(code), time.Minute*5)

	resCode = utils.Msg("获取成功，验证码5分钟内有效")
}

/***
通过token登录
 */
func loginByToken(s *zero.Session, msg *zero.Message) {
	var res = utils.Msg("")
	defer func() {
		if glog.V(3) {
			glog.Infoln("token登录:",res.ToJson())
		}
		res.ToSend(game.ResLoginByToken, s)
	}()
	var data domain.ReqLoginByToken
	err := json.Unmarshal(msg.GetData(), &data)
	if err != nil {
		res = utils.Msg(err.Error()).Code(-1)
		return
	}

	if data.Token == "" {
		res = utils.Msg("没有携带token").Code(-2)
		return
	}

	j := utils.NewJWT()
	// parseToken 解析token包含的信息
	user, err := j.ParseToken(data.Token)
	if err != nil {
		if err == utils.TokenExpired {
			res = utils.Msg("授权已过期").Code(-3)
			return
		}
		res = utils.Msg(err.Error()).Code(-3)
		return
	}

	game.AddPlayer(s,&game.Player{
		Uid:  int(user.Uid),
		Nick: user.Nick,
	})

	newToken, err := j.RefreshToken(data.Token)
	if err != nil {
		if err == utils.TokenExpired {
			res = utils.Msg("授权已过期").Code(-3)
			return
		}
		res = utils.Msg(err.Error()).Code(-3)
		return
	}

	res.AddData("token", newToken)
}
