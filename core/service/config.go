package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"strconv"
	"strings"
	"time"

	"varconf-server/core/dao"
	"varconf-server/core/moudle/poll"
)

type ConfigService struct {
	appDao        *dao.AppDao
	configDao     *dao.ConfigDao
	releaseDao    *dao.ReleaseDao
	releaseLogDao *dao.ReleaseLogDao
	manageTxDao   *dao.ManageTxDao
	messagePoll   *poll.MessagePoll
	lastIndexMap  map[string]int
}

func NewConfigService(db *sql.DB) *ConfigService {
	configService := ConfigService{
		appDao:        dao.NewAppDao(db),
		configDao:     dao.NewConfigDao(db),
		releaseDao:    dao.NewReleaseDao(db),
		releaseLogDao: dao.NewReleaseLogDao(db),
		manageTxDao:   dao.NewManageTxDao(db),
		messagePoll:   poll.NewMessagePoll(),
		lastIndexMap:  make(map[string]int),
	}
	return &configService
}

func (_self *ConfigService) PageQuery(appId int64, likeKey string, pageIndex, pageSize int64) ([]*dao.ConfigData, int64, int64) {
	start := (pageIndex - 1) * pageSize
	end := pageSize

	pageData := _self.configDao.QueryConfigs(dao.QueryConfigData{AppId: appId, LikeKey: likeKey, Start: start, End: end})
	totalCount := _self.configDao.CountConfigs(dao.QueryConfigData{AppId: appId, LikeKey: likeKey})
	pageCount := totalCount / pageSize
	if totalCount%pageSize != 0 {
		pageCount += 1
	}
	return pageData, pageCount, totalCount
}

func (_self *ConfigService) QueryConfig(appId, configId int64) *dao.ConfigData {
	configs := _self.configDao.QueryConfigs(dao.QueryConfigData{AppId: appId, ConfigId: configId})
	if len(configs) != 1 {
		return nil
	}
	return configs[0]
}

func (_self *ConfigService) CreateConfig(data *dao.ConfigData) bool {
	data.Operate = dao.OPERATE_NEW
	data.Status = dao.STATUS_UN
	data.CreateTime.Time = time.Now()
	data.UpdateTime.Time = time.Now()

	rowCnt := _self.configDao.InsertConfig(data)
	if rowCnt != 1 {
		return false
	}

	return true
}

func (_self *ConfigService) UpdateConfig(data dao.ConfigData) bool {
	data.Operate = dao.OPERATE_UPDATE
	data.Status = dao.STATUS_UN
	data.UpdateTime.Time = time.Now()

	rowCnt := _self.configDao.SelectedUpdateConfig(data)
	if rowCnt != 1 {
		return false
	}

	return true
}

func (_self *ConfigService) DeleteConfig(data dao.ConfigData) bool {
	data.Operate = dao.OPERATE_DELETE
	data.Status = dao.STATUS_UN
	data.UpdateTime.Time = time.Now()

	rowCnt := _self.configDao.SelectedUpdateConfig(data)
	if rowCnt != 1 {
		return false
	}

	return true
}

func (_self *ConfigService) ReleaseConfig(appId int64, user string) bool {
	// query all config
	configs := _self.configDao.QueryConfigs(dao.QueryConfigData{AppId: appId})
	if len(configs) < 1 {
		return false
	}

	// parse allConfigs and update config status
	success, keys := _self.manageTxDao.ReleaseConfig(appId, configs, user)
	if !success {
		return false
	}

	// push message
	_self.pushRelease(appId, keys)
	return true
}

func (_self *ConfigService) QueryRelease(appId int64) ([]dao.ConfigData, int) {
	releaseData := _self.releaseDao.QueryRelease(appId)
	if releaseData == nil {
		return nil, 0
	}

	configList := make([]dao.ConfigData, 0)
	if err := json.Unmarshal([]byte(releaseData.ConfigList), &configList); err != nil {
		return nil, 0
	}
	return configList, releaseData.ReleaseIndex
}

func (_self *ConfigService) CronRelease(spec string) {
	c := cron.New()
	c.AddFunc(spec, func() {
		// query keys
		keys := _self.messagePoll.Keys()
		if keys == nil || len(keys) == 0 {
			return
		}

		// parse appId and lastIndex
		appIds := make([]int64, 0, len(keys))
		appIndexMap := make(map[int64]int)
		for _, key := range keys {
			// parse lastIndex
			lastIndex, exist := _self.lastIndexMap[key]
			if !exist {
				continue
			}

			// parse appId
			arrays := strings.Split(key, "_")
			if len(arrays) != 2 && len(arrays) != 3 {
				continue
			}
			appId, err := strconv.ParseInt(arrays[1], 10, 64)
			if err != nil {
				continue
			}

			appIds = append(appIds, appId)
			appIndexMap[appId] = lastIndex
		}
		if len(appIds) < 1 {
			return
		}

		// query release data
		releases := _self.releaseDao.QueryReleases(appIds)
		if releases == nil || len(releases) == 0 {
			return
		}

		// parse release data
		for _, release := range releases {
			// check app is have been released
			lastIndex := appIndexMap[release.AppId]
			if lastIndex == release.ReleaseIndex {
				continue
			}

			// parse config data
			configList := make([]dao.ConfigData, 0)
			if err := json.Unmarshal([]byte(release.ConfigList), &configList); err != nil {
				continue
			}
			if configList == nil || len(configList) == 0 {
				continue
			}

			// parse key and push data
			appKeys := make([]string, 0, len(configList))
			for _, config := range configList {
				appKeys = append(appKeys, config.Key)
			}
			_self.pushRelease(release.AppId, appKeys)
		}
	})
	c.Start()
}

func (_self *ConfigService) PullRelease(appId int64, key string, lastIndex int) (*poll.MessagePoll, *poll.Element) {
	pollKey := fmt.Sprintf("app_%d", appId)
	if key != "" {
		pollKey = fmt.Sprintf("key_%d_%s", appId, key)
	}

	// put key's lastIndex
	_self.lastIndexMap[pollKey] = lastIndex

	// long poll for config
	return _self.messagePoll, _self.messagePoll.Poll(pollKey)
}

func (_self *ConfigService) pushRelease(appId int64, keys []string) {
	pollKey := fmt.Sprintf("app_%d", appId)
	if _self.messagePoll.Contain(pollKey) {
		_self.messagePoll.Push(pollKey, appId)
	}

	if keys == nil {
		return
	}

	for _, key := range keys {
		pollKey = fmt.Sprintf("key_%d_%s", appId, key)
		if _self.messagePoll.Contain(pollKey) {
			_self.messagePoll.Push(pollKey, key)
		}
	}
}
