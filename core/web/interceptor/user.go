package interceptor

import (
	"net/http"

	"varconf-server/core/moudle/router"
	"varconf-server/core/service"
)

type UserAuthInterceptor struct {
	authService *service.AuthService
}

func InitUserAuthInterceptor(s *router.Router, authService *service.AuthService) *UserAuthInterceptor {
	authInterceptor := UserAuthInterceptor{authService: authService}

	s.AddFilter("/(.*)", []string{"/", "/static(.*)", "/api(.*)", "/user/login", "/user/logout"}, &authInterceptor)

	return &authInterceptor
}

func (_self *UserAuthInterceptor) PreHandleFunc(w http.ResponseWriter, r *http.Request, c *router.Context) bool {
	token, err := r.Cookie("token")
	if token == nil || err != nil {
		http.Error(w, "Permission deny!", http.StatusForbidden)
		return false
	}

	success, userData := _self.authService.Auth(token.Value)
	if !success {
		http.Error(w, "Permission deny!", http.StatusForbidden)
		return false
	}

	userData.Password = ""
	c.Data["user"] = userData
	return true
}

func (_self *UserAuthInterceptor) PostHandleFunc(w http.ResponseWriter, r *http.Request, c *router.Context) {
}
