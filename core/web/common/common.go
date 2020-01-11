package common

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	applicationJson = "application/json"
)

const (
	CODE_SUCCEED = 0
	CODE_FAILED  = -1
)

type ResponseData struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

func ReadJson(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func WriteJson(w http.ResponseWriter, v interface{}, code int) {
	content, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(code)
	w.Write(content)
}

func WriteErrorResponse(w http.ResponseWriter, data interface{}) {
	responseData := ResponseData{Success: false, Code: CODE_FAILED, Data: data}
	WriteJson(w, responseData, http.StatusOK)
}

func WriteErrorResponseWithCode(w http.ResponseWriter, data interface{}, httpCode int) {
	responseData := ResponseData{Success: false, Code: CODE_FAILED, Data: data}
	WriteJson(w, responseData, httpCode)
}

func WriteSucceedResponse(w http.ResponseWriter, data interface{}) {
	responseData := ResponseData{Success: true, Code: CODE_SUCCEED, Data: data}
	WriteJson(w, responseData, http.StatusOK)
}
