package field

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

type LocalTime time.Time

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	tTime := time.Time(*t)
	return []byte(fmt.Sprintf("\"%v\"", tTime.Format(time.DateTime))), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	if str == "" {
		return errors.New("time is error")
	}
	ts, err := time.Parse(time.DateTime, str)
	if err != nil {
		return err
	}
	*t = LocalTime(ts)
	return nil
}

func (t LocalTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)
	//判断给定时间是否和默认零时间的时间戳相同
	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if v == nil {
		*t = LocalTime(time.Time{})
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		*t = LocalTime(val)
		return nil
	case []byte:
		_t, err := time.Parse(time.DateTime, string(val))
		if err != nil {
			return err
		}
		*t = LocalTime(_t)
		return nil
	case string:
		_t, err := time.Parse(time.DateTime, val)
		if err != nil {
			return err
		}
		*t = LocalTime(_t)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
