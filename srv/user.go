package srv

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"qipai/dao"
	"qipai/enum"
	"qipai/model"
	"qipai/utils"
	"time"
)

var User userSrv

func initUser() {
	User.j = utils.NewJWT()
}

type userSrv struct {
	j *utils.JWT
}

func (userSrv) Register(user *model.User) (err error) {

	if len(user.Auths) != 1 {
		err = errors.New("缺少账号信息")
		return
	}
	// 检查授权信息是否存在
	a := model.Auth{UserType: user.Auths[0].UserType, Name: user.Auths[0].Name}
	dao.Db().Where(&a).First(&a)

	if a.Model.ID > 0 {
		err = errors.New(a.Name + " 已被注册，请更换一个账号")
		return
	}

	dao.Db().Create(&user)
	return
}

func (userSrv) Bind(uid uint, auth *model.Auth) (err error) {

	// 检查用户是否存在
	var u model.User
	dao.Db().Where("id=?", uid).First(&u)
	if u.ID == 0 {
		err = errors.New("要绑定的用户不存在")
		return
	}

	// 检查授权信息是否存在
	a := model.Auth{UserType: auth.UserType, Name: auth.Name}
	dao.Db().Where(&a).First(&a)

	if a.Model.ID > 0 {
		err = errors.New(a.Name + " 已被绑定过，请更换一个账号")
		return
	}

	dao.Db().Create(auth)

	return
}

func (this userSrv) Login(auth *model.Auth) (token string,u model.User, err error) {
	var a model.Auth
	dao.Db().Where(&model.Auth{UserType: auth.UserType, Name: auth.Name}).First(&a)

	if a.ID == 0 {
		err = errors.New("账号不存在，请确认登录类型和账号正确")
		return
	}

	if !utils.PassCompare(auth.Pass, a.Pass) {
		err = errors.New("密码不正确")
		return
	}


	res:=dao.Db().Where(a.UserId).First(&u)
	if res.Error!=nil {
		err = errors.New("查询用户信息出错")
		return
	}
	if res.RecordNotFound() {
		err = errors.New("用户信息不存在")
		return
	}

	admin := false
	if u.ID == 1 {
		admin = true
	}

	claims := utils.UserInfo{
		Uid:   u.ID,
		Nick:  u.Nick,
		Admin: admin,
		StandardClaims: jwt.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000),       // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 3600*24*30), // 过期时间 一小时
			Issuer:    "qipai",                               //签名的发行者
		},
	}

	token, err = this.j.CreateToken(claims)

	return
}

func (this userSrv) TokenRefresh(token string) (string, error) {
	return this.j.RefreshToken(token)
}

func (userSrv) GetInfo(uid uint) (*model.User, error) {
	var user model.User
	dao.Db().Where("id=?", uid).First(&user)

	if user.ID == 0 {
		return nil, errors.New(fmt.Sprintf("没找到id为[%d]的用户信息", uid))
	}
	return &user, nil
}


func (userSrv)ChangePass(userType enum.UserType, name, pass string) (err error){
	auth,e:=dao.Auth.Get(userType,name)
	if e!=nil{
		err = e
		return
	}

	auth.Pass = pass
	dao.Db().Save(&auth)

	return
}