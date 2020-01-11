package dao

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"varconf-server/core/dao/common"
)

type ManageTxDao struct {
	common.Dao
}

func NewManageTxDao(db *sql.DB) *ManageTxDao {
	manageTxDao := ManageTxDao{common.Dao{DB: db}}
	return &manageTxDao
}

func (_self *ManageTxDao) ReleaseConfig(appId int64, allConfigs []*ConfigData, user string) (bool, []string) {
	// start tx
	tx, err := _self.DB.Begin()
	if err != nil {
		return false, nil
	}
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()

	// parse allConfigs and update config status
	releaseConfigs, keys, err := _self.batchReleaseConfigTx(tx, allConfigs, user)
	if err != nil {
		return false, nil
	}

	// encode json
	configList, err := json.Marshal(releaseConfigs)
	if err != nil {
		return false, nil
	}

	// get release index
	releaseIndex, err := _self.releaseIndexTx(tx, appId)
	if releaseIndex < 0 || err != nil {
		return false, nil
	}

	// upsert release data
	releaseData := &ReleaseData{
		AppId:        appId,
		ConfigList:   string(configList),
		ReleaseTime:  common.NowJsonTime(),
		ReleaseIndex: releaseIndex,
	}
	_, err = _self.upsertReleaseTx(tx, releaseData)
	if err != nil {
		return false, nil
	}

	// upsert release log
	releaseLogData := &ReleaseLogData{
		AppId:        releaseData.AppId,
		ConfigList:   releaseData.ConfigList,
		ReleaseTime:  releaseData.ReleaseTime,
		ReleaseIndex: releaseData.ReleaseIndex,
		ReleaseBy:    user,
	}
	_, err = _self.insertReleaseLogTx(tx, releaseLogData)
	if err != nil {
		return false, nil
	}

	// commit tx
	err = tx.Commit()
	if err != nil {
		return false, nil
	}
	return true, keys
}

func (_self *ManageTxDao) DeleteApp(appId int64) bool {
	// start tx
	tx, err := _self.DB.Begin()
	if err != nil {
		return false
	}
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()

	// delete app's data
	sql := "DELETE FROM `app` WHERE `app_id` = ?"
	_, err = _self.ExecWithTx(tx, sql, appId)
	if err != nil {
		return false
	}
	sql = "DELETE FROM `config` WHERE `app_id` = ?"
	_, err = _self.ExecWithTx(tx, sql, appId)
	if err != nil {
		return false
	}
	sql = "DELETE FROM `release` WHERE `app_id` = ?"
	_, err = _self.ExecWithTx(tx, sql, appId)
	if err != nil {
		return false
	}
	sql = "DELETE FROM `release_log` WHERE `app_id` = ?"
	_, err = _self.ExecWithTx(tx, sql, appId)
	if err != nil {
		return false
	}

	// commit tx
	err = tx.Commit()
	if err != nil {
		return false
	}
	return true
}

func (_self *ManageTxDao) batchReleaseConfigTx(tx *sql.Tx, configs []*ConfigData, user string) ([]*ConfigData, []string, error) {
	if len(configs) < 1 {
		return nil, nil, errors.New("configs is empty")
	}

	// parse data
	now := common.NowJsonTime()
	releaseConfigs := make([]*ConfigData, 0)
	keys := make([]string, 0)
	updateIds := make([]int64, 0)
	deleteIds := make([]int64, 0)
	for _, config := range configs {
		keys = append(keys, config.Key)
		if config.Status == STATUS_UN {
			config.ReleaseTime = &now
			config.ReleaseBy = &user

			updateIds = append(updateIds, config.ConfigId)
			if config.Operate == OPERATE_DELETE {
				deleteIds = append(deleteIds, config.ConfigId)
				continue
			}
		}
		releaseConfigs = append(releaseConfigs, config)
	}

	// update data
	var err error
	if len(updateIds) > 0 {
		_, err = _self.batchUpdateConfigTx(tx, STATUS_IN, now.Time, user, updateIds)
	}
	if len(deleteIds) > 0 {
		_, err = _self.batchDeleteConfigTx(tx, deleteIds)
	}
	if err != nil {
		return nil, keys, err
	}

	return releaseConfigs, keys, nil
}

func (_self *ManageTxDao) batchUpdateConfigTx(tx *sql.Tx, status int, date time.Time, user string, configIds []int64) (int64, error) {
	values := make([]interface{}, 0)
	sql := "UPDATE `config` SET `status` = ?, `release_time` = ?, `release_by` = ? WHERE `config_id` in "

	values = append(values, status)
	values = append(values, date)
	values = append(values, user)
	var ids bytes.Buffer
	for _, configId := range configIds {
		ids.WriteString(fmt.Sprintf("%d, ", configId))
	}
	sql = sql + "(" + strings.Trim(ids.String(), ", ") + ")"

	return _self.Exec(sql, values...)
}

func (_self *ManageTxDao) batchDeleteConfigTx(tx *sql.Tx, configIds []int64) (int64, error) {
	values := make([]interface{}, 0)
	sql := "DELETE FROM `config` WHERE `config_id` in "

	var ids bytes.Buffer
	for _, configId := range configIds {
		ids.WriteString(fmt.Sprintf("%d, ", configId))
	}
	sql = sql + "(" + strings.Trim(ids.String(), ", ") + ")"

	return _self.Exec(sql, values...)
}

func (_self *ManageTxDao) releaseIndexTx(tx *sql.Tx, appId int64) (int, error) {
	appData := AppData{}
	success, err := _self.StructSelectByPKWithTx(tx, &appData, appId)
	if !success || err != nil {
		return -1, err
	}

	sql := "UPDATE `app` SET `release_index` = `release_index` + 1 WHERE `app_id` = ? And `release_index` = ?"
	rowCnt, err := _self.ExecWithTx(tx, sql, appId, appData.ReleaseIndex)
	if rowCnt != 1 || err != nil {
		return -1, err
	}
	return appData.ReleaseIndex + 1, nil
}

func (_self *ManageTxDao) upsertReleaseTx(tx *sql.Tx, data *ReleaseData) (int64, error) {
	return _self.StructUpsertWithTx(tx, data)
}

func (_self *ManageTxDao) insertReleaseLogTx(tx *sql.Tx, data *ReleaseLogData) (int64, error) {
	return _self.StructInsertWithTx(tx, data, false)
}
