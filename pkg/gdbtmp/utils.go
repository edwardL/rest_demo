package gdbtmp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// MapsToStruct map数组转结构体指针数组 兼容gdb标签
func MapsToStruct(mList []map[string]any, dest any) (err error) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		err = fmt.Errorf("panic: %v", r)
	//	}
	//}()
	var modelValue = reflect.ValueOf(dest)
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Slice || modelValue.Elem().Type().Elem().Kind() != reflect.Ptr {
		return errors.New("接收结果类型错误,应该为[]*结构体指针")
	}
	var val = reflect.Indirect(modelValue)
	var typElem = val.Type().Elem().Elem()

	// 只反射一次获取结构体信息
	var vt = reflect.Indirect(reflect.New(typElem))
	var ft = reflect.ValueOf(vt.Interface()).Type()
	var fnType = map[string]string{}
	var fieldTags = make(map[string]structFieldInfo) // 保存字段标签映射
	var field reflect.StructField
	var tag GdbTag
	for i := 0; i < ft.NumField(); i++ {
		field = ft.Field(i)
		tag = getStructFileName(field)
		if tag.Name == "-" {
			continue
		}
		fnType[tag.Name] = field.Type.String()
		// 保存标签到字段名的映射
		fieldTags[tag.Name] = structFieldInfo{
			Name:      field.Name,
			Anonymous: field.Anonymous,
		}
	}
	var mVal, elem, mapField reflect.Value
	var mapVal any
	var ok bool
	for _, r := range mList {
		mVal = reflect.Indirect(reflect.New(typElem)).Addr()
		elem = mVal.Elem()
		for tagName, sFieldName := range fieldTags {
			mapField = elem.FieldByName(sFieldName.Name)
			if sFieldName.Anonymous { // 处理嵌套结构体
				err = MapToStruct(r, mapField)
				if err != nil {
					return err
				}
				continue
			}
			if mapVal, ok = r[tagName]; ok {
				if mapField.IsValid() && mapField.CanSet() {
					err = structAssignment(mapField, tagName, mapVal)
					if err != nil {
						return err
					}
				}
			}
		}

		val = reflect.Append(val, mVal)
	}
	reflect.Indirect(modelValue).Set(val)
	return nil
}

// SlicesToStruct 将[][]string转换为结构体切片（第一项为字段名，其余为值）
func SlicesToStruct(sList [][]string, dest any) (err error) {
	if len(sList) < 1 {
		return nil
	}
	var modelValue = reflect.ValueOf(dest)
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Slice || modelValue.Elem().Type().Elem().Kind() != reflect.Ptr {
		return errors.New("接收结果类型错误,应该为[]*结构体指针")
	}
	var val = reflect.Indirect(modelValue)
	var typElem = val.Type().Elem().Elem()

	var fieldNames = sList[0]  // 第一项为字段名
	var valuesList = sList[1:] // 其余为值列表
	var fieldNameIndex = map[string]int{}
	for index, fn := range fieldNames {
		fieldNameIndex[fn] = index
	}

	// 只反射一次获取结构体信息
	var vt = reflect.Indirect(reflect.New(typElem))
	var ft = reflect.ValueOf(vt.Interface()).Type()
	var fnType = map[string]string{}
	var fieldTags = make(map[string]structFieldInfo) // 保存字段标签映射
	var field reflect.StructField
	var tag GdbTag
	for i := 0; i < ft.NumField(); i++ {
		field = ft.Field(i)
		tag = getStructFileName(field)
		if tag.Name == "-" {
			continue
		}
		fnType[tag.Name] = field.Type.String()
		// 保存标签到字段名的映射
		fieldTags[tag.Name] = structFieldInfo{
			Name:      field.Name,
			Anonymous: field.Anonymous,
		}
	}
	var mVal, elem, mapField reflect.Value
	var fIndex int
	var ok bool
	// 遍历值列表转换为结构体
	for _, values := range valuesList {
		// 检查当前行的值数量是否与字段名数量一致
		if len(values) != len(fieldNames) {
			return fmt.Errorf("值数量不匹配，期望 %d 个，实际 %d 个", len(fieldNames), len(values))
		}

		mVal = reflect.Indirect(reflect.New(typElem)).Addr()
		elem = mVal.Elem()
		for tagName, sFieldName := range fieldTags {
			mapField = elem.FieldByName(sFieldName.Name)
			if sFieldName.Anonymous { // 处理嵌套结构体
				err = SliceToStruct([][]string{fieldNames, values}, mapField)
				if err != nil {
					return err
				}
				continue
			}
			if fIndex, ok = fieldNameIndex[tagName]; ok {
				if mapField.IsValid() && mapField.CanSet() {
					err = structAssignment(mapField, tagName, values[fIndex])
					if err != nil {
						return err
					}
				}
			}
		}

		val = reflect.Append(val, mVal)
	}

	// 设置最终结果
	reflect.Indirect(modelValue).Set(val)
	return nil
}

// MapToStruct map转结构体 兼容gdb标签
func MapToStruct(a map[string]any, s any) (err error) {
	if len(a) == 0 {
		return nil
	}
	//defer func() {
	//	if r := recover(); r != nil {
	//		err = fmt.Errorf("panic: %v", r)
	//	}
	//}()
	// 获取 s 的反射值
	var v reflect.Value
	if s != nil {
		if vf, ok := s.(reflect.Value); ok {
			v = vf
		} else {
			v = reflect.ValueOf(s)
			if v.Kind() != reflect.Ptr {
				return errors.New("接收结果类型错误,应该为结构体指针")
			}
		}
		for v.Kind() == reflect.Ptr {
			if v.IsNil() && v.CanAddr() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
	} else {
		v = reflect.ValueOf(s)
	}

	// 获取类型信息
	var t = v.Type()
	var tag GdbTag
	var field reflect.StructField
	// 直接将 map 数据填充到结构体中
	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		if field.Anonymous { // 处理嵌套结构体
			err = MapToStruct(a, v.Field(i))
			if err != nil {
				return err
			}
			continue
		}
		tag = getStructFileName(field)
		if tag.Name == "-" {
			continue
		}

		// 根据标签名称从 map 中获取值
		if val, ok := a[tag.Name]; ok {
			err = structAssignment(v.Field(i), tag.Name, val)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// SliceToStruct 将[][]string转换为转结构体 兼容gdb标签
func SliceToStruct(sList [][]string, s any) (err error) {
	if len(sList) < 2 {
		return nil
	}
	// 获取 s 的反射值
	var v reflect.Value
	if s != nil {
		if vf, ok := s.(reflect.Value); ok {
			v = vf
		} else {
			v = reflect.ValueOf(s)
			if v.Kind() != reflect.Ptr {
				return errors.New("接收结果类型错误,应该为结构体指针")
			}
		}
		for v.Kind() == reflect.Ptr {
			if v.IsNil() && v.CanAddr() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
	} else {
		v = reflect.ValueOf(s)
	}

	var fieldNames = sList[0] // 第一项为字段名
	var valuesList = sList[1] // 第二项为值列表
	var fieldNameIndex = map[string]int{}
	for index, fn := range fieldNames {
		fieldNameIndex[fn] = index
	}

	// 获取类型信息
	var t = v.Type()
	var tag GdbTag
	var field reflect.StructField
	var fIndex int
	var ok bool
	// 直接将 map 数据填充到结构体中
	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		if field.Anonymous { // 处理嵌套结构体
			err = SliceToStruct(sList, v.Field(i))
			if err != nil {
				return err
			}
			continue
		}
		tag = getStructFileName(field)
		if tag.Name == "-" {
			continue
		}

		// 根据标签名称从 切片 中获取值
		if fIndex, ok = fieldNameIndex[tag.Name]; ok {
			err = structAssignment(v.Field(i), tag.Name, valuesList[fIndex])
			if err != nil {
				return err
			}
		}
	}
	return err
}

// StructToMap 将结构体转换为map[string]any  兼容gdb标签
func StructToMap(value any) (result map[string]any, err error) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		err = fmt.Errorf("panic: %v", r)
	//	}
	//}()
	// 获取value的反射值
	var v reflect.Value
	if vf, ok := value.(reflect.Value); ok {
		v = vf
	} else {
		v = reflect.ValueOf(value)
	}
	// 处理指针类型，递归解引用
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 检查是否为结构体类型
	if v.Kind() != reflect.Struct {
		return nil, errors.New(notStructErrMsg)
	}
	result = make(map[string]any)
	var field reflect.StructField
	var childResult map[string]any
	var t = v.Type()
	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		if field.Anonymous { // 处理嵌套结构体
			childResult, err = StructToMap(v.Field(i))
			if err != nil {
				return nil, err
			}
			for ck, cv := range childResult {
				result[ck] = cv
			}
			continue
		}
		mapAssignment(field, v.Field(i), result)
	}
	return result, err
}

// StructToMaps 将结构体切片转换为[]map[string]any  兼容gdb标签
func StructToMaps(value any) (result []map[string]any, err error) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		err = fmt.Errorf("panic: %v", r)
	//	}
	//}()

	// 获取value的反射值
	var v reflect.Value
	if vf, ok := value.(reflect.Value); ok {
		v = vf
	} else {
		v = reflect.ValueOf(value)
	}
	// 处理指针类型，递归解引用
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// 检查是否为切片类型
	if v.Kind() != reflect.Slice {
		return nil, errors.New("value must be a slice or pointer to slice")
	}
	var field reflect.StructField
	var itemValue reflect.Value
	// 遍历切片中的每个元素
	for i := 0; i < v.Len(); i++ {
		itemValue = v.Index(i)
		// 处理指针类型，递归解引用
		for itemValue.Kind() == reflect.Ptr {
			itemValue = itemValue.Elem()
		}

		// 检查是否为结构体类型
		if itemValue.Kind() != reflect.Struct {
			return nil, errors.New("slice items must be structs or pointers to structs")
		}

		var t = itemValue.Type()
		var itemMap = make(map[string]any)
		var childResult map[string]any
		// 遍历结构体字段
		for j := 0; j < t.NumField(); j++ {
			field = t.Field(j)
			if field.Anonymous { // 处理嵌套结构体
				childResult, err = StructToMap(itemValue.Field(j))
				if err != nil {
					return nil, err
				}
				for ck, cv := range childResult {
					itemMap[ck] = cv
				}
				continue
			}
			mapAssignment(field, itemValue.Field(j), itemMap)
		}

		result = append(result, itemMap)
	}

	return result, err
}

// 结构体赋值
func structAssignment(f reflect.Value, fn string, val any) error {
	// 处理指针值
	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			if f.CanSet() {
				var elem = reflect.New(f.Type().Elem()).Elem()
				f.Set(elem.Addr()) // 为指针赋值
			} else {
				return fmt.Errorf("%s cannot dereference nil pointer (unsettable)", fn)
			}
		}
		f = f.Elem() // 解引用指针
	}
	if conf.StructAssignment != nil {
		var next, err = conf.StructAssignment(f, val)
		if !next || err != nil {
			return err
		}
	}
	if f.Type() == TimeType {
		var t, err = ToTime(val)
		if err != nil {
			return err
		}
		f.Set(reflect.ValueOf(t))
		return nil
	}
	switch f.Kind() {
	case reflect.String:
		f.SetString(ToString(val))
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if intVal, e := ToInt64(val); e == nil {
			f.SetInt(intVal)
		} else {
			return fmt.Errorf("cannot convert %v to %s[int]", val, fn)
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		if uIntVal, e := ToUint64(val); e == nil {
			f.SetUint(uIntVal)
		} else {
			return fmt.Errorf("cannot convert %v to %s[uint]", val, fn)
		}
	case reflect.Float32, reflect.Float64:
		if floatVal, e := ToFloat64E(val); e == nil {
			f.SetFloat(floatVal)
		} else {
			return fmt.Errorf("cannot convert %v to %s[float]", val, fn)
		}
	case reflect.Bool:
		if boolVal, e := ToBool(val); e == nil {
			f.SetBool(boolVal)
		} else {
			return fmt.Errorf("cannot convert %v to %s[bool]", val, fn)
		}
	default:
		return fmt.Errorf("%s unsupported type: %s", fn, f.Kind())
	}
	return nil
}

// map赋值
func mapAssignment(field reflect.StructField, fieldValue reflect.Value, itemMap map[string]any) {
	var err error
	var tag GdbTag
	tag = getStructFileName(field)
	if tag.Name == "-" {
		return
	}
	if isNilValue(fieldValue) { // nil 直接不处理
		return
	}
	for fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}
	if conf.MapAssignment != nil {
		var next = conf.MapAssignment(field, fieldValue, itemMap, tag)
		if !next {
			return
		}
	}
	if tag.Type == "" {
		itemMap[tag.Name] = fieldValue.Interface()
		return
	}
	// 处理特殊类型
	var timeVal time.Time
	timeVal, err = ToTime(fieldValue.Interface())
	if err != nil {
		defLog.CtxWarn(context.Background(), tag.Name, "类型转换失败：", err)
		return
	}
	switch tag.Type {
	case TagTypeTimeUnixMilli:
		if timeVal.IsZero() {
			itemMap[tag.Name] = 0
		} else {
			itemMap[tag.Name] = timeVal.UnixMilli()
		}
	case TagTypeTimeUnix:
		if timeVal.IsZero() {
			itemMap[tag.Name] = 0
		} else {
			itemMap[tag.Name] = timeVal.Unix()
		}
	default:
		if _, ok := fieldValue.Interface().(time.Time); ok {
			if !timeVal.IsZero() {
				itemMap[tag.Name] = timeVal.Format(tag.Type)
			}
		}
	}
}

// isNilValue 判断值是否为nil
func isNilValue(v reflect.Value) bool {
	// 先检查类型是否支持nil
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	default:
		// 非引用类型不存在nil
		return false
	}
}

type GdbTag struct {
	Name string
	Type string
}

// getStructFileName 获取结构体字段名 兼容db字段标签
func getStructFileName(f reflect.StructField) GdbTag {
	var res GdbTag
	var isTime = f.Type == TimeType || f.Type == TimePtrType
	if isTime { // 默认时间类型
		res.Type = time.DateTime
	}
	var tag = f.Tag.Get("gdbtmp")
	var jsonTag = f.Tag.Get("json")
	if jsonTag == "" {
		jsonTag = f.Name
	}
	if strings.Contains(jsonTag, ",") {
		jsonTag = strings.Split(jsonTag, ",")[0]
	}
	if tag == "" {
		res.Name = jsonTag
		return res
	}
	if !strings.Contains(tag, ":") {
		res.Name = tag
		return res
	}
	var tgRes GdbTag
	for _, tStr := range strings.Split(tag, ";") {
		if !strings.Contains(tStr, ":") {
			tgRes.Name = tStr
		} else {
			var tgT = strings.Split(tStr, ":")
			if len(tgT) >= 2 && tgT[0] == TagTypeKey {
				tgRes.Type = strings.Join(tgT[1:], ":")
			}
		}
	}
	if tgRes.Name == "" {
		tgRes.Name = jsonTag
	}
	return tgRes
}

// toMap any 转 map[string]any
func toMap(d any) (map[string]any, error) {
	var m, err = MapToMapAny(d)
	if err == nil {
		err = nil
		return m, nil
	}
	// 不是map类型那就是结构体
	// 内部会对d进行类型判断 无需在调用前重复判断
	m, err = StructToMap(d)
	if err == nil {
		return m, nil
	}
	if err.Error() != notStructErrMsg {
		defLog.CtxWarn(context.Background(), "StructToMap 失败 使用默认转换：", err)
	}
	err = nil
	err = ConvAToB(d, &m)
	return m, err
}

// toMaps any 转 []map[string]any
func toMaps(d any) ([]map[string]any, error) {
	var m, err = MapsToMapsAny(d)
	if err == nil {
		err = nil
		return m, nil
	}
	m, err = StructToMaps(d)
	if err == nil {
		return m, nil
	}
	if err.Error() != notStructErrMsg {
		defLog.CtxWarn(context.Background(), "StructToMaps 失败 使用默认转换：", err)
	}
	err = nil
	err = ConvAToB(d, &m)
	return m, err
}

// StructDbField 获取结构体里的字段名
func StructDbField(s any) []string {
	fieldArr := make([]string, 0)
	var v reflect.Value
	if s != nil {
		v = reflect.ValueOf(s)
		for v.Kind() == reflect.Ptr {
			if v.IsNil() && v.CanAddr() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
	} else {
		v = reflect.ValueOf(s)
	}
	var t = v.Type()
	var tag GdbTag
	var field reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		tag = getStructFileName(field)
		if tag.Name != "" && tag.Name != "-" {
			fieldArr = append(fieldArr, tag.Name)
		}
	}
	return fieldArr
}

// ConvAToB 类型互转
func ConvAToB(a, b any) error {
	w := &bytes.Buffer{}
	err := json.NewEncoder(w).Encode(a)
	if err != nil {
		return err
	}
	err = json.NewDecoder(w).Decode(b)
	if err != nil {
		return err
	}
	return nil
}

// IsEmpty 判断any类型的值是否为空（零值）
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	return isEmptyValue(val)
}

// IsSlice 判断值是否为切片
func IsSlice(v any) bool {
	var t = reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Slice
}

// IsArray 判断值是否为数组
func IsArray(v any) bool {
	var t reflect.Type
	if vr, ok := v.(reflect.Value); ok {
		t = vr.Type()
	} else {
		t = reflect.TypeOf(v)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Array
}

// IsArrayOrSlice 判断是否为数组或切片
func IsArrayOrSlice(v any) bool {
	var t reflect.Type
	if vr, ok := v.(reflect.Value); ok {
		t = vr.Type()
	} else {
		t = reflect.TypeOf(v)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Array || t.Kind() == reflect.Slice
}

// IsStruct 判断值是否为结构体或结构体指针
func IsStruct(v any) bool {
	if v == nil {
		return false
	}
	var t reflect.Type
	if vr, ok := v.(reflect.Value); ok {
		t = vr.Type()
	} else {
		t = reflect.TypeOf(v)
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

// Dump 格式化打印
func Dump(d any) {
	DumpCtx(setLogCallDepthCtx(context.Background(), 4), d)
}

// DumpCtx 格式化打印
func DumpCtx(ctx context.Context, d any) {
	ctx = setLogCallDepthCtx(ctx, getLogCallDepthCtx(ctx)) // 修正打印的行号
	var jsonByte []byte
	switch v := d.(type) {
	case string:
		jsonByte = []byte(v)
	case []byte:
		jsonByte = v
	default:
		marshal, err := json.Marshal(d)
		if err != nil {
			defLog.CtxInfof(ctx, "\n%v", d)
			return
		}
		jsonByte = marshal
	}
	if len(jsonByte) == 0 {
		jsonByte = []byte("{}")
	}
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, jsonByte, "", "\t")
	if err != nil {
		defLog.CtxInfof(ctx, "\n%v", string(jsonByte))
		return
	}
	defLog.CtxInfof(ctx, "\n%v", string(prettyJSON.Bytes()))
}

// isEmptyValue 检查reflect.Value是否为空值
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理time.Time（如果需要）
		if v.Type() == TimeType {
			return v.Interface().(time.Time).IsZero()
		}
		if v.Type() == TimePtrType {
			if isNilValue(v) {
				return true
			}
			return v.Interface().(*time.Time).IsZero()
		}
		return true
	default:
		return false
	}
}

func strToArr(str string, sep string, delEmpty bool) []string {
	arr := make([]string, 0)
	strArr := strings.Split(str, sep)
	if !delEmpty {
		return strArr
	}
	for _, v := range strArr {
		if v != "" {
			arr = append(arr, v)
		}
	}
	return arr
}

// replaceIndex 替换出现在指定位置的 ?
func replaceIndex(query string, pos int, new string) (q string, err error) {
	idx := strIndex(query, "?", pos)
	if idx == -1 {
		return "", errors.New("[ " + query + " ] 参数和占位符不匹配")
	}
	return query[0:idx] + new + query[idx+1:], nil
}

// strIndex 字符串第几次出现的位置
func strIndex(s, substr string, pos int) int {
	if pos == 0 {
		pos = 1
	}
	if pos == 1 {
		return strings.Index(s, substr)
	}
	lastI := -1
	st := s
	for i := 1; i <= pos; i++ {
		idx := strings.Index(st, substr)
		if idx == -1 {
			return idx
		}
		st = st[idx+1:]
		lastI += idx + 1
	}
	return lastI
}

// genQueryList 构建数组查询 [1,2,3] => ?,?,?
func genQueryList[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](query string, arr []T, phIndex *int, args *[]any) (string, error) {
	var argsLen = len(arr)
	var err error
	query, err = genPrePil(query, *phIndex, argsLen)
	if err != nil {
		return "", err
	}
	*phIndex += argsLen
	for _, av := range arr {
		*args = append(*args, av)
	}
	return query, nil
}

// genQueryGroupList 构建数组查询 [1,2,3],[1,2,3] => (?,?,?),(?,?,?)
func genQueryGroupList[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](query string, arr [][]T, phIndex *int, args *[]any) (string, error) {
	var argsLen = len(arr)
	var childLen = len(arr[0])
	var i int
	var err error
	query, err = genPrePilGroup(query, *phIndex, argsLen, childLen)
	if err != nil {
		return "", err
	}
	*phIndex += argsLen * childLen
	for _, av := range arr {
		for i = 0; i < childLen; i++ {
			*args = append(*args, av[i])
		}
	}
	return query, nil
}

// genQueryGroupList 构建数组查询 [1,2,3],[1,2,3] => (?,?,?),(?,?,?)
func genQueryGroupAnyList(query string, arr [][]any, phIndex *int, args *[]any) (string, error) {
	var argsLen = len(arr)
	var childLen = len(arr[0])
	var err error
	query, err = genPrePilGroup(query, *phIndex, argsLen, childLen)
	if err != nil {
		return "", err
	}
	*phIndex += argsLen * childLen
	for _, v := range arr {
		*args = append(*args, v[:childLen]...)
	}
	return query, nil
}

// genPrePil 构建预编译sql where id = ?
func genPrePil(query string, pos, len int) (q string, err error) {
	var zw = strings.Repeat(", ?", len)
	zw = strings.TrimLeft(zw, ", ")
	return replaceIndex(query, pos, zw)
}

// genPrePilGroup 构建预编译sql组 where (id,ts) = (?,?)
func genPrePilGroup(query string, pos, len, size int) (q string, err error) {
	var zw = strings.Repeat(", ?", size)
	zw = strings.TrimLeft(zw, ", ")
	zw = "(" + zw + ")"
	zw = strings.Repeat(", "+zw, len)
	zw = strings.TrimLeft(zw, ", ")
	return replaceIndex(query, pos, zw)
}

// arrToArrList 构建数组列表
func arrToArrList(a []any) ([][]any, error) {
	var res [][]any = make([][]any, len(a))
	for k, arg := range a {
		switch arr := arg.(type) {
		case []int:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int8:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int16:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint8:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint16:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []string:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []float32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []float64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []bool:
			for _, v := range arr {
				if v {
					res[k] = append(res[k], 1)
				} else {
					res[k] = append(res[k], 0)
				}
			}
		case []any:
			res[k] = arr
		default:
			return nil, errors.New("参数类型错误,不支持的嵌套切片类型")
		}
	}
	return res, nil
}

// 解析类型并调用TableName方法
func getTableNameRecursive(val reflect.Value) (string, error) {
	switch val.Kind() {
	case reflect.Ptr:
		// 处理指针类型（如*Aaa、**Aaa）
		if val.IsNil() {
			// 空指针：尝试通过类型创建非空实例（避免解引用空指针）
			var ptrType = val.Type()
			var elemType = ptrType.Elem()
			// 创建指针指向的元素的实例（如*Aaa的元素是Aaa，创建Aaa{}）
			var newElem = reflect.New(elemType).Elem()
			return getTableNameRecursive(newElem)
		}
		// 非空指针：先检查自身是否实现接口，若实现则直接调用
		if val.Type().Implements(modelFaceType) {
			return (val.Interface().(ModelFace)).TableName(), nil
		}
		// 若自身不实现，解引用后递归（如**Aaa -> *Aaa）
		return getTableNameRecursive(val.Elem())
	case reflect.Slice, reflect.Array:
		// 处理切片/数组（如[]Aaa、[]*Aaa、*[]Aaa等）
		var elemType = val.Type().Elem()
		// 创建元素类型的实例（用于检查和调用方法）
		var elemVal reflect.Value
		if elemType.Kind() == reflect.Ptr {
			// 元素是指针类型（如[]*Aaa）：创建指针实例（&Aaa{}）
			elemVal = reflect.New(elemType.Elem())
		} else {
			// 元素是值类型（如[]Aaa）：创建值实例（Aaa{}）
			elemVal = reflect.New(elemType).Elem()
		}
		// 递归检查元素类型
		return getTableNameRecursive(elemVal)
	default:
		// 处理值类型（如Aaa、Bbb）
		// 检查值类型是否实现接口（值接收者情况）
		if val.Type().Implements(modelFaceType) {
			return (val.Interface().(ModelFace)).TableName(), nil
		}
		// 检查指针类型是否实现接口（指针接收者情况）
		var ptrVal = reflect.New(val.Type()) // 创建指针实例（&Aaa{}）
		if ptrVal.Type().Implements(modelFaceType) {
			return (ptrVal.Interface().(ModelFace)).TableName(), nil
		}
		// 均不实现
		return "", fmt.Errorf("类型 %s 及其指针均不实现ModelFace", val.Type())
	}
}
