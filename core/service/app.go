package service

import (
	"database/sql"
	"github.com/google/uuid"
	"time"

	"varconf-server/core/dao"
)

type AppService struct {
	appDao      *dao.AppDao
	manageTxDao *dao.ManageTxDao
}

func NewAppService(db *sql.DB) *AppService {
	appService := AppService{
		appDao:      dao.NewAppDao(db),
		manageTxDao: dao.NewManageTxDao(db),
	}
	return &appService
}

func (_self *AppService) PageQuery(likeName string, pageIndex, pageSize int64) ([]*dao.AppData, int64, int64) {
	start := (pageIndex - 1) * pageSize
	end := start + pageSize

	pageData := _self.appDao.QueryApps(dao.QueryAppData{LikeName: likeName, Start: start, End: end})
	totalCount := _self.appDao.CountApps(dao.QueryAppData{LikeName: likeName})
	pageCount := totalCount / pageSize
	if totalCount%pageSize != 0 {
		pageCount += 1
	}
	return pageData, pageCount, totalCount
}

func (_self *AppService) QueryApp(appId int64) *dao.AppData {
	apps := _self.appDao.QueryApps(dao.QueryAppData{AppId: appId})
	if len(apps) != 1 {
		return nil
	}

	return apps[0]
}

func (_self *AppService) CreateApp(appData *dao.AppData) bool {
	if appData == nil {
		return false
	}

	appData.CreateTime.Time = time.Now()
	appData.UpdateTime.Time = time.Now()
	appData.ApiKey = appData.Code + ":" + uuid.New().String()
	rowCnt := _self.appDao.InsertApp(appData)
	if rowCnt != 1 {
		return false
	}
	return true
}

func (_self *AppService) SelectedUpdateApp(appData dao.AppData) bool {
	appData.UpdateTime.Time = time.Now()

	rowCnt := _self.appDao.SelectedUpdateApp(appData)
	if rowCnt != 1 {
		return false
	}
	return true
}

func (_self *AppService) DeleteApp(appId int64) bool {
	return _self.manageTxDao.DeleteApp(appId)
}
