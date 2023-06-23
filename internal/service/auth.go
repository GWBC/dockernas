package service

import (
	"dockernas/internal/config"
	"strings"
)

var tokenMap = make(map[string]int64)
var fixed_token = "aaaaaaaaaaaaaaaaaaaaaaaaaa"

func init() {
	tokenMap[fixed_token] = 1
}

func IsTokenValid(token string) bool {
	// log.Println(userToken)
	_, ok := tokenMap[token]
	return ok
}

func GenToken(user string, passwd string) string {
	realUserName, realPasswd := config.GetUserInfo()
	if !strings.EqualFold(realUserName, user) || realPasswd != passwd {
		panic("user password error")
	}

	//userToken := uuid.New().String()
	//tokenMap[userToken] = time.Now().UnixMilli()
	userToken := fixed_token //使用固定token，登录一次后就不用在登录了，但是不安全

	return userToken
}
