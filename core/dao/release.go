package dao

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"varconf-server/core/dao/common"
)

type ReleaseData struct {
	AppId        int64           `json:"appId" DB_COL:"app_id" DB_PK:"app_id" DB_TABLE:"release"`
	ConfigList   string          `json:"configList" DB_COL:"config_list"`
	ReleaseTime  common.JsonTime `json:"releaseTime" DB_COL:"release_time"`
	ReleaseIndex int             `json:"releaseIndex" DB_COL:"release_index"`
}

type ReleaseDao struct {
	common.Dao
}

func NewReleaseDao(db *sql.DB) *ReleaseDao {
	releaseDao := ReleaseDao{common.Dao{DB: db}}
	return &releaseDao
}

func (_self *ReleaseDao) QueryReleases(appIds []int64) []*ReleaseData {
	values := make([]interface{}, 0)
	sql := "SELECT * FROM `release` WHERE `app_id` in "

	var ids bytes.Buffer
	for _, appId := range appIds {
		ids.WriteString(fmt.Sprintf("%d, ", appId))
	}
	sql = sql + "(" + strings.Trim(ids.String(), ", ") + ")"

	releases := make([]*ReleaseData, 0)
	success, err := _self.StructSelect(&releases, sql, values...)
	if err != nil {
		panic(err)
	}
	if success {
		return releases
	}
	return nil
}

func (_self *ReleaseDao) QueryRelease(appId int64) *ReleaseData {
	release := ReleaseData{}
	success, err := _self.StructSelectByPK(&release, appId)
	if err != nil {
		panic(err)
	}
	if success {
		return &release
	}
	return nil
}

func (_self *ReleaseDao) InsertRelease(data *ReleaseData) int64 {
	rowCnt, err := _self.StructInsert(data, true)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ReleaseDao) UpsertRelease(data *ReleaseData) int64 {
	rowCnt, err := _self.StructUpsert(data)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ReleaseDao) SelectedUpdateRelease(data ReleaseData) int64 {
	sql, values := _self.prepareSelectedUpdate(data)
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ReleaseDao) prepareSelectedUpdate(data ReleaseData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("UPDATE `release` SET ")

	values := make([]interface{}, 0)
	if data.ConfigList != "" {
		values = append(values, data.ConfigList)
		buffer.WriteString("`config_list` = ?,")
	}
	if !data.ReleaseTime.IsZero() {
		values = append(values, data.ReleaseTime)
		buffer.WriteString("`release_time` = ?,")
	}
	if data.ReleaseIndex != 0 {
		values = append(values, data.ReleaseIndex)
		buffer.WriteString("`release_index` = ?,")
	}

	sql := strings.TrimSuffix(buffer.String(), ",") + " WHERE `app_id` = ?"
	values = append(values, data.AppId)

	return sql, values
}
