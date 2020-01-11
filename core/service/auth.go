package service

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"varconf-server/core/dao"
)

type AuthService struct {
	appDao       *dao.AppDao
	userDao      *dao.UserDao
	tokenTimeout int64
}

func NewAuthService(db *sql.DB) *AuthService {
	authService := AuthService{
		appDao:       dao.NewAppDao(db),
		userDao:      dao.NewUserDao(db),
		tokenTimeout: 48 * 60 * 60 * 1000,
	}
	return &authService
}

func (_self *AuthService) Login(name, password string) (bool, string) {
	users := _self.userDao.QueryUsers(dao.QueryUserData{Name: name})
	if len(users) > 0 {
		user := users[0]
		if user != nil && user.Name == name && user.Password == password {
			gen := _self.Gen(user.Name, user.Password)
			info := name + ":" + gen + ":" + strconv.FormatInt(time.Now().Unix(), 10)
			return true, base64.StdEncoding.EncodeToString([]byte(info))
		}
	}
	return false, ""
}

func (_self *AuthService) Auth(token string) (bool, *dao.UserData) {
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false, nil
	}

	arrays := strings.Split(string(bytes), ":")
	if len(arrays) != 3 {
		return false, nil
	}

	users := _self.userDao.QueryUsers(dao.QueryUserData{Name: arrays[0]})
	if len(users) != 1 {
		return false, nil
	}

	user := users[0]
	gen := _self.Gen(user.Name, user.Password)
	if gen == arrays[1] {
		stamp, err := strconv.ParseInt(arrays[2], 10, 64)
		if err == nil {
			if time.Now().Unix()-stamp < _self.tokenTimeout {
				return true, user
			}
		}
	}

	return false, nil
}

func (_self *AuthService) Gen(name, password string) string {
	hash := md5.New()
	io.WriteString(hash, name)
	io.WriteString(hash, password)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (_self *AuthService) ApiAuth(token string) (bool, *dao.AppData) {
	apps := _self.appDao.QueryApps(dao.QueryAppData{ApiKey: token})
	if len(apps) != 1 {
		return false, nil
	}
	return true, apps[0]
}
