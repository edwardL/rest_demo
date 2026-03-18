package validator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// split 分割字符串 资产[]内特殊条件
func split(s string, sep byte) []string {
	var fg = false
	var arr []string
	var nowStr []byte
	var sArr = []byte(s)
	for i := 0; i < len(sArr); i++ {
		if sArr[i] == defTagOptions.GroupOptSepLeft {
			fg = true
		}
		if sArr[i] == defTagOptions.GroupOptSepRight {
			fg = false
		}
		if !fg && sArr[i] == sep {
			arr = append(arr, string(nowStr))
			nowStr = make([]byte, 0)
		} else {
			nowStr = append(nowStr, sArr[i])
		}
	}
	if len(nowStr) > 0 {
		arr = append(arr, string(nowStr))
	}
	return arr
}

// getCond 获取 [] 内的条件
func getCond(s string) string {
	var cond []byte
	var start = false
	for i := 0; i < len(s); i++ {
		if !start {
			if s[i] == defTagOptions.GroupOptSepLeft {
				start = true
				continue
			}
			continue
		}
		if s[i] == defTagOptions.GroupOptSepRight {
			break
		}
		cond = append(cond, s[i])
	}
	return string(cond)
}

func toString(v reflect.Value) string {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	fmtStringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	i := v.Interface()
	switch s := i.(type) {
	case string:
		return s
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32)
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case int32:
		return strconv.Itoa(int(s))
	case int16:
		return strconv.FormatInt(int64(s), 10)
	case int8:
		return strconv.FormatInt(int64(s), 10)
	case uint:
		return strconv.FormatUint(uint64(s), 10)
	case uint64:
		return strconv.FormatUint(uint64(s), 10)
	case uint32:
		return strconv.FormatUint(uint64(s), 10)
	case uint16:
		return strconv.FormatUint(uint64(s), 10)
	case uint8:
		return strconv.FormatUint(uint64(s), 10)
	case json.Number:
		return s.String()
	case []byte:
		return string(s)
	case nil:
		return ""
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return fmt.Sprintf("%v", s)
	}
}
