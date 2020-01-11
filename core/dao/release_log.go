package dao

import (
	"database/sql"

	"varconf-server/core/dao/common"
)

// 配置历史
type ReleaseLogData struct {
	Id           int64           `json:"id" DB_COL:"id" DB_PK:"id" DB_TABLE:"release_log"`
	AppId        int64           `json:"appId" DB_COL:"app_id"`
	ConfigList   string          `json:"configList" DB_COL:"config_list"`
	ReleaseTime  common.JsonTime `json:"releaseTime" DB_COL:"release_time"`
	ReleaseIndex int             `json:"releaseIndex" DB_COL:"release_index"`
	ReleaseBy    string          `json:"releaseBy" DB_COL:"release_by"`
}

type QueryReleaseLogData struct {
	AppId            int64
	LessReleaseIndex int
	Start            int64
	End              int64
}

type ReleaseLogDao struct {
	common.Dao
}

func NewReleaseLogDao(db *sql.DB) *ReleaseLogDao {
	releaseLogDao := ReleaseLogDao{common.Dao{DB: db}}
	return &releaseLogDao
}

func (_self *ReleaseLogDao) QueryReleaseLogs(appId int64) []*ReleaseLogData {
	sql := "SELECT * FROM `release_log` WHERE `app_id` = ?"

	releaseLogs := make([]*ReleaseLogData, 0)
	success, err := _self.StructSelect(&releaseLogs, sql, appId)
	if err != nil {
		panic(err)
	}
	if success {
		return releaseLogs
	}
	return nil
}

func (_self *ReleaseLogDao) CountReleaseLogs() int64 {
	sql := "SELECT count(1) FROM `release_log`"
	return _self.Count(sql)
}

func (_self *ReleaseLogDao) InsertReleaseLog(data *ReleaseLogData) int64 {
	rowCnt, err := _self.StructInsert(data, false)
	if err != nil {
		panic(err)
	}
	return rowCnt
}
