package controller

import (
	"net/http"
	"strconv"

	"varconf-server/core/dao"
	"varconf-server/core/moudle/router"
	"varconf-server/core/service"
	"varconf-server/core/web/common"
)

type AppController struct {
	common.Controller

	appService    *service.AppService
	configService *service.ConfigService
}

func InitAppController(s *router.Router, appService *service.AppService, configService *service.ConfigService) *AppController {
	appController := AppController{appService: appService, configService: configService}

	s.Get("/app", appController.list)
	s.Get("/app/:appId([0-9]+)", appController.detail)
	s.Delete("/app/:appId([0-9]+)", appController.delete)
	s.Put("/app", appController.create)
	s.Patch("/app/:appId([0-9]+)", appController.update)

	return &appController
}

// GET /app
func (_self *AppController) list(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read page
	pageIndex, pageSize := _self.ReadPageInfo(r)
	pageData, pageCount, totalCount := _self.appService.PageQuery(r.URL.Query().Get("likeName"), pageIndex, pageSize)

	// remove ApiKey
	for _, v := range pageData {
		v.ApiKey = ""
	}

	_self.WritePageData(w, pageData, pageIndex, pageCount, pageSize, totalCount)
}

// GET /app/:appId([0-9]+)/
func (_self *AppController) detail(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// query app
	appData := _self.appService.QueryApp(appId)
	common.WriteSucceedResponse(w, appData)
}

// DELETE /app/:appId([0-9]+)
func (_self *AppController) delete(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	appId, err := strconv.ParseInt(params.Get(":appId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// delete app
	success := _self.appService.DeleteApp(appId)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}

// PUT /app
func (_self *AppController) create(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	appData := dao.AppData{}
	err := common.ReadJson(r, &appData)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// create app
	success := _self.appService.CreateApp(&appData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, appData)
}

// PATCH /app/:appId([0-9]+)
func (_self *AppController) update(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	appData := dao.AppData{}
	err := common.ReadJson(r, &appData)
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

	// update app
	appData.AppId = appId
	success := _self.appService.SelectedUpdateApp(appData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, appData)
}
