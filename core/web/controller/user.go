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

type UserController struct {
	common.Controller

	authService *service.AuthService
	userService *service.UserService
}

func InitUserController(s *router.Router, authService *service.AuthService, userService *service.UserService) *UserController {
	userController := UserController{authService: authService, userService: userService}

	s.Get("/user/login", userController.login)
	s.Post("/user/logout", userController.logout)
	s.Get("/user/profile", userController.profile)
	s.Post("/user/passwd", userController.passwd)
	s.Get("/user", userController.list)
	s.Get("/user/:userId([0-9]+)", userController.detail)
	s.Delete("/user/:userId([0-9]+)", userController.delete)
	s.Put("/user", userController.create)
	s.Patch("/user/:userId([0-9]+)", userController.update)

	return &userController
}

// GET /user/login
func (_self *UserController) login(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	name := params.Get("name")
	password := params.Get("password")

	// login
	success, token := _self.authService.Login(name, password)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: token, Path: "/", Expires: time.Now().AddDate(0, 0, 1)})
	common.WriteSucceedResponse(w, token)
}

// POST /user/logout
func (_self *UserController) logout(w http.ResponseWriter, r *http.Request, c *router.Context) {
	http.SetCookie(w, &http.Cookie{Name: "token", Value: "", Path: "/", Expires: time.Now()})
	common.WriteSucceedResponse(w, nil)
}

// GET /user/profile
func (_self *UserController) profile(w http.ResponseWriter, r *http.Request, c *router.Context) {
	common.WriteSucceedResponse(w, c.Data["user"])
}

// POST /user/passwd
func (_self *UserController) passwd(w http.ResponseWriter, r *http.Request, c *router.Context) {
	r.ParseForm()
	password1 := r.FormValue("password1")
	password2 := r.FormValue("password2")

	// permission
	operator := c.Data["user"].(*dao.UserData)
	if operator == nil {
		common.WriteErrorResponse(w, nil)
		return
	}

	// query user detail
	userData := _self.userService.QueryUser(operator.UserId)
	if userData.Password != password1 {
		common.WriteErrorResponse(w, nil)
		return
	}

	// passwd user
	userData.Password = password2
	success := _self.userService.SelectedUpdateUser(*userData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}

// GET /user
func (_self *UserController) list(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read page
	pageIndex, pageSize := _self.ReadPageInfo(r)
	pageData, pageCount, totalCount := _self.userService.PageQuery(r.URL.Query().Get("likeName"), pageIndex, pageSize)
	for i := range pageData {
		pageData[i].Password = ""
	}

	// remove password
	for _, v := range pageData {
		v.Password = ""
	}

	_self.WritePageData(w, pageData, pageIndex, pageCount, pageSize, totalCount)
}

// GET /user/:userId([0-9]+)
func (_self *UserController) detail(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	userId, err := strconv.ParseInt(params.Get(":userId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// permission
	operator := c.Data["user"].(*dao.UserData)
	if operator == nil || operator.UserId != userId && operator.Permission != dao.USER_ADMIN {
		common.WriteErrorResponse(w, nil)
		return
	}

	// query user detail
	userData := _self.userService.QueryUser(userId)
	common.WriteSucceedResponse(w, userData)
}

// DELETE /user/:userId([0-9]+)
func (_self *UserController) delete(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	userId, err := strconv.ParseInt(params.Get(":userId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// permission
	operator := c.Data["user"].(*dao.UserData)
	if operator == nil || operator.Permission != dao.USER_ADMIN {
		common.WriteErrorResponse(w, nil)
		return
	}

	// delete user
	success := _self.userService.DeleteUser(userId)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}

// PUT /user
func (_self *UserController) create(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	userData := dao.UserData{}
	err := common.ReadJson(r, &userData)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// permission
	operator := c.Data["user"].(*dao.UserData)
	if operator == nil || operator.Permission != dao.USER_ADMIN {
		common.WriteErrorResponse(w, nil)
		return
	}

	// create user
	success := _self.userService.CreateUser(&userData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, userData)
}

// PATCH /user/:userId([0-9]+)
func (_self *UserController) update(w http.ResponseWriter, r *http.Request, c *router.Context) {
	// read param
	params := r.URL.Query()
	userId, err := strconv.ParseInt(params.Get(":userId"), 10, 64)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// permission
	operator := c.Data["user"].(*dao.UserData)
	if operator == nil || operator.UserId != userId && operator.Permission != dao.USER_ADMIN {
		common.WriteErrorResponse(w, nil)
		return
	}

	// query user
	userData := dao.UserData{}
	err = common.ReadJson(r, &userData)
	if err != nil {
		common.WriteErrorResponse(w, err.Error())
		return
	}

	// update user
	userData.UserId = userId
	success := _self.userService.SelectedUpdateUser(userData)
	if !success {
		common.WriteErrorResponse(w, nil)
		return
	}
	common.WriteSucceedResponse(w, nil)
}
