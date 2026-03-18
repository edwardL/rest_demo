package arrays

// ForEach 数组遍历
func ForEach[T any](arr []T, f func(T)) {
	for i := 0; i < len(arr); i++ {
		f(arr[i])
	}
}

// Split 数组分段
func Split[T any](arr []T, size int) [][]T {
	if size < 1 {
		return [][]T{arr}
	}
	var result [][]T
	var subArr []T
	for i := 0; i < len(arr); i++ {
		item := arr[i]
		if len(subArr) >= size && size != 0 {
			result = append(result, subArr)
			subArr = []T{}
		}
		subArr = append(subArr, item)
	}
	result = append(result, subArr)
	return result
}

// Filter 数组过滤
func Filter[T any](arr []T, f func(T) bool) []T {
	var result []T
	for i := 0; i < len(arr); i++ {
		item := arr[i]
		if f(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterEmptyString 数组过滤空字符串
func FilterEmptyString(arr []string) []string {
	var result []string
	for i := 0; i < len(arr); i++ {
		item := arr[i]
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

// Map 将数组映射成另一个数组
func Map[T any, R any](arr []T, f func(T) R) []R {
	r := make([]R, len(arr))
	for i := 0; i < len(arr); i++ {
		r[i] = f(arr[i])
	}
	return r
}

// MapError 将数组映射成另一个数组
func MapError[T any, R any](arr []T, f func(T) (R, error)) ([]R, error) {
	r := make([]R, len(arr))
	for i := 0; i < len(arr); i++ {
		item := arr[i]
		v, err := f(item)
		if err != nil {
			return nil, err
		}
		r[i] = v
	}
	return r, nil
}

// Distinct 数组去重
func Distinct[T comparable](arr []T) []T {
	var arrLen = len(arr) - 1
	for ; arrLen > 0; arrLen-- {
		for j := arrLen - 1; j >= 0; j-- {
			if arr[arrLen] == arr[j] {
				arr = append(arr[:arrLen], arr[arrLen+1:]...)
				break
			}
		}
	}
	return arr
}

// Contains 判断数组是否包含某个元素
func Contains[T comparable](arr []T, i T) bool {
	for _, item := range arr {
		if item == i {
			return true
		}
	}
	return false
}

// ContainsIf 判断数组是否包含某个元素，根据条件判断
func ContainsIf[T any](arr []T, f func(i T) bool) bool {
	for _, item := range arr {
		if f(item) {
			return true
		}
	}
	return false
}

// Group 分组
func Group[K comparable, T any](arr []T, f func(T) K) map[K][]T {
	result := make(map[K][]T)
	for i := 0; i < len(arr); i++ {
		k := f(arr[i])
		result[k] = append(result[k], arr[i])
	}
	return result
}
