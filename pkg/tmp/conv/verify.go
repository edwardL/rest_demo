package conv

import (
	"reflect"
	"strconv"
	"time"
)

// isNum 是否数字
func isNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsNilValue 判断值是否为nil
func IsNilValue(v reflect.Value) bool {
	// 先检查类型是否支持nil
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	default:
		// 非引用类型不存在nil
		return false
	}
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
			if IsNilValue(v) {
				return true
			}
			return v.Interface().(*time.Time).IsZero()
		}
		return true
	default:
		return false
	}
}
