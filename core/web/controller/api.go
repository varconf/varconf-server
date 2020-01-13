package controller

import (
	"net/http"
	"strconv"
	"time"

	"varconf-server/core/dao"
	"varconf-server/core/moudle/router"
	"varconf-server/core/service"
	"varconf-server/core/web/common"
)

type ApiController struct {
	common.Controller

	authService   *service.AuthService
	configService *service.ConfigService
}

type ConfigValue struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

func InitApiController(s *router.Router, authService *service.AuthService, configService *service.ConfigService) *ApiController {
	apiController := ApiController{authService: authService, configService: configService}

	s.Get("/api/config", apiController.watchApp)
	s.Get("/api/config/:key", apiController.watchKey)

	return &apiController
}

// GET /api/config
func (_self *ApiController) watchApp(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// get appData from context
	appData := c.Data["app"].(*dao.AppData)
	if appData == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// parse releaseIndex form url
	params := r.URL.Query()
	lastIndex, _ := strconv.ParseInt(params.Get("lastIndex"), 10, 32)
	longPull, _ := strconv.ParseBool(params.Get("longPull"))

	// http long poll for config
	if longPull == true {
		_self.pullAndResponse(w, appData.AppId, "", int(lastIndex))
		return
	}

	// query config
	_self.queryAndResponse(w, appData.AppId, "", 0, true)
}

// GET /api/config/:key
func (_self *ApiController) watchKey(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// get appData from context
	appData := c.Data["app"].(*dao.AppData)
	if appData == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// parse key and releaseIndex form url
	params := r.URL.Query()
	key := params.Get(":key")
	lastIndex, _ := strconv.ParseInt(params.Get("lastIndex"), 10, 32)
	longPull, _ := strconv.ParseBool(params.Get("longPull"))

	// http long poll for config
	if longPull == true {
		_self.pullAndResponse(w, appData.AppId, key, int(lastIndex))
		return
	}

	// query config
	_self.queryAndResponse(w, appData.AppId, key, 0, true)
}

func (_self *ApiController) pullAndResponse(w http.ResponseWriter, appId int64, key string, lastIndex int) {
	success := _self.queryAndResponse(w, appId, key, lastIndex, false)
	if success {
		return
	}

	messagePoll, pollElement := _self.configService.PullRelease(appId, key, lastIndex)
	select {
	case <-pollElement.Chan():
		messagePoll.Remove(pollElement)
		_self.queryAndResponse(w, appId, key, 0, true)

	case <-time.After(60 * time.Second):
		messagePoll.Remove(pollElement)
		http.Error(w, "", http.StatusNotModified)

	case <-w.(http.CloseNotifier).CloseNotify():
		messagePoll.Remove(pollElement)
	}
}

func (_self *ApiController) queryAndResponse(w http.ResponseWriter, appId int64, key string, lastIndex int, lastCall bool) bool {
	configMap, recentIndex := _self.queryReleaseConfig(appId, lastIndex)
	if configMap == nil {
		if lastCall {
			http.Error(w, "", http.StatusNotFound)
		}
		return false
	}

	dataMap := make(map[string]interface{})
	if key != "" {
		// key watch
		configValue := configMap[key]
		if configValue == nil {
			if lastCall {
				http.Error(w, "", http.StatusNotFound)
			}
			return false
		}

		dataMap["recentIndex"] = recentIndex
		dataMap["data"] = configValue
	} else {
		// app watch
		if len(configMap) == 0 {
			if lastCall {
				http.Error(w, "", http.StatusNotFound)
			}
			return false
		}

		dataMap["recentIndex"] = recentIndex
		dataMap["data"] = configMap
	}

	common.WriteJson(w, dataMap, http.StatusOK)
	return true
}

func (_self *ApiController) queryReleaseConfig(appId int64, lastIndex int) (map[string]*ConfigValue, int) {
	configList, releaseIndex := _self.configService.QueryRelease(appId)
	if configList == nil || releaseIndex == lastIndex {
		return nil, 0
	}

	configMap := make(map[string]*ConfigValue)
	for _, configData := range configList {
		configMap[configData.Key] = &ConfigValue{
			Key:       configData.Key,
			Value:     configData.Value,
			Timestamp: configData.UpdateTime.Unix(),
		}
	}

	return configMap, releaseIndex
}
