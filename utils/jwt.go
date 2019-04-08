package utils

import (
	"github.com/dgrijalva/jwt-go"
	"qipai/config"
	"time"
)

func NewToken(uid uint, username string) (string, error) {
	// 带权限创建令牌
	claims := make(jwt.MapClaims)
	claims["uid"] = uid
	claims["username"] = username
	if username == "admin" {
		claims["admin"] = "true"
	} else {
		claims["admin"] = "false"
	}
	claims["exp"] = time.Now().Add(time.Hour * 480).Unix() //20天有效期，过期需要重新登录获取token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用自定义字符串加密 and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(config.Config.AppKey))
	return tokenString, err
}

func Parse(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.AppKey), nil
	})
	return token, err
}
