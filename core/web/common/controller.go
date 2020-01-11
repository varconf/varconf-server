package common

import (
	"net/http"
	"strconv"
)

type Controller struct {
}

func (_self *Controller) ReadPageInfo(r *http.Request) (int64, int64) {
	params := r.URL.Query()
	pageIndex, err := strconv.ParseInt(params.Get("pageIndex"), 10, 64)
	if err != nil || pageIndex <= 0 {
		pageIndex = 1
	}

	pageSize, err := strconv.ParseInt(params.Get("pageSize"), 10, 64)
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	return pageIndex, pageSize
}

func (_self *Controller) WritePageData(w http.ResponseWriter, pageData interface{}, pageIndex, pageCount, pageSize, totalCount int64) {
	data := make(map[string]interface{})
	data["pageData"] = pageData
	data["pageIndex"] = pageIndex
	data["pageCount"] = pageCount
	data["pageSize"] = pageSize
	data["totalCount"] = totalCount

	WriteJson(w, ResponseData{Success: true, Code: CODE_SUCCEED, Data: data}, http.StatusOK)
}
