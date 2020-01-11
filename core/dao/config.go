package dao

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"varconf-server/core/dao/common"
)

type ConfigData struct {
	ConfigId    int64            `json:"configId" DB_COL:"config_id" DB_PK:"config_id" DB_TABLE:"config"`
	AppId       int64            `json:"appId" DB_COL:"app_id"`
	Key         string           `json:"key" DB_COL:"key"`
	Value       string           `json:"value" DB_COL:"value"`
	Desc        string           `json:"desc" DB_COL:"desc"`
	Status      int              `json:"status" DB_COL:"status"`
	Operate     int              `json:"operate" DB_COL:"operate"`
	CreateTime  common.JsonTime  `json:"createTime" DB_COL:"create_time"`
	CreateBy    string           `json:"createBy" DB_COL:"create_by"`
	UpdateTime  common.JsonTime  `json:"updateTime" DB_COL:"update_time"`
	UpdateBy    string           `json:"updateBy" DB_COL:"update_by"`
	ReleaseTime *common.JsonTime `json:"releaseTime" DB_COL:"release_time"`
	ReleaseBy   *string          `json:"releaseBy" DB_COL:"release_by"`
}

type QueryConfigData struct {
	ConfigId int64
	AppId    int64
	Status   int
	Key      string
	LikeKey  string
	Start    int64
	End      int64
}

const (
	// 1-未发布、2-已发布
	STATUS_UN = 1
	STATUS_IN = 2
)

const (
	// 1-新增、2-更新、3-删除
	OPERATE_NEW    = 1
	OPERATE_UPDATE = 2
	OPERATE_DELETE = 3
)

type ConfigDao struct {
	common.Dao
}

func NewConfigDao(db *sql.DB) *ConfigDao {
	configDao := ConfigDao{common.Dao{DB: db}}
	return &configDao
}

func (_self *ConfigDao) QueryConfigs(query QueryConfigData) []*ConfigData {
	sql, values := _self.prepareSelectedQuery(false, query)
	configs := make([]*ConfigData, 0)
	success, err := _self.StructSelect(&configs, sql, values...)
	if err != nil {
		panic(err)
	}
	if success {
		return configs
	}
	return nil
}

func (_self *ConfigDao) CountConfigs(query QueryConfigData) int64 {
	sql, values := _self.prepareSelectedQuery(true, query)
	return _self.Count(sql, values...)
}

func (_self *ConfigDao) InsertConfig(data *ConfigData) int64 {
	rowCnt, err := _self.StructInsert(data, false)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ConfigDao) SelectedUpdateConfig(data ConfigData) int64 {
	sql, values := _self.prepareSelectedUpdate(data)
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ConfigDao) DeleteConfig(appId, configId int64) int64 {
	sql := "DELETE FROM `config` WHERE `app_id` = ? AND `config_id` = ?"
	rowCnt, err := _self.Exec(sql, appId, configId)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ConfigDao) BatchUpdateConfig(status int, date time.Time, user string, configIds []int64) int64 {
	values := make([]interface{}, 0)
	sql := "UPDATE `config` SET `status` = ?, `update_time` = ?, `update_by` = ? WHERE `config_id` in "

	values = append(values, status)
	values = append(values, date)
	values = append(values, user)
	var ids bytes.Buffer
	for _, configId := range configIds {
		ids.WriteString(fmt.Sprintf("%d, ", configId))
	}
	sql = sql + "(" + strings.Trim(ids.String(), ", ") + ")"

	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ConfigDao) BatchDeleteConfig(configIds []int64) int64 {
	values := make([]interface{}, 0)
	sql := "DELETE FROM `config` WHERE `config_id` in "

	var ids bytes.Buffer
	for _, configId := range configIds {
		ids.WriteString(fmt.Sprintf("%d, ", configId))
	}
	sql = sql + "(" + strings.Trim(ids.String(), ", ") + ")"

	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		panic(err)
	}
	return rowCnt
}

func (_self *ConfigDao) prepareSelectedQuery(count bool, query QueryConfigData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("SELECT")
	if count {
		buffer.WriteString(" COUNT(1)")
	} else {
		buffer.WriteString(" *")
	}
	buffer.WriteString(" FROM `config` WHERE 1 = 1")

	values := make([]interface{}, 0)
	if query.ConfigId != 0 {
		buffer.WriteString(" AND `config_id` = ?")
		values = append(values, query.ConfigId)
	}
	if query.AppId != 0 {
		buffer.WriteString(" AND `app_id` = ?")
		values = append(values, query.AppId)
	}
	if query.Status != 0 {
		buffer.WriteString(" AND `Status` = ?")
		values = append(values, query.Status)
	}
	if query.Key != "" {
		buffer.WriteString(" AND `key` = ?")
		values = append(values, query.Key)
	}
	if query.LikeKey != "" {
		buffer.WriteString(" AND `key` like '" + query.LikeKey + "%'")
	}
	if query.Start >= 0 && query.End > 0 {
		buffer.WriteString(" LIMIT ?, ?")
		values = append(values, query.Start, query.End)
	}

	return buffer.String(), values
}

func (_self *ConfigDao) prepareSelectedUpdate(data ConfigData) (string, []interface{}) {
	buffer := bytes.Buffer{}
	buffer.WriteString("UPDATE `config` SET ")

	values := make([]interface{}, 0)
	if data.AppId != 0 {
		values = append(values, data.AppId)
		buffer.WriteString("`app_id` = ?,")
	}
	if data.Key != "" {
		values = append(values, data.Key)
		buffer.WriteString("`key` = ?,")
	}
	if data.Value != "" {
		values = append(values, data.Value)
		buffer.WriteString("`value` = ?,")
	}
	if data.Desc != "" {
		values = append(values, data.Desc)
		buffer.WriteString("`desc` = ?,")
	}
	if data.Status != 0 {
		values = append(values, data.Status)
		buffer.WriteString("`status` = ?,")
	}
	if data.Operate != 0 {
		values = append(values, data.Operate)
		buffer.WriteString("`operate` = ?,")
	}
	if !data.CreateTime.IsZero() {
		values = append(values, data.CreateTime)
		buffer.WriteString("`create_time` = ?,")
	}
	if data.CreateBy != "" {
		values = append(values, data.CreateBy)
		buffer.WriteString("`create_by` = ?,")
	}
	if !data.UpdateTime.IsZero() {
		values = append(values, data.UpdateTime)
		buffer.WriteString("`update_time` = ?,")
	}
	if data.UpdateBy != "" {
		values = append(values, data.UpdateBy)
		buffer.WriteString("`update_by` = ?,")
	}
	if data.ReleaseTime != nil {
		values = append(values, data.ReleaseTime)
		buffer.WriteString("`release_time` = ?,")
	}
	if data.ReleaseBy != nil {
		values = append(values, data.ReleaseBy)
		buffer.WriteString("`release_by` = ?,")
	}

	sql := strings.TrimSuffix(buffer.String(), ",") + " WHERE `config_id` = ?"
	values = append(values, data.ConfigId)

	return sql, values
}
