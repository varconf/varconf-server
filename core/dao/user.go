package dao

import (
	"bytes"
	"database/sql"
	"strings"

	"varconf-server/core/dao/common"
)

const (
	USER_ORDINARY = 1
	USER_ADMIN    = 2
)

type UserData struct {
	UserId     int64           `json:"userId" DB_COL:"user_id" DB_PK:"user_id" DB_TABLE:"user"`
	Name       string          `json:"name" DB_COL:"name"`
	Password   string          `json:"password" DB_COL:"password"`
	Permission int             `json:"permission,string,omitempty" DB_COL:"permission"`
	CreateTime common.JsonTime `json:"createTime" DB_COL:"create_time"`
	UpdateTime common.JsonTime `json:"UpdateTime" DB_COL:"update_time"`
}

type QueryUserData struct {
	UserId   int64
	Name     string
	LikeName string
	Start    int64
	End      int64
}

type UserDao struct {
	common.Dao
}

func NewUserDao(db *sql.DB) *UserDao {
	userDao := UserDao{common.Dao{DB: db}}
	return &userDao
}

func (_self *UserDao) QueryUsers(queryUserData QueryUserData) []*UserData {
	sql, values := _self.prepareSelectedQuery(false, queryUserData)
	users := make([]*UserData, 0)
	success, err := _self.StructSelect(&users, sql, values...)
	if err != nil {
		panic(err)
	}
	if success {
		return users
	}
	return nil
}

func (_self *UserDao) CountUsers(queryUserData QueryUserData) int64 {
	sql, values := _self.prepareSelectedQuery(true, queryUserData)
	return _self.Count(sql, values...)
}

func (_self *UserDao) InsertUser(user *UserData) int64 {
	rowCnt, err := _self.StructInsert(user, false)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *UserDao) SelectedUpdateUser(user UserData) int64 {
	sql, values := _self.prepareSelectedUpdate(user)
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *UserDao) DeleteUser(userId int64) int64 {
	sql := "DELETE FROM `user` WHERE `user_id` = ?"
	rowCnt, err := _self.Exec(sql, userId)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *UserDao) prepareSelectedQuery(count bool, queryUserData QueryUserData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("SELECT")
	if count {
		buffer.WriteString(" COUNT(1)")
	} else {
		buffer.WriteString(" *")
	}
	buffer.WriteString(" FROM `user` WHERE 1 = 1")

	values := make([]interface{}, 0)
	if queryUserData.UserId != 0 {
		buffer.WriteString(" AND `user_id` = ?")
		values = append(values, queryUserData.UserId)
	}
	if queryUserData.Name != "" {
		buffer.WriteString(" AND `name` = ?")
		values = append(values, queryUserData.Name)
	}
	if queryUserData.LikeName != "" {
		buffer.WriteString(" AND `name` like '")
		buffer.WriteString(queryUserData.LikeName)
		buffer.WriteString("%'")
	}
	if queryUserData.Start >= 0 && queryUserData.End > 0 {
		buffer.WriteString(" LIMIT ?, ?")
		values = append(values, queryUserData.Start, queryUserData.End)
	}

	return buffer.String(), values
}

func (_self *UserDao) prepareSelectedUpdate(user UserData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("UPDATE `user` SET ")

	values := make([]interface{}, 0)
	if user.Name != "" {
		values = append(values, user.Name)
		buffer.WriteString("`name` = ?,")
	}
	if user.Password != "" {
		values = append(values, user.Password)
		buffer.WriteString("`password` = ?,")
	}
	if user.Permission != 0 {
		values = append(values, user.Permission)
		buffer.WriteString("`permission` = ?,")
	}
	if !user.CreateTime.IsZero() {
		values = append(values, user.CreateTime)
		buffer.WriteString("`create_time` = ?,")
	}
	if !user.UpdateTime.IsZero() {
		values = append(values, user.UpdateTime)
		buffer.WriteString("`update_time` = ?,")
	}

	sql := strings.TrimSuffix(buffer.String(), ",") + " WHERE `user_id` = ?"
	values = append(values, user.UserId)

	return sql, values
}
