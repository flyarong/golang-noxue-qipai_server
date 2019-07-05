package srv

import (
	"qipai/enum"
	"qipai/model"
	"testing"
)

func TestUserSrv_Register(t *testing.T) {
	user :=model.User{
		Nick:"admin",
	}
	user.Auths = append(user.Auths,model.Auth{
		Name:"13788888888",
		Pass:"137888888888",
		UserType:enum.MobilePass,
	})

	err:=User.Register(&user)
	if err!=nil {
		t.Error(err)
		return
	}
	t.Log("注册成功")
}
