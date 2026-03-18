package conv

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// fieldInfo 结构体定义，用于缓存结构体字段信息
type fieldInfo struct {
	index     int
	name      string
	structTag StructTag
	fieldType reflect.Type
	anonymous bool
	isPtr     bool
	kind      reflect.Kind
}

// fieldCache 缓存结构体类型到字段信息的映射，使用LRU策略限制大小
type fieldCacheType struct {
	mu      sync.RWMutex
	items   map[reflect.Type][]fieldInfo
	order   []reflect.Type // 记录访问顺序
	maxSize int            // 最大缓存大小
}

func newFieldCache(maxSize int) *fieldCacheType {
	return &fieldCacheType{
		items:   make(map[reflect.Type][]fieldInfo),
		order:   make([]reflect.Type, 0),
		maxSize: maxSize,
	}
}

var fieldCache = newFieldCache(10000) // 限制缓存最多10000种类型 200万个字段 大约 200m
// getStructFieldInfo 获取结构体字段信息，带缓存
func getStructFieldInfo(t reflect.Type) []fieldInfo {
	// 先尝试从缓存获取
	fieldCache.mu.RLock()
	if info, exists := fieldCache.items[t]; exists {
		fieldCache.mu.RUnlock()
		return info
	}
	fieldCache.mu.RUnlock()

	fieldCache.mu.Lock()
	defer fieldCache.mu.Unlock()

	// 再次检查，万一命中了勒
	if info, exists := fieldCache.items[t]; exists {
		return info
	}

	var fieldInfoList = make([]fieldInfo, 0, t.NumField())
	var field reflect.StructField
	var tag StructTag
	for i := 0; i < t.NumField(); i++ {
		field = t.Field(i)
		tag = getStructTag(field)
		if tag.Name == "-" {
			continue
		}
		fieldInfoList = append(fieldInfoList, fieldInfo{
			index:     i,
			name:      field.Name,
			structTag: tag,
			fieldType: field.Type,
			anonymous: field.Anonymous,
			isPtr:     field.Type.Kind() == reflect.Ptr,
			kind:      field.Type.Kind(),
		})
	}

	// 检查是否需要清理缓存
	if len(fieldCache.items) >= fieldCache.maxSize {
		// 移除最早添加的项
		if len(fieldCache.order) > 0 {
			delete(fieldCache.items, fieldCache.order[0])
			fieldCache.order = fieldCache.order[1:]
		}
	}

	// 添加新项
	fieldCache.items[t] = fieldInfoList
	fieldCache.order = append(fieldCache.order, t)

	return fieldInfoList
}

// GetStructFieldList 获取结构体字段列表
func GetStructFieldList(t reflect.Type) []string {
	var fileList = make([]string, 0)
	var fi = getStructFieldInfo(t)
	for _, f := range fi {
		if f.anonymous { // 处理嵌套结构体
			fileList = append(fileList, GetStructFieldList(f.fieldType)...)
			continue
		}
		fileList = append(fileList, f.structTag.Name)
	}
	return fileList
}

// MapsToStruct map数组转结构体指针数组 兼容gdb标签
func MapsToStruct(mList []map[string]any, dest any, dbConvInitPtr ...bool) (err error) {
	var modelValue = reflect.ValueOf(dest)
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Slice {
		return errors.New("接收结果类型错误,应该为*[]any(struct)")
	}
	var val = reflect.Indirect(reflect.MakeSlice(modelValue.Type().Elem(), 0, len(mList)))
	var typElem = val.Type().Elem()
	var fieldTypElem = typElem
	for fieldTypElem.Kind() == reflect.Ptr {
		fieldTypElem = fieldTypElem.Elem()
	}

	// 使用缓存的字段信息
	var fieldInfoList = getStructFieldInfo(fieldTypElem)
	var mVal, elem, mapField reflect.Value
	var mapVal any
	var ok bool
	for _, r := range mList {
		mVal = reflect.Indirect(reflect.New(typElem)).Addr()
		elem = mVal // 原始值 用来赋值
		for elem.Kind() == reflect.Ptr {
			if elem.IsNil() && elem.CanAddr() {
				elem.Set(reflect.New(elem.Type().Elem()))
			}
			elem = elem.Elem()
		}

		// 使用预计算的字段信息
		for _, fi := range fieldInfoList {
			if fi.anonymous { // 处理嵌套结构体
				err = MapToStruct(r, elem.Field(fi.index), dbConvInitPtr...)
				if err != nil {
					return err
				}
				continue
			}
			if mapVal, ok = r[fi.structTag.Name]; ok {
				mapField = elem.Field(fi.index)
				if mapField.IsValid() && mapField.CanSet() {
					err = structAssignment(mapField, fi.structTag, mapVal, dbConvInitPtr...)
					if err != nil {
						return err
					}
				}
			}
		}

		val = reflect.Append(val, mVal.Elem())
	}
	reflect.Indirect(modelValue).Set(val)
	return nil
}

// MapToStruct map转结构体 兼容gdb标签
func MapToStruct(a map[string]any, s any, dbConvInitPtr ...bool) (err error) {
	if len(a) == 0 {
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

	// 使用缓存的字段信息
	var fieldInfoList = getStructFieldInfo(v.Type())
	var val any
	var ok bool
	// 直接将 map 数据填充到结构体中
	// 使用预计算的字段信息
	for _, fi := range fieldInfoList {
		if fi.anonymous { // 处理嵌套结构体
			// 获取嵌套结构体的map数据
			err = MapToStruct(a, v.Field(fi.index), dbConvInitPtr...)
			if err != nil {
				return err
			}
			continue
		}

		// 根据标签名称从 map 中获取值
		if val, ok = a[fi.structTag.Name]; ok {
			err = structAssignment(v.Field(fi.index), fi.structTag, val, dbConvInitPtr...)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// StructToMap 将结构体转换为map[string]any  兼容gdb标签
func StructToMap(value any, dbConvInitPtr ...bool) (result map[string]any, err error) {
	// 获取value的反射值
	var v reflect.Value
	if vf, ok := value.(reflect.Value); ok {
		v = vf
	} else {
		v = reflect.ValueOf(value)
	}
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// 处理指针类型，递归解引用
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 检查是否为结构体类型
	if v.Kind() != reflect.Struct {
		return nil, errors.New(NotStructErrMsg)
	}
	result = make(map[string]any)
	var field reflect.StructField
	var childResult map[string]any
	var t = v.Type()

	// 使用缓存的字段信息
	var fieldInfoList = getStructFieldInfo(t)

	// 遍历结构体字段
	for _, fi := range fieldInfoList {
		field = t.Field(fi.index)
		if field.Anonymous { // 处理嵌套结构体
			childResult, err = StructToMap(v.Field(fi.index), dbConvInitPtr...)
			if err != nil {
				return nil, err
			}
			for ck, cv := range childResult {
				result[ck] = cv
			}
			continue
		}
		err = mapAssignment(field, fi.fieldType, fi.structTag, v.Field(fi.index), result, dbConvInitPtr...)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
	}
	return result, err
}

// StructToMaps 将结构体切片转换为[]map[string]any  兼容gdb标签
func StructToMaps(value any, dbConvInitPtr ...bool) (result []map[string]any, err error) {
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

		for itemValue.Kind() == reflect.Interface {
			itemValue = itemValue.Elem()
		}
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

		// 使用缓存的字段信息
		var fieldInfoList = getStructFieldInfo(t)

		// 遍历结构体字段
		for _, fi := range fieldInfoList {
			field = t.Field(fi.index)
			if field.Anonymous { // 处理嵌套结构体
				childResult, err = StructToMap(itemValue.Field(fi.index), dbConvInitPtr...)
				if err != nil {
					return nil, err
				}
				for ck, cv := range childResult {
					itemMap[ck] = cv
				}
				continue
			}
			err = mapAssignment(field, fi.fieldType, fi.structTag, itemValue.Field(fi.index), itemMap, dbConvInitPtr...)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", field.Name, err)
			}
		}

		result = append(result, itemMap)
	}

	return result, err
}

// 结构体赋值
func structAssignment(f reflect.Value, tag StructTag, val any, dbConvInitPtr ...bool) error {
	if tag.Name == "" || tag.Name == "-" {
		return nil
	}
	// 判断是否自定义了解析函数
	if f.CanAddr() {
		if f.Addr().Type().Implements(fieldScanFaceType) {
			var v, ok = f.Addr().Interface().(sql.Scanner)
			if !ok {
				return errors.New("sql.Scanner 断言失败")
			}
			return v.Scan(val)
		}
	}

	var fn = tag.Name
	if len(dbConvInitPtr) <= 0 || !dbConvInitPtr[0] { // 零值不初始化指针
		var originalField = f
		var isPtrField = originalField.Kind() == reflect.Ptr
		// 检查是否是零值，如果是零值且字段是指针类型，则将指针设置为nil
		if isPtrField && isZeroValue(val) && originalField.CanSet() {
			originalField.Set(reflect.Zero(originalField.Type()))
			return nil
		}
		if isPtrField && originalField.CanSet() {
			defer func() {
				// 处理特殊类型的零值 比如时间
				if isZeroValue(f.Interface()) {
					originalField.Set(reflect.Zero(originalField.Type()))
				}
			}()
		}
	}

	// 处理指针值
	for f.Kind() == reflect.Ptr {
		if f.IsNil() {
			if f.CanSet() {
				f.Set(reflect.New(f.Type().Elem())) // 为指针赋值
			} else {
				return fmt.Errorf("%s cannot dereference nil pointer (unsettable)", fn)
			}
		}
		f = f.Elem() // 解引用指针
	}

	// 特殊处理 json.RawMessage 类型
	if f.Type() == RawMessageType || f.Type() == SliceByteType {
		switch v := val.(type) {
		case json.RawMessage:
			f.Set(reflect.ValueOf([]byte(v)))
		case []byte:
			f.Set(reflect.ValueOf(v))
		case string:
			f.Set(reflect.ValueOf([]byte(v)))
		default:
			return fmt.Errorf("%s unsupported type for json.RawMessage: %T", fn, val)
		}
		return nil
	}

	if f.Type() == TimeType || f.Type() == TimePtrType {
		var t, err = ToTime(val)
		if err != nil {
			return err
		}
		if f.Type() == TimeType {
			f.Set(reflect.ValueOf(t))
		} else {
			f.Set(reflect.ValueOf(&t))
		}
		return nil
	}

	// 根据值的类型进行优化处理
	switch v := val.(type) {
	case string:
		// 如果值是字符串，直接根据字段类型转换，避免ToString调用开销
		switch f.Kind() {
		case reflect.String:
			f.SetString(v)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, e := ToInt64(v); e == nil {
				f.SetInt(intVal)
			} else {
				return fmt.Errorf("cannot convert %v to %s[int]", val, fn)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if uIntVal, e := ToUint64(v); e == nil {
				f.SetUint(uIntVal)
			} else {
				return fmt.Errorf("cannot convert %v to %s[uint]", val, fn)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, e := ToFloat64E(v); e == nil {
				f.SetFloat(floatVal)
			} else {
				return fmt.Errorf("cannot convert %v to %s[float]", val, fn)
			}
		case reflect.Bool:
			if boolVal, e := ToBool(v); e == nil {
				f.SetBool(boolVal)
			} else {
				return fmt.Errorf("cannot convert %v to %s[bool]", val, fn)
			}
		default:
			return fmt.Errorf("%s unsupported type: %s", fn, f.Kind())
		}
	default:
		// 其他类型按原有逻辑处理
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
	}
	return nil
}

// map 赋值
func mapAssignment(field reflect.StructField, fileType reflect.Type, tag StructTag, fieldValue reflect.Value, itemMap map[string]any, dbConvInitPtr ...bool) error {
	if tag.Name == "" || tag.Name == "-" {
		return nil
	}
	var err error
	// 判断是否自定义了解析函数
	if fieldValue.Type().Implements(fieldValueFaceType) {
		var v, ok = fieldValue.Interface().(driver.Valuer)
		if !ok {
			return errors.New("driver.Valuer 断言失败")
		}

		var resultVal driver.Value
		resultVal, err = v.Value()
		if err != nil {
			return err
		}
		itemMap[tag.Name] = resultVal
		return nil
	}
	for fileType.Kind() == reflect.Ptr {
		fileType = fileType.Elem()
	}
	if IsNilValue(fieldValue) && fileType != TimeType && fileType != RawMessageType { // json.RawMessage 零值也要输出
		if len(dbConvInitPtr) > 0 && dbConvInitPtr[0] { // 零值初始化指针
			itemMap[tag.Name] = reflect.New(field.Type).Interface()
		} else {
			itemMap[tag.Name] = nil
		}
		return nil
	}
	for fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			// 指针为 nil，直接返回 nil
			itemMap[tag.Name] = nil
			return nil
		}
		fieldValue = fieldValue.Elem()
	}
	if tag.Type == "" {
		itemMap[tag.Name] = fieldValue.Interface()
		return nil
	}
	if fileType == TimeType { // 处理时间类型
		var timeVal time.Time
		if !IsNilValue(fieldValue) && !isZeroValue(fieldValue) {
			timeVal, err = ToTime(fieldValue.Interface())
			if err != nil {
				return err
			}
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
			if !timeVal.IsZero() {
				itemMap[tag.Name] = timeVal.Format(tag.Type)
			} else {
				itemMap[tag.Name] = ""
			}
		}
	}
	return nil
}

type StructTag struct {
	Name string
	Type string
}

// getStructTag 获取结构体字段名 兼容db字段标签
func getStructTag(f reflect.StructField) StructTag {
	var res StructTag
	var isTime = f.Type == TimeType || f.Type == TimePtrType
	if isTime { // 默认时间类型
		res.Type = time.DateTime
	}
	var tag = f.Tag.Get("gdb")
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
	var tgRes StructTag
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

// AToB 类型互转 兜底方案谨慎使用
// 优先使用 MapsToStruct 和 StructToMaps 系列函数
func AToB(a, b any) error {
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

// isZeroValue 检查值是否为零值
func isZeroValue(val any) bool {
	switch v := val.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case int:
		return v == 0
	case int8:
		return v == 0
	case int16:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case uint:
		return v == 0
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0
	case float32:
		return v == 0.0
	case float64:
		return v == 0.0
	case bool:
		return v == false
	case []byte:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	default:
		// 其他类型，使用反射检查零值
		var ov = reflect.ValueOf(val)
		if ov.IsZero() || ov.Interface() == reflect.Zero(ov.Type()).Interface() {
			return true
		}
		for ov.Kind() == reflect.Ptr {
			ov = ov.Elem()
		}
		// 时间类型特殊处理
		if ov.Type() == TimeType {
			return (ov.Interface().(time.Time)).UnixNano() == 0
		}
	}
	return false
}

// findIDField 递归查找结构体中的ID字段Id，包括嵌套结构体
func findIDField(v reflect.Value) reflect.Value {
	idField := v.FieldByName("Id")
	if !idField.IsValid() {
		idField = v.FieldByName("ID")
		return idField
	}
	if idField.IsValid() {
		return idField
	}

	// 在嵌套结构体中查找
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 如果是匿名嵌入的结构体（embedded struct），也进行递归查找
		if fieldType.Anonymous {
			if field.Kind() == reflect.Struct {
				nestedIDField := findIDField(field)
				if nestedIDField.IsValid() {
					return nestedIDField
				}
			} else if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
				if !field.IsNil() {
					nestedIDField := findIDField(field.Elem())
					if nestedIDField.IsValid() {
						return nestedIDField
					}
				}
			}
		}
	}
	return reflect.Value{} // 返回无效的Value
}

// AssignId 为结构体的ID字段赋值
func AssignId(s any, id any) error {
	if s == nil {
		return errors.New("目标结构体不能为nil")
	}

	v := reflect.ValueOf(s)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return errors.New("目标结构体不能为nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("目标必须是结构体")
	}

	// 递归查找ID字段，包括嵌套结构体
	idField := findIDField(v)
	if !idField.IsValid() {
		return errors.New("结构体中未找到ID字段")
	}

	if !idField.CanSet() {
		return errors.New("ID字段不可设置")
	}

	// 如果ID字段是指针类型，需要特殊处理
	for idField.Kind() == reflect.Ptr {
		if idField.IsNil() {
			// 创建指针指向的值
			idField.Set(reflect.New(idField.Type().Elem()))
		}
		idField = idField.Elem()
	}

	// 将id值转换为ID字段的类型并赋值
	idValue := reflect.ValueOf(id)
	if idValue.Type().ConvertibleTo(idField.Type()) {
		idField.Set(idValue.Convert(idField.Type()))
	} else {
		return fmt.Errorf("无法将类型 %s 转换为 %s", idValue.Type(), idField.Type())
	}

	return nil
}
