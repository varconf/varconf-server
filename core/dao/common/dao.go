package common

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Dao struct {
	DB *sql.DB
}

const (
	TABLE = "DB_TABLE"
	COL   = "DB_COL"
	PK    = "DB_PK"
)

func (_self *Dao) Insert(sql string, args ...interface{}) (int64, int64, error) {
	stmt, err := _self.DB.Prepare(sql)
	if err != nil {
		return 0, 0, err
	}
	return _self.insertWithStmt(stmt, args)
}

func (_self *Dao) InsertWithTx(tx *sql.Tx, sql string, args ...interface{}) (int64, int64, error) {
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return 0, 0, err
	}
	return _self.insertWithStmt(stmt, args)
}

func (_self *Dao) insertWithStmt(stmt *sql.Stmt, args []interface{}) (int64, int64, error) {
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, 0, err
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	return lastId, rowCnt, err
}

func (_self *Dao) Exec(sql string, args ...interface{}) (int64, error) {
	stmt, err := _self.DB.Prepare(sql)
	if err != nil {
		return 0, err
	}
	return _self.execWithStmt(stmt, args)
}

func (_self *Dao) ExecWithTx(tx *sql.Tx, sql string, args ...interface{}) (int64, error) {
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return 0, err
	}
	return _self.execWithStmt(stmt, args)
}

func (_self *Dao) execWithStmt(stmt *sql.Stmt, args []interface{}) (int64, error) {
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowCnt, err
}

func (_self *Dao) Query(sql string, args ...interface{}) (*sql.Rows, error) {
	return _self.DB.Query(sql, args...)
}

func (_self *Dao) QueryWithTx(tx *sql.Tx, sql string, args ...interface{}) (*sql.Rows, error) {
	return tx.Query(sql, args...)
}

func (_self *Dao) Count(sql string, args ...interface{}) int64 {
	row := _self.DB.QueryRow(sql, args...)
	count := int64(0)
	err := row.Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (_self *Dao) StructInsert(src interface{}, usePK bool) (int64, error) {
	// Parse struct Insert
	pk, mapper, sql, values := _self.structInsertParse(src, usePK)

	// Exec sql
	lastId, rowCnt, err := _self.Insert(sql, values...)
	if err != nil {
		return 0, err
	}
	if pk != "" && !usePK {
		mapper[pk].Set(reflect.ValueOf(lastId))
	}

	return rowCnt, nil
}

func (_self *Dao) StructInsertWithTx(tx *sql.Tx, src interface{}, usePK bool) (int64, error) {
	// Parse struct Insert
	pk, mapper, sql, values := _self.structInsertParse(src, usePK)

	// Exec sql
	lastId, rowCnt, err := _self.InsertWithTx(tx, sql, values...)
	if err != nil {
		return 0, err
	}
	if pk != "" && !usePK {
		mapper[pk].Set(reflect.ValueOf(lastId))
	}

	return rowCnt, nil
}

func (_self *Dao) structInsertParse(src interface{}, usePK bool) (string, map[string]reflect.Value, string, []interface{}) {
	// Get COL Mapper
	table, pk, mapper := _self.structReflect(src)

	// Concat sql string
	var col bytes.Buffer
	var val bytes.Buffer
	values := make([]interface{}, 0)
	for k, v := range mapper {
		if pk == k && !usePK {
			continue
		}
		col.WriteString(fmt.Sprintf("`%s`, ", k))
		val.WriteString("?, ")
		values = append(values, v.Addr().Interface())
	}
	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s) ", table, strings.Trim(col.String(), ", "), strings.Trim(val.String(), ", "))

	return pk, mapper, sql, values
}

func (_self *Dao) StructBatchInsert(usePK bool, beans ...interface{}) (int64, error) {
	if len(beans) < 1 {
		return 0, errors.New("no element")
	}

	// Concat table string
	keys := make([]string, 0)
	var col bytes.Buffer
	table, pk, mapper := _self.structReflect(beans[0])
	for k := range mapper {
		if pk == k && !usePK {
			continue
		}
		col.WriteString(fmt.Sprintf("`%s`, ", k))
		keys = append(keys, k)
	}
	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES ", table, strings.Trim(col.String(), ", "))

	// Concat values string
	values := make([]interface{}, 0)
	for _, bean := range beans {
		_, _, mapper := _self.structReflect(bean)
		var val bytes.Buffer
		for _, key := range keys {
			val.WriteString("?, ")
			values = append(values, mapper[key].Addr().Interface())
		}
		sql = sql + fmt.Sprintf("(%s), ", strings.Trim(val.String(), ", "))
	}
	sql = strings.Trim(sql, ", ")

	// Exec sql
	_, rowCnt, err := _self.Insert(sql, values...)
	if err != nil {
		return 0, err
	}

	return rowCnt, nil
}

func (_self *Dao) StructUpdateByPK(src interface{}) (int64, error) {
	table, pk, mapper := _self.structReflect(src)

	// Concat sql string
	var col bytes.Buffer
	var pv interface{}
	values := make([]interface{}, 0)
	for k, v := range mapper {
		if pk == k {
			pv = v.Addr().Interface()
			continue
		}
		col.WriteString(fmt.Sprintf("`%s` = ?, ", k))
		values = append(values, v.Addr().Interface())
	}
	values = append(values, pv)
	sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s = ?", table, strings.Trim(col.String(), ", "), pk)

	// Exec sql
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		return 0, err
	}

	return rowCnt, nil
}

func (_self *Dao) StructUpsert(src interface{}) (int64, error) {
	// Parse upsert param
	sql, values := _self.structUpsertParse(src)

	// Exec sql
	rowCnt, err := _self.Exec(sql, values...)
	if err != nil {
		return 0, err
	}

	return rowCnt, nil
}

func (_self *Dao) StructUpsertWithTx(tx *sql.Tx, src interface{}) (int64, error) {
	// Parse upsert param
	sql, values := _self.structUpsertParse(src)

	// Exec sql
	rowCnt, err := _self.ExecWithTx(tx, sql, values...)
	if err != nil {
		return 0, err
	}

	return rowCnt, nil
}

func (_self *Dao) structUpsertParse(src interface{}) (string, []interface{}) {
	// Get COL Mapper
	table, _, mapper := _self.structReflect(src)

	// Concat sql string
	var col1 bytes.Buffer
	var col2 bytes.Buffer
	var val bytes.Buffer
	values1 := make([]interface{}, 0)
	values2 := make([]interface{}, 0)
	for k, v := range mapper {
		col1.WriteString(fmt.Sprintf("`%s`, ", k))
		col2.WriteString(fmt.Sprintf("`%s` = ?, ", k))
		val.WriteString("?, ")
		values1 = append(values1, v.Addr().Interface())
		values2 = append(values2, v.Addr().Interface())
	}
	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s", table, strings.Trim(col1.String(), ", "), strings.Trim(val.String(), ", "), strings.Trim(col2.String(), ", "))
	values1 = append(values1, values2...)

	return sql, values1
}

func (_self *Dao) StructSelectByPK(dst interface{}, arg interface{}) (bool, error) {
	table, pk, _ := _self.structReflect(dst)
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE %s = ?", table, pk)

	rows, err := _self.Query(sql, arg)
	if err != nil {
		return false, nil
	}
	defer rows.Close()

	for rows.Next() {
		err := _self.structScan(rows, dst)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (_self *Dao) StructSelectByPKWithTx(tx *sql.Tx, dst interface{}, arg interface{}) (bool, error) {
	table, pk, _ := _self.structReflect(dst)
	sql := fmt.Sprintf("SELECT * FROM `%s` WHERE %s = ?", table, pk)

	rows, err := _self.QueryWithTx(tx, sql, arg)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		err := _self.structScan(rows, dst)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (_self *Dao) StructSelect(dst interface{}, sql string, args ...interface{}) (bool, error) {
	value := reflect.ValueOf(dst)
	if value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Type().Kind() != reflect.Slice {
		return false, errors.New("must pass a none nil slice pointer")
	}

	direct := reflect.Indirect(value)
	slice := reflect.Indirect(value.Elem())
	elemType := reflect.TypeOf(dst).Elem().Elem()
	isPtr := elemType.Kind() == reflect.Ptr

	rows, err := _self.Query(sql, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	// Start scan
	for rows.Next() {
		// New struct
		var elem reflect.Value
		if isPtr {
			elem = reflect.New(elemType.Elem())
		} else {
			elem = reflect.New(elemType)
		}

		// Scan row data
		err := _self.structScan(rows, elem.Interface())
		if err != nil {
			return false, err
		}

		// Append to slice
		if isPtr {
			slice = reflect.Append(slice, elem)
		} else {
			slice = reflect.Append(slice, reflect.Indirect(elem))
		}
	}
	direct.Set(slice)

	return true, nil
}

func (_self *Dao) structScan(rows *sql.Rows, dst interface{}) error {
	// Get COL Mapper
	_, _, mapper := _self.structReflect(dst)

	// Get columns description
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Receive data
	values := make([]interface{}, len(columns))
	for i, column := range columns {
		if v, ok := mapper[column]; ok {
			values[i] = v.Addr().Interface()
		} else {
			values[i] = new(sql.RawBytes)
		}
	}

	// Scan data
	err = rows.Scan(values...)
	if err != nil {
		return err
	}

	return nil
}

func (_self *Dao) structReflect(dst interface{}) (string, string, map[string]reflect.Value) {
	var table, pk string
	mapper := make(map[string]reflect.Value)

	elem := reflect.ValueOf(dst).Elem()
	for i := 0; i < elem.NumField(); i++ {
		if table == "" {
			table = elem.Type().Field(i).Tag.Get(TABLE)
		}
		if pk == "" {
			pk = elem.Type().Field(i).Tag.Get(PK)
		}

		tag := elem.Type().Field(i).Tag.Get(COL)
		if tag != "" {
			mapper[tag] = elem.Field(i)
		}
	}

	return table, pk, mapper
}
