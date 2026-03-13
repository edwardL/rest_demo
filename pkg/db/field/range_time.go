package field

import (
	"errors"
	"fmt"
	"strings"
	st "time"

	"github.com/opdss/go-helper/time"
	"gorm.io/gorm"
)

type RangeTime struct {
	loc       *st.Location
	rangeTime *time.Range
	TimeField string   `json:"time_field" form:"time_field" query:"time_field" binding:"omitempty"`
	DateTimes []string `json:"range_time" form:"range_time[]" query:"range_time[]" binding:"omitempty"`
}

func (rt *RangeTime) Validate(loc *st.Location) (err error) {
	if rt.rangeTime != nil {
		return nil
	}
	if len(rt.DateTimes) != 2 {
		return errors.New("range_time error")
	}
	loc = rt.getLocal(loc)
	var rtt time.Range
	if loc == nil {
		rtt, err = time.NewRangeFromString(rt.DateTimes[0], rt.DateTimes[1], st.DateTime)
	} else {
		rtt, err = time.NewRangeFromStringInLocation(rt.DateTimes[0], rt.DateTimes[1], st.DateTime, loc)
	}
	if err == nil {
		rt.rangeTime = &rtt
	}
	return
}

// DayQuery YYYY-MM-DD 字符串形式查询
func (rt *RangeTime) DayQuery(loc *st.Location) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if rt.Validate(loc) == nil {
			_st := rt.DateTimes[0][:10]
			_et := rt.DateTimes[1][:10]
			if _st == _et {
				return db.Where(fmt.Sprintf("`%s` = ?", rt.GetField()), _st)
			}
			return db.Where(fmt.Sprintf("`%s` >= ? and `%s` <= ?", rt.GetField(), rt.GetField()), _st, _et)
		}
		return db
	}
}

// DateQuery YYYY-MM-DD 00:00:00 字符串形式查询
func (rt *RangeTime) DateQuery(loc *st.Location) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if rt.Validate(loc) == nil {
			return db.Where(fmt.Sprintf("`%s` >= ? and `%s` <= ?", rt.GetField(), rt.GetField()), fmt.Sprintf("%s 00:00:00", rt.DateTimes[0][:10]), fmt.Sprintf("%s 23:59:59", rt.DateTimes[1][:10]))
		}
		return db
	}
}

// DateTimeQuery YYYY-MM-DD HH:MM:SS 字符串形式查询
func (rt *RangeTime) DateTimeQuery(loc *st.Location) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if rt.Validate(loc) == nil {
			if rt.DateTimes[0] == rt.DateTimes[1] {
				return db.Where(fmt.Sprintf("`%s` = ? ", rt.GetField()), rt.DateTimes[0])
			}
			return db.Where(fmt.Sprintf("`%s` >= ? and `%s` <= ?", rt.GetField(), rt.GetField()), rt.DateTimes[0], rt.DateTimes[1])
		}
		return db
	}
}

// TimestampQuery int 字段时间戳形式查询
func (rt *RangeTime) TimestampQuery(loc *st.Location) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if rt.Validate(loc) == nil {
			_st := rt.rangeTime.StartTime.Unix()
			_et := rt.rangeTime.EndTime.Unix()
			if _st == _et {
				return db.Where(fmt.Sprintf("`%s` = ?", rt.GetField()), _st)
			} else {
				return db.Where(fmt.Sprintf("`%s` >= ? and `%s` <= ?", rt.GetField(), rt.GetField()), _st, _et)
			}
		}
		return db
	}
}

// DateTimestampQuery int 字段时间戳形式查询, 去掉时分秒取整日的时间戳
func (rt *RangeTime) DateTimestampQuery(loc *st.Location) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if rt.Validate(loc) == nil {
			_st, _ := st.ParseInLocation(st.DateTime, rt.rangeTime.StartTime.Format(st.DateOnly)+" 00:00:00", rt.getLocal(loc))
			_et, _ := st.ParseInLocation(st.DateTime, rt.rangeTime.EndTime.Format(st.DateOnly)+" 23:59:59", rt.getLocal(loc))
			return db.Where(fmt.Sprintf("`%s` >= ? and `%s` < ?", rt.GetField(), rt.GetField()),
				_st.Unix(), _et.Unix())
		}
		return db
	}
}

func (rt *RangeTime) GetField() string {
	if rt.TimeField == "" {
		rt.TimeField = "created_at"
	}
	return rt.TimeField
}

func (rt *RangeTime) SetLocation(l *st.Location) *RangeTime {
	rt.loc = l
	return rt
}

func (rt *RangeTime) SetField(f string) *RangeTime {
	rt.TimeField = strings.Trim(f, "`")
	return rt
}

func (rt *RangeTime) GetStartTime() st.Time {
	if rt.rangeTime != nil {
		return rt.rangeTime.StartTime
	}
	return st.Time{}
}

func (rt *RangeTime) GetEndTime() st.Time {
	if rt.rangeTime != nil {
		return rt.rangeTime.EndTime
	}
	return st.Time{}
}

func (rt *RangeTime) getLocal(loc *st.Location) *st.Location {
	if loc != nil {
		return loc
	}
	if rt.loc != nil {
		return rt.loc
	}
	return st.Local
}
