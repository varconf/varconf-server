package interceptor

import (
	"net/http"

	"varconf-server/core/moudle/router"
	"varconf-server/core/service"
)

type ApiAuthInterceptor struct {
	authService *service.AuthService
}

func InitApiAuthInterceptor(s *router.Router, authService *service.AuthService) *ApiAuthInterceptor {
	apiAuthInterceptor := ApiAuthInterceptor{authService: authService}

	s.AddFilter("/api(.*)", []string{}, &apiAuthInterceptor)

	return &apiAuthInterceptor
}

func (_self *ApiAuthInterceptor) PreHandleFunc(w http.ResponseWriter, r *http.Request, c *router.Context) bool {
	params := r.URL.Query()
	token := params.Get("token")
	if token == "" {
		http.Error(w, "Permission deny!", http.StatusForbidden)
		return false
	}

	success, appData := _self.authService.ApiAuth(token)
	if !success {
		http.Error(w, "Permission deny!", http.StatusForbidden)
		return false
	}

	c.Data["app"] = appData
	return true
}

func (_self *ApiAuthInterceptor) PostHandleFunc(w http.ResponseWriter, r *http.Request, c *router.Context) {
}
