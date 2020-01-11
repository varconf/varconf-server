package resolver

import (
	"net/http"

	"varconf-server/core/moudle/router"
	"varconf-server/core/web/common"
)

type ErrorResolver struct {
}

func InitErrorRecover(s *router.Router) *ErrorResolver {
	errorRecover := ErrorResolver{}

	s.AddResolver(nil, errorRecover.Error)

	return &errorRecover
}

func (_self *ErrorResolver) Error(w http.ResponseWriter, r *http.Request, err error) {
	common.WriteErrorResponse(w, err.Error())
}
