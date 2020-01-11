package common

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type JsonTime struct {
	time.Time
}

const (
	timeFormat = "2006-01-02 15:04:05"
)

func NowJsonTime() JsonTime {
	return JsonTime{Time: time.Now()}
}

func (t *JsonTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = JsonTime{Time: now}
	return
}

func (t JsonTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t.Time).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

func (t JsonTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *JsonTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JsonTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
