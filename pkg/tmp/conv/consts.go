package conv

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"time"
)

const (
	TagTypeKey           = "type"      // 模型标签类型标识
	TagTypeTimeUnixMilli = "UnixMilli" // 毫秒时间戳
	TagTypeTimeUnix      = "Unix"      // 秒时间戳
)

var (
	TimeType        = reflect.TypeOf(time.Time{})
	TimePtrType     = reflect.TypeOf(&time.Time{})
	RawMessageType  = reflect.TypeOf(json.RawMessage{})
	SliceByteType   = reflect.TypeOf([]byte{})
	NotStructErrMsg = "value must be a struct or pointer to struct"

	// fieldScanFaceType sql.Scanner 接口的反射类型
	fieldScanFaceType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	// fieldValueFaceType driver.Valuer 接口的反射类型
	fieldValueFaceType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)
