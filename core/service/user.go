package service

import (
	"database/sql"
	"time"

	"varconf-server/core/dao"
)

type UserService struct {
	userDao *dao.UserDao
}

func NewUserService(db *sql.DB) *UserService {
	userService := UserService{
		userDao: dao.NewUserDao(db),
	}
	return &userService
}

func (_self *UserService) PageQuery(likeName string, pageIndex, pageSize int64) ([]*dao.UserData, int64, int64) {
	start := (pageIndex - 1) * pageSize
	end := start + pageSize

	pageData := _self.userDao.QueryUsers(dao.QueryUserData{LikeName: likeName, Start: start, End: end})
	totalCount := _self.userDao.CountUsers(dao.QueryUserData{LikeName: likeName})
	pageCount := totalCount / pageSize
	if totalCount%pageSize != 0 {
		pageCount += 1
	}
	return pageData, pageCount, totalCount
}

func (_self *UserService) QueryUser(userId int64) *dao.UserData {
	users := _self.userDao.QueryUsers(dao.QueryUserData{UserId: userId})
	if len(users) > 0 {
		return users[0]
	}
	return nil
}

func (_self *UserService) CreateUser(userData *dao.UserData) bool {
	users := _self.userDao.QueryUsers(dao.QueryUserData{Name: userData.Name})
	if len(users) != 0 {
		return false
	}

	userData.CreateTime.Time = time.Now()
	userData.UpdateTime.Time = time.Now()

	rowCnt := _self.userDao.InsertUser(userData)
	if rowCnt != 1 {
		return false
	}
	return true
}

func (_self *UserService) SelectedUpdateUser(userData dao.UserData) bool {
	userData.UpdateTime.Time = time.Now()

	rowCnt := _self.userDao.SelectedUpdateUser(userData)
	if rowCnt != 1 {
		return false
	}
	return true
}

func (_self *UserService) DeleteUser(userId int64) bool {
	rowCnt := _self.userDao.DeleteUser(userId)
	if rowCnt != 1 {
		return false
	}
	return true
}
