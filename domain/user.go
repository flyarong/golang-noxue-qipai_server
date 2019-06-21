package domain

import (
	"qipai/enum"
	"zero"
)

type ReqLogin struct {
	Type    enum.UserType `form:"type" json:"type" binding:"required"`
	Name    string        `form:"name" json:"name" binding:"required"`
	Pass    string        `form:"pass" json:"pass" binding:"required"`
	Session *zero.Session `json:"-"`
}


type ReqReg struct {
	UserType enum.UserType `form:"type" json:"type" binding:"required"`
	Nick     string        `form:"nick" json:"nick" binding:"required"`
	Pass     string        `form:"pass" json:"pass" binding:"required"`
	Name     string        `form:"name" json:"name" binding:"required"`
	Code     string        `form:"code" json:"code" binding:"required"`
}



type ReqLoginByToken struct {
	Token string `json:"token"`
}