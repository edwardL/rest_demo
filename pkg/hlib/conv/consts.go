package conv

import (
	"encoding/json"
	"reflect"
	"time"
)

var (
	TimeType       = reflect.TypeOf(time.Time{})
	TimePtrType    = reflect.TypeOf(&time.Time{})
	RawMessageType = reflect.TypeOf(json.RawMessage{})
	SliceByteType  = reflect.TypeOf([]byte{})
)
