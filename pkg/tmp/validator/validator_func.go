package validator

import (
	"errors"
	"reflect"
)

// trueFunc 是否为真
func trueFunc(v any) (bool, error) {
	switch pv := v.(type) {
	case bool:
		return pv, nil
	case string:
		return pv == "true", nil
	case int:
		return pv > 0, nil
	case int8:
		return pv > 0, nil
	case int16:
		return pv > 0, nil
	case int32:
		return pv > 0, nil
	case int64:
		return pv > 0, nil
	case uint:
		return pv > 0, nil
	case uint8:
		return pv > 0, nil
	case uint16:
		return pv > 0, nil
	case uint32:
		return pv > 0, nil
	case uint64:
		return pv > 0, nil
	case float32:
		return pv > 0, nil
	case float64:
		return pv > 0, nil
	}
	return !isZeroValue(reflect.ValueOf(v)), nil
}

// lenFunc 获取长度
func lenFunc(val any) (int, error) {
	var v reflect.Value
	if rv, ok := val.(reflect.Value); ok {
		v = rv
	} else {
		v = reflect.ValueOf(val)
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if isZeroValue(v) {
		return 0, nil
	}
	var sourceLen = 0
	switch v.Kind() {
	case reflect.String:
		sourceLen = len([]rune(v.String()))
	case reflect.Slice, reflect.Map, reflect.Array:
		sourceLen = v.Len()
	default:
		return 0, errors.New("类型不支持") // 类型不支持
	}
	return sourceLen, nil
}
