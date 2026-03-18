package conv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ValType 支持的值类型
type ValType reflect.Kind

const (
	// TypeBool 布尔类型
	TypeBool ValType = ValType(reflect.Bool)
	// TypeInt 有符号整数类型
	TypeInt   ValType = ValType(reflect.Int)
	TypeInt8  ValType = ValType(reflect.Int8)
	TypeInt16 ValType = ValType(reflect.Int16)
	TypeInt32 ValType = ValType(reflect.Int32)
	TypeInt64 ValType = ValType(reflect.Int64)
	// TypeUint 无符号整数类型
	TypeUint   ValType = ValType(reflect.Uint)
	TypeUint8  ValType = ValType(reflect.Uint8)
	TypeUint16 ValType = ValType(reflect.Uint16)
	TypeUint32 ValType = ValType(reflect.Uint32)
	TypeUint64 ValType = ValType(reflect.Uint64)
	// TypeFloat32 浮点类型
	TypeFloat32 ValType = ValType(reflect.Float32)
	TypeFloat64 ValType = ValType(reflect.Float64)
	// TypeString 字符串类型
	TypeString ValType = ValType(reflect.String)
)

type float64EProvider interface {
	Float64() (float64, error)
}

type float64Provider interface {
	Float64() float64
}

// ToString 将任意类型转换为字符串。
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
func ToTime(v any, timeLocation ...*time.Location) (time.Time, error) {
	var tv = time.Time{}
	var vt = reflect.ValueOf(v)
	if IsNilValue(vt) || isZeroValue(v) {
		return tv, nil
	}
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	v = vt.Interface()
	if v == nil {
		return tv, nil
	}
	if t, ok := v.(time.Time); ok {
		return t, nil
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
	var timeLoc = time.Local
	if len(timeLocation) > 0 {
		timeLoc = timeLocation[0]
	}
	return parseFormattedTimeString(vStr, timeLoc)
}

// parseTimestamp 解析各种时间戳
func parseTimestamp(tm int64) (time.Time, error) {
	if tm > 1e18 {
		// 纳秒级时间戳
		return time.Unix(0, tm), nil
	} else if tm > 1e15 {
		// 微秒级时间戳
		return time.UnixMicro(tm), nil
	} else if tm > 1e12 {
		// 毫秒级时间戳
		return time.UnixMilli(tm), nil
	}
	// 秒时间戳
	return time.Unix(tm, 0), nil
}

// parseFormattedTimeString 解析格式化的时间字符串
func parseFormattedTimeString(tFmt string, timeLocation *time.Location) (time.Time, error) {
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
		if t, err = time.ParseInLocation(format, tFmt, timeLocation); err == nil {
			return t, nil
		} else {
			errMsg.WriteString(err.Error())
		}
	}
	return time.Time{}, errors.New(errMsg.String())
}

// ToPtr 创建指针
func ToPtr[T any](v T) *T {
	return &v
}

// ParseValToStrSlice 解析 1,2,3 或 [1,2,3] 为切片
func ParseValToStrSlice(val string) []string {
	val = strings.TrimSpace(val)
	if val == "" {
		return []string{}
	}
	// 处理 [1,2,3] 或 [aa,bb,cc] 格式
	if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
		// 去除前后 []，再按逗号分割
		var content = strings.TrimSuffix(strings.TrimPrefix(val, "["), "]")
		return SplitTrim(content, ",")
	}
	// 处理 1,2,3,aa,vv 格式
	return SplitTrim(val, ",")
}

// SplitTrim 按分隔符分割字符串，并去除每个元素的前后空格
func SplitTrim(s, sep string) []string {
	var parts = strings.Split(s, sep)
	var res []string
	var trimmed string
	for _, part := range parts {
		trimmed = strings.TrimSpace(part)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

// ToValType 将任意类型的 any 值转换为 ValType 指定类型的 any 值
func ToValType(input any, targetType ValType) any {
	// 处理 nil 输入 结果累的都不支持nil
	if input == nil {
		input = ""
	}

	// 获取输入值的反射值（处理指针解引用）
	var inputVal = reflect.ValueOf(input)
	for inputVal.Kind() == reflect.Ptr {
		inputVal = inputVal.Elem() // 解引用指针，获取底层值
	}
	var inputKind = inputVal.Kind()

	// 如果输入类型已经是目标类型，直接返回
	if ValType(inputKind) == targetType {
		return inputVal.Interface()
	}

	// 核心转换逻辑：按目标类型分支处理
	switch targetType {
	case TypeString:
		// 任意类型转字符串
		return ToString(input)
	case TypeBool:
		var b, _ = ToBool(input)
		return b

	case TypeInt:
		// 其他类型转 int
		var i, _ = ToInt(input)
		return i

	case TypeInt8:
		var i, _ = ToInt(input)
		return int8(i)

	case TypeInt16:
		var i, _ = ToInt(input)
		return int16(i)

	case TypeInt32:
		var i, _ = ToInt32(input)
		return i

	case TypeInt64:
		var i, _ = ToInt64(input)
		return i

	case TypeUint:
		var i, _ = ToInt(input)
		return uint(i)

	case TypeUint8:
		var i, _ = ToInt(input)
		return uint8(i)

	case TypeUint16:
		var i, _ = ToInt(input)
		return uint16(i)

	case TypeUint32:
		var i, _ = ToInt32(input)
		return uint32(i)

	case TypeUint64:
		var i, _ = ToInt64(input)
		return uint64(i)

	case TypeFloat32:
		var i, _ = ToFloat64E(input)
		return float32(i)

	case TypeFloat64:
		var i, _ = ToFloat64E(input)
		return i

	default:
		return input
	}
}

// SliceToSlice 将[][]string第一个字段转换为切片
func SliceToSlice(sList [][]string, s any) (err error) {
	// 边界：sList 长度不足，直接返回
	if len(sList) < 2 {
		return nil
	}

	var origVal = reflect.ValueOf(s)
	// 校验入参是指针
	if origVal.Kind() != reflect.Ptr {
		return fmt.Errorf("入参必须是指针类型，当前类型：%s", origVal.Kind())
	}
	// 解引用
	var sliceVal = origVal
	for sliceVal.Kind() == reflect.Ptr {
		sliceVal = sliceVal.Elem()
	}
	// 校验最终指向切片
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("指针最终必须指向切片类型，当前指向：%s", sliceVal.Kind())
	}

	// 解析切片元素类型（剥离多层指针，获取基础类型和指针层数）
	var elemType = sliceVal.Type().Elem()
	var baseElemType = elemType
	var ptrLayers = 0
	// 统计指针层数 + 提取基础类型（如 **string → string，ptrLayers=2）
	for baseElemType.Kind() == reflect.Ptr {
		ptrLayers++
		baseElemType = baseElemType.Elem()
	}
	var baseElemKind = baseElemType.Kind()

	// 遍历 sList 转换值并适配多层指针
	var newSlice = reflect.MakeSlice(sliceVal.Type(), 0, len(sList)-1)
	var sliceValues []reflect.Value

	for _, v := range sList[1:] {
		// 防护：sList 某行是空切片
		if len(v) == 0 {
			return errors.New("sList 中存在空行，无法转换")
		}
		var strVal = v[0]
		// 转换为基础类型值
		var baseVal reflect.Value
		switch baseElemKind {
		case reflect.String:
			baseVal = reflect.ValueOf(strVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, e := ToInt64(strVal)
			if e != nil {
				return fmt.Errorf("转换int失败：%v，值：%s", e, strVal)
			}
			baseVal = reflect.ValueOf(intVal).Convert(baseElemType)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			uintVal, e := ToUint64(strVal)
			if e != nil {
				return fmt.Errorf("转换uint失败：%v，值：%s", e, strVal)
			}
			baseVal = reflect.ValueOf(uintVal).Convert(baseElemType)
		case reflect.Float32, reflect.Float64:
			floatVal, e := ToFloat64E(strVal)
			if e != nil {
				return fmt.Errorf("转换float失败：%v，值：%s", e, strVal)
			}
			baseVal = reflect.ValueOf(floatVal).Convert(baseElemType)
		case reflect.Bool:
			boolVal, e := ToBool(strVal)
			if e != nil {
				return fmt.Errorf("转换bool失败：%v，值：%s", e, strVal)
			}
			baseVal = reflect.ValueOf(boolVal)
		default:
			return fmt.Errorf("不支持的基础元素类型：%s", baseElemKind)
		}

		// 适配多层指针：递归创建指针（如 ptrLayers=2 → **string）
		var currentVal = baseVal
		for i := 0; i < ptrLayers; i++ {
			var ptr = reflect.New(currentVal.Type())
			ptr.Elem().Set(currentVal)
			currentVal = ptr
		}
		sliceValues = append(sliceValues, currentVal)
	}

	// 赋值到原始多层指针（自动解引用找到切片并赋值）
	newSlice = reflect.Append(newSlice, sliceValues...)
	// 逐层解引用原始指针，找到切片后赋值
	var tempVal = origVal
	for {
		if tempVal.Kind() != reflect.Ptr {
			break
		}
		elem := tempVal.Elem()
		if elem.Kind() == reflect.Slice {
			elem.Set(newSlice)
			break
		}
		tempVal = elem
	}

	return nil
}

// StrToTargetType 转换为目标反射类型
func StrToTargetType(strVal string, targetType reflect.Type) (reflect.Value, error) {
	if targetType.Kind() == reflect.String {
		return reflect.ValueOf(strVal), nil
	}
	var i64, _ = ToInt64(strVal)
	switch targetType.Kind() {
	// 有符号整数
	case reflect.Int:
		return reflect.ValueOf(int(i64)), nil
	case reflect.Int8:
		return reflect.ValueOf(int8(i64)), nil
	case reflect.Int16:
		return reflect.ValueOf(int16(i64)), nil
	case reflect.Int32:
		return reflect.ValueOf(int32(i64)), nil
	case reflect.Int64:
		return reflect.ValueOf(i64), nil
	// 无符号整数
	case reflect.Uint:
		return reflect.ValueOf(uint(i64)), nil
	case reflect.Uint8:
		return reflect.ValueOf(uint8(i64)), nil
	case reflect.Uint16:
		return reflect.ValueOf(uint16(i64)), nil
	case reflect.Uint32:
		return reflect.ValueOf(uint32(i64)), nil
	case reflect.Uint64:
		return reflect.ValueOf(uint64(i64)), nil
	case reflect.Uintptr:
		return reflect.ValueOf(uintptr(i64)), nil
	default:
		return reflect.Value{}, errors.New("conv target type error")
	}
}
