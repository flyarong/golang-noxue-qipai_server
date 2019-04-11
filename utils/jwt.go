package utils

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"qipai/config"
	"time"
)

// JWT 签名结构
type JWT struct {
	SigningKey []byte
}

// 一些常量
var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = config.Config.AppKey
)

// 载荷，可以加一些自己需要的信息
type UserInfo struct {
	Uid   uint   `json:"uid"`
	Nick  string `json:"nick"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取signKey
func GetSignKey() string {
	return SignKey
}

// 这是SignKey
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

// CreateToken 生成一个token
func (j *JWT) CreateToken(claims UserInfo) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析Tokne
func (j *JWT) ParseToken(tokenString string) (*UserInfo, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserInfo{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("令牌格式不合法") //TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, errors.New("登录超时，请重新登录") //TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("令牌尚未激活") //TokenNotValidYet
			} else {
				return nil, errors.New("无法处理此令牌") //TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*UserInfo); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("无法处理此令牌") //TokenInvalid
}

// 更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {

	token, err := jwt.ParseWithClaims(tokenString, &UserInfo{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return "", errors.New("令牌格式不合法") //TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return "", errors.New("登录超时，请重新登录") //TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return "", errors.New("令牌尚未激活1") //TokenNotValidYet
			} else {
				return "", errors.New("无法处理此令牌") //TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*UserInfo); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(30 * 24 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", errors.New("无法处理此令牌") //TokenInvalid
}
