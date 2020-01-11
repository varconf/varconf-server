package controller

import (
	"net/http"
	"strconv"

	"varconf-server/core/dao"
	"varconf-server/core/moudle/router"
	"varconf-server/core/service"
	"varconf-server/core/web/common"
)

type ConfigController struct {
	common.Controller

	configService *service.ConfigService
}

func InitConfigController(s *router.Router, configService *service.ConfigService) *ConfigController {
	configController := ConfigController{configService: configService}

	s.Get("/config/:appId([0-9]+)", configController.list)
	s.Post("/config/:appId([0-9]+)/release", configController.release)
	s.Get("/config/:appId([0-9]+)/:configId([0-9]+)", configController.detail)
	s.Delete("/config/:appId([0-9]+)/:configId([0-9]+)", configController.delete)
	s.Put("/config/:appId([0-9]+)", configController.create)
	s.Patch("/config/:appId([0-9]+)/:configId([0-9]+)", configController.update)

	return &configController
}

// GET /config/:appId([0-9]+)
func (_self *ConfigController) list(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// read config
	pageIndex, pageSize := _self.ReadPageInfo(r)
	pageData, pageCount, totalCount := _self.configService.PageQuery(appId, params.Get("likeKey"), pageIndex, pageSize)

	_self.WritePageData(w, pageData, pageIndex, pageCount, pageSize, totalCount)
}

// POST /config/:appId([0-9]+)/release
func (_self *ConfigController) release(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// release config
	user := context.Data["user"].(*dao.UserData)
	success := _self.configService.ReleaseConfig(appId, user.Name)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}

// GET /config/:appId([0-9]+)/:configId([0-9]+)
func (_self *ConfigController) detail(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	configId, err := strconv.ParseInt(params.Get(":configId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// query config
	configData := _self.configService.QueryConfig(appId, configId)
	common.WriteSucceedResponse(w, configData)
}

// DELETE /config/:appId([0-9]+)/:configId([0-9]+)
func (_self *ConfigController) delete(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	configId, err := strconv.ParseInt(params.Get(":configId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// delete config
	user := context.Data["user"].(*dao.UserData)
	configData := dao.ConfigData{}
	configData.AppId = appId
	configData.ConfigId = configId
	configData.UpdateBy = user.Name

	success := _self.configService.DeleteConfig(configData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}

// PUT /config/:appId([0-9]+)
func (_self *ConfigController) create(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	configData := dao.ConfigData{}
	err := common.ReadJson(r, &configData)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// create config
	user := context.Data["user"].(*dao.UserData)
	configData.AppId = appId
	configData.CreateBy = user.Name
	configData.UpdateBy = user.Name

	success := _self.configService.CreateConfig(&configData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, configData)
}

// PATCH /config/:appId([0-9]+)/:configId([0-9]+)
func (_self *ConfigController) update(w http.ResponseWriter, r *http.Request, context *router.Context) {
	// read param
	configData := dao.ConfigData{}
	err := common.ReadJson(r, &configData)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	configId, err := strconv.ParseInt(params.Get(":configId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// update config data
	user := context.Data["user"].(*dao.UserData)
	configData.AppId = appId
	configData.ConfigId = configId
	configData.UpdateBy = user.Name

	success := _self.configService.UpdateConfig(configData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, configData)
}
