package convert

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ToInt 转Int类型
func ToInt(s any) (int, error) {
	switch v := s.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case uintptr:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	}
	return 0, fmt.Errorf("%v Unable to convert to Int %v", reflect.ValueOf(s), s)
}

// ToString 转为字符串
func ToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []byte:
		return string(v)
	case []any:
		var result []string
		for _, item := range v {
			result = append(result, ToString(item))
		}
		return strings.Join(result, ",")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ToBits 整数转二进制数组
func ToBits[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](n T, bitSize int) []T {
	var result = make([]T, bitSize)
	for i := 0; i < bitSize; i++ {
		result[i] = n & 1
		n >>= 1
	}
	return result
}

// FromBits 二进制数组转整数
func FromBits[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](bits []T) T {
	var result T
	var bitSize = len(bits)
	for i := 0; i < bitSize; i++ {
		result |= bits[bitSize-i-1] << (bitSize - i - 1)
	}
	return result
}
