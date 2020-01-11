package service

import (
	"database/sql"
	"varconf-server/core/dao"
)

type HomeService struct {
	appDao        *dao.AppDao
	userDao       *dao.UserDao
	configDao     *dao.ConfigDao
	releaseLogDao *dao.ReleaseLogDao
}

func NewHomeService(db *sql.DB) *HomeService {
	homeService := HomeService{
		appDao:        dao.NewAppDao(db),
		userDao:       dao.NewUserDao(db),
		configDao:     dao.NewConfigDao(db),
		releaseLogDao: dao.NewReleaseLogDao(db),
	}
	return &homeService
}

func (_self *HomeService) Overall() map[string]interface{} {
	dataMap := make(map[string]interface{})

	count := _self.appDao.CountApps(dao.QueryAppData{})
	dataMap["app"] = count

	count = _self.userDao.CountUsers(dao.QueryUserData{})
	dataMap["user"] = count

	count = _self.configDao.CountConfigs(dao.QueryConfigData{})
	dataMap["config"] = count

	count = _self.releaseLogDao.CountReleaseLogs()
	dataMap["releaseLog"] = count

	return dataMap
}
