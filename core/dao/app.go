package dao

import (
	"bytes"
	"database/sql"
	"strings"

	"varconf-server/core/dao/common"
)

// App信息
type AppData struct {
	AppId        int64           `json:"appId" DB_COL:"app_id" DB_PK:"app_id" DB_TABLE:"app"`
	Name         string          `json:"name" DB_COL:"name"`
	Code         string          `json:"code" DB_COL:"code"`
	Desc         string          `json:"desc" DB_COL:"desc"`
	ApiKey       string          `json:"apiKey" DB_COL:"api_key"`
	CreateTime   common.JsonTime `json:"createTime" DB_COL:"create_time"`
	UpdateTime   common.JsonTime `json:"updateTime" DB_COL:"update_time"`
	ReleaseIndex int             `json:"releaseIndex" DB_COL:"release_index"`
}

type QueryAppData struct {
	AppId    int64
	Name     string
	Code     string
	LikeName string
	ApiKey   string
	Start    int64
	End      int64
}

type AppDao struct {
	common.Dao
}

func NewAppDao(db *sql.DB) *AppDao {
	appDao := AppDao{common.Dao{DB: db}}
	return &appDao
}

func (_self *AppDao) QueryApps(queryAppData QueryAppData) []*AppData {
	sql, values := _self.prepareSelectedQuery(false, queryAppData)
	apps := make([]*AppData, 0)
	success, err := _self.StructSelect(&apps, sql, values...)
	if err != nil {
		panic(err)
	}
	if success {
		return apps
	}
	return nil
}

func (_self *AppDao) CountApps(queryAppData QueryAppData) int64 {
	sql, values := _self.prepareSelectedQuery(true, queryAppData)
	return _self.Count(sql, values...)
}

func (_self *AppDao) InsertApp(app *AppData) int64 {
	rowCnt, err := _self.StructInsert(app, false)
	if err != nil {
		panic(err)
	}

	return rowCnt
}

func (_self *AppDao) SelectedUpdateApp(app AppData) int64 {
	sql, values := _self.prepareSelectedUpdate(app)
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}

	return rowCnt
}

func (_self *AppDao) DeleteApp(appId int64) int64 {
	sql := "DELETE FROM `app` WHERE `app_id` = ?"
	rowCnt, err := _self.Exec(sql, appId)
	if err != nil {
		panic(err)
	}

	return rowCnt
}

func (_self *AppDao) ReleaseIndex(appId int64) int {
	apps := _self.QueryApps(QueryAppData{AppId: appId})
	if len(apps) != 1 {
		return -1
	}
	appData := apps[0]

	sql := "UPDATE `app` SET `release_index` = `release_index` + 1 WHERE `app_id` = ? And `release_index` = ?"
	rowCnt, err := _self.Exec(sql, appData.AppId, appData.ReleaseIndex)
	if err != nil {
		panic(err)
	}
	if rowCnt != 1 {
		return -1
	}

	return appData.ReleaseIndex + 1
}

func (_self *AppDao) prepareSelectedQuery(count bool, queryAppData QueryAppData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("SELECT")
	if count {
		buffer.WriteString(" COUNT(1)")
	} else {
		buffer.WriteString(" *")
	}
	buffer.WriteString(" FROM `app` WHERE 1 = 1")

	values := make([]interface{}, 0)
	if queryAppData.AppId > 0 {
		buffer.WriteString(" AND `app_id` = ?")
		values = append(values, queryAppData.AppId)
	}
	if queryAppData.Name != "" {
		buffer.WriteString(" AND `name` = ?")
		values = append(values, queryAppData.Name)
	}
	if queryAppData.LikeName != "" {
		buffer.WriteString(" AND `name` like '")
		buffer.WriteString(queryAppData.LikeName)
		buffer.WriteString("%'")
	}
	if queryAppData.ApiKey != "" {
		buffer.WriteString(" AND `api_key` = ?")
		values = append(values, queryAppData.ApiKey)
	}
	if queryAppData.Start >= 0 && queryAppData.End > 0 {
		buffer.WriteString(" LIMIT ?, ?")
		values = append(values, queryAppData.Start, queryAppData.End)
	}

	return buffer.String(), values
}

func (_self *AppDao) prepareSelectedUpdate(app AppData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("UPDATE `app` SET ")

	values := make([]interface{}, 0)
	if app.Name != "" {
		values = append(values, app.Name)
		buffer.WriteString("`name` = ?,")
	}
	if app.Desc != "" {
		values = append(values, app.Desc)
		buffer.WriteString("`desc` = ?,")
	}
	if app.ApiKey != "" {
		values = append(values, app.ApiKey)
		buffer.WriteString("`api_key` = ?,")
	}
	if !app.CreateTime.IsZero() {
		values = append(values, app.CreateTime)
		buffer.WriteString("`create_time` = ?,")
	}
	if !app.UpdateTime.IsZero() {
		values = append(values, app.UpdateTime)
		buffer.WriteString("`update_time` = ?,")
	}

	sql := strings.TrimSuffix(buffer.String(), ",") + " WHERE `app_id` = ?"
	values = append(values, app.AppId)

	return sql, values
}
