package gdbtmp

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type float64EProvider interface {
	Float64() (float64, error)
}

type float64Provider interface {
	Float64() float64
}

func ToString(i any) string {
	i = indirectToStringerOrError(i)

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

// ToInt64 casts an interface to an int64 type.
func ToInt64(i any) (int64, error) {
	i = indirect(i)

	intv, ok := toInt(i)
	if ok {
		return int64(intv), nil
	}

	switch s := i.(type) {
	case int64:
		return s, nil
	case int32:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case uint:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	case float64:
		return int64(s), nil
	case float32:
		return int64(s), nil
	case string:
		v, err := strconv.ParseInt(trimZeroDecimal(s), 0, 0)
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", i, i)
	case json.Number:
		return ToInt64(string(s))
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", i, i)
	}
}

// ToUint64 将任意类型转换为uint64
func ToUint64(v any) (uint64, error) {
	if v == nil {
		return 0, nil
	}

	var val = reflect.ValueOf(v)

	// 解引用指针
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0, nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Uint:
		return uint64(val.Uint()), nil
	case reflect.Uint8:
		return uint64(val.Uint()), nil
	case reflect.Uint16:
		return uint64(val.Uint()), nil
	case reflect.Uint32:
		return uint64(val.Uint()), nil
	case reflect.Uint64:
		return val.Uint(), nil
	case reflect.Uintptr:
		return uint64(val.Uint()), nil
	case reflect.Int:
		i := val.Int()
		if i < 0 {
			return 0, errors.New("cannot convert negative int to uint64")
		}
		return uint64(i), nil
	case reflect.Int8:
		i := val.Int()
		if i < 0 {
			return 0, errors.New("cannot convert negative int8 to uint64")
		}
		return uint64(i), nil
	case reflect.Int16:
		i := val.Int()
		if i < 0 {
			return 0, errors.New("cannot convert negative int16 to uint64")
		}
		return uint64(i), nil
	case reflect.Int32:
		i := val.Int()
		if i < 0 {
			return 0, errors.New("cannot convert negative int32 to uint64")
		}
		return uint64(i), nil
	case reflect.Int64:
		i := val.Int()
		if i < 0 {
			return 0, errors.New("cannot convert negative int64 to uint64")
		}
		return uint64(i), nil
	case reflect.Float32, reflect.Float64:
		f := val.Float()
		if f < 0 {
			return 0, errors.New("cannot convert negative float to uint64")
		}
		if f > float64(uint64(^uint64(0))) {
			return 0, errors.New("value exceeds uint64 maximum limit")
		}
		return uint64(f), nil
	case reflect.String:
		num, err := strconv.ParseUint(val.String(), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid string for uint64 conversion: %w", err)
		}
		return num, nil
	case reflect.Bool:
		if val.Bool() {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type %s for uint64 conversion", val.Kind())
	}
}

// ToInt32 casts an interface to an int32 type.
func ToInt32(i any) (int32, error) {
	i = indirect(i)

	intv, ok := toInt(i)
	if ok {
		return int32(intv), nil
	}

	switch s := i.(type) {
	case int64:
		return int32(s), nil
	case int32:
		return s, nil
	case int16:
		return int32(s), nil
	case int8:
		return int32(s), nil
	case uint:
		return int32(s), nil
	case uint64:
		return int32(s), nil
	case uint32:
		return int32(s), nil
	case uint16:
		return int32(s), nil
	case uint8:
		return int32(s), nil
	case float64:
		return int32(s), nil
	case float32:
		return int32(s), nil
	case string:
		v, err := strconv.ParseInt(trimZeroDecimal(s), 0, 0)
		if err == nil {
			return int32(v), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", i, i)
	case json.Number:
		return ToInt32(string(s))
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", i, i)
	}
}

// ToInt casts an interface to an int type.
func ToInt(i any) (int, error) {
	i = indirect(i)

	intv, ok := toInt(i)
	if ok {
		return intv, nil
	}

	switch s := i.(type) {
	case int64:
		return int(s), nil
	case int32:
		return int(s), nil
	case int16:
		return int(s), nil
	case int8:
		return int(s), nil
	case uint:
		return int(s), nil
	case uint64:
		return int(s), nil
	case uint32:
		return int(s), nil
	case uint16:
		return int(s), nil
	case uint8:
		return int(s), nil
	case float64:
		return int(s), nil
	case float32:
		return int(s), nil
	case string:
		v, err := strconv.ParseInt(trimZeroDecimal(s), 0, 0)
		if err == nil {
			return int(v), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", i, i)
	case json.Number:
		return ToInt(string(s))
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", i, i)
	}
}

// ToBool 将任意类型转换为bool
func ToBool(v any) (bool, error) {
	if v == nil {
		return false, nil
	}

	var val = reflect.ValueOf(v)

	// 解引用指针
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false, nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Bool:
		return val.Bool(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num := val.Int()
		// 0为false，非0为true
		return num != 0, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		num := val.Uint()
		// 0为false，非0为true
		return num != 0, nil

	case reflect.Float32, reflect.Float64:
		num := val.Float()
		// 0.0为false，非0为true
		return num != 0, nil

	case reflect.String:
		str := strings.TrimSpace(val.String())
		// 支持常见的布尔字符串表示
		switch strings.ToLower(str) {
		case "true", "t", "1", "yes", "y":
			return true, nil
		case "false", "f", "0", "no", "n":
			return false, nil
		default:
			return false, fmt.Errorf("invalid string for bool conversion: %q", str)
		}
	default:
		return false, fmt.Errorf("unsupported type %s for bool conversion", val.Kind())
	}
}

func toInt(v any) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case time.Weekday:
		return int(v), true
	case time.Month:
		return int(v), true
	default:
		return 0, false
	}
}

func indirect(a any) any {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

func indirectToStringerOrError(a any) any {
	if a == nil {
		return nil
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	fmtStringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	v := reflect.ValueOf(a)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

func trimZeroDecimal(s string) string {
	var foundZero bool
	for i := len(s); i > 0; i-- {
		switch s[i-1] {
		case '.':
			if foundZero {
				return s[:i-1]
			}
		case '0':
			foundZero = true
		default:
			return s
		}
	}
	return s
}

// ToFloat64E casts an interface to a float64 type.
func ToFloat64E(i any) (float64, error) {
	i = indirect(i)

	intv, ok := toInt(i)
	if ok {
		return float64(intv), nil
	}

	switch s := i.(type) {
	case float64:
		return s, nil
	case float32:
		return float64(s), nil
	case int64:
		return float64(s), nil
	case int32:
		return float64(s), nil
	case int16:
		return float64(s), nil
	case int8:
		return float64(s), nil
	case uint:
		return float64(s), nil
	case uint64:
		return float64(s), nil
	case uint32:
		return float64(s), nil
	case uint16:
		return float64(s), nil
	case uint8:
		return float64(s), nil
	case string:
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", i, i)
	case float64EProvider:
		v, err := s.Float64()
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", i, i)
	case float64Provider:
		return s.Float64(), nil
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", i, i)
	}
}

// ToTime 时间转换
func ToTime(v any) (time.Time, error) {
	var tv = time.Time{}
	if v == nil {
		return tv, nil
	}
	if t, ok := v.(time.Time); ok {
		return t, nil
	}
	if t, ok := v.(*time.Time); ok {
		if t == nil {
			return tv, nil
		}
		return *t, nil
	}
	var vStr = ToString(v)
	if vStr == "" {
		return tv, errors.New("param is empty")
	}
	if isNum(vStr) {
		var vInt, err = ToInt64(vStr)
		if err != nil {
			return tv, err
		}
		return parseTimestamp(vInt)
	}
	return parseFormattedTimeString(vStr)
}

// parseTimestamp 解析各种时间戳
func parseTimestamp(tm int64) (time.Time, error) {
	if tm > 1e12 {
		// 毫秒级时间戳
		return time.UnixMilli(tm), nil
	}
	// 秒时间戳
	return time.Unix(tm, 0), nil
}

// parseFormattedTimeString 解析格式化的时间字符串
func parseFormattedTimeString(tFmt string) (time.Time, error) {
	// 常见的时间格式列表，按优先级排序
	var timeFormats = []string{
		time.DateTime,
		"2006年01月02日",
		"2006年01月02日 15:04:05",
		time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.DateOnly,
		time.TimeOnly,
	}
	var err error
	var errMsg = strings.Builder{}
	var t time.Time
	for _, format := range timeFormats {
		if t, err = time.Parse(format, tFmt); err == nil {
			return t, nil
		} else {
			errMsg.WriteString(err.Error())
		}
	}
	return time.Time{}, errors.New(errMsg.String())
}

// ToPtr 创建指针
func ToPtr[T uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64 | string | time.Time](v T) *T {
	return &v
}

// MapToMapAny 任意map[string]T类型转 map[string]any
func MapToMapAny(d any) (map[string]any, error) {
	var m map[string]any
	switch dt := d.(type) {
	case map[string]any:
		return dt, nil
	case *map[string]any:
		return *dt, nil
	case map[string]RawBody:
		for k, v := range dt {
			m[k] = v
		}
		return m, nil
	case map[string]string:
		for k, v := range dt {
			m[k] = v
		}
		return m, nil
	}
	// 无匹配 使用反射统一处理所有 map[string]T 类型
	var v = reflect.ValueOf(d)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map && v.Type().Key().Kind() == reflect.String {
		// 检查是否为 map[string]T 格式
		var result = make(map[string]any, v.Len())
		var iter = v.MapRange()
		var key string
		var value any
		for iter.Next() {
			key = iter.Key().String()
			value = iter.Value().Interface()
			result[key] = value
		}
		return result, nil
	}
	return nil, errors.New("not a map[string]T")
}

// MapsToMapsAny 任意[]map[string]T类型转 []map[string]any
func MapsToMapsAny(d any) ([]map[string]any, error) {
	var ml = make([]map[string]any, 0)
	if d == nil {
		return ml, nil
	}
	// 处理常用类型
	switch dt := d.(type) {
	case []map[string]any:
		return dt, nil
	case *[]map[string]any:
		if dt == nil {
			return ml, nil
		}
		return *dt, nil
	case []map[string]RawBody:
		var result = make([]map[string]any, len(dt))
		for i, m := range dt {
			resultMap := make(map[string]any, len(m))
			for k, v := range m {
				resultMap[k] = v
			}
			result[i] = resultMap
		}
		return result, nil
	case []map[string]string:
		var result = make([]map[string]any, len(dt))
		for i, m := range dt {
			resultMap := make(map[string]any, len(m))
			for k, v := range m {
				resultMap[k] = v
			}
			result[i] = resultMap
		}
		return result, nil
	}
	// 使用反射处理所有 []map[string]T 类型
	var v = reflect.ValueOf(d)
	// 解引用指针
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ml, nil
		}
		v = v.Elem()
	}

	// 检查是否为切片类型
	if v.Kind() == reflect.Slice && v.Len() > 0 {
		var ft = v.Index(0)
		if ft.Kind() == reflect.Map && ft.Type().Key().Kind() == reflect.String {
			var result = make([]map[string]any, v.Len())
			var key string
			// 遍历切片中的每个元素
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i)
				// 检查元素是否为 map[string]T 类型
				if item.Kind() == reflect.Map && item.Type().Key().Kind() == reflect.String {
					// 转换单个 map
					resultMap := make(map[string]any, item.Len())
					iter := item.MapRange()
					for iter.Next() {
						key = iter.Key().String()
						value := iter.Value().Interface()
						resultMap[key] = value
					}
					result[i] = resultMap
				} else {
					break
				}
			}
			return result, nil
		}
	}
	return nil, errors.New("not a []map[string]T")
}

// isNum 是否数字
func isNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
