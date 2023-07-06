package service

import (
	"crypto/md5"
	"dockernas/internal/config"
	"encoding/base64"
	"strings"

	"github.com/shirou/gopsutil/net"
)

var tokenMap = make(map[string]int64)
var fixed_token = ""

func init() {
	netInterfaces, _ := net.Interfaces()
	for _, v := range netInterfaces {
		fixed_token = base64.StdEncoding.EncodeToString([]byte(v.HardwareAddr))
		if len(fixed_token) != 0 {
			break
		}
	}

	if len(fixed_token) == 0 {
		bMd5 := md5.Sum([]byte("nasdocker"))
		fixed_token = base64.StdEncoding.EncodeToString(bMd5[:])
	}

	tokenMap[fixed_token] = 1
}

func IsTokenValid(token string) bool {
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
