package str

import (
	"regexp"
	"strconv"
	"strings"

	"rest_demo/pkg/utils/arrays"
)

var (
	matchNonAlphaNumeric = regexp.MustCompile(`[^a-zA-Z0-9]+`)     // 非常规字符
	matchFirstCap        = regexp.MustCompile("(.)([A-Z][a-z]+)")  // 拆分出连续大写
	matchAllCap          = regexp.MustCompile("([a-z0-9])([A-Z])") // 拆分单词
)

// ToSnakeLower 转下划线
func ToSnakeLower(s string) string {
	s = matchNonAlphaNumeric.ReplaceAllString(s, "_")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// EmptyIf 如果为空则返回第二个参数
func EmptyIf(str1 string, str2 string) string {
	if len(str1) != 0 {
		return str1
	}
	return str2
}

// NilIf 如果为空则返回第二个参数
func NilIf(str1 *string, str2 string) string {
	if str1 != nil {
		return *str1
	}
	return str2
}

// JoinLimit Join
func JoinLimit(elems []string, sep string, limit int) string {
	if len(elems) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for i, item := range elems {
		builder.WriteString(item)
		if i >= limit-1 {
			break
		}
		if i != len(elems)-1 {
			builder.WriteString(sep)
		}
	}
	return builder.String()
}

// JoinInt Join去重加过滤空字符
func JoinInt(elems []int, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for i, item := range elems {
		builder.WriteString(strconv.Itoa(item))
		if i != len(elems)-1 {
			builder.WriteString(sep)
		}
	}
	return builder.String()
}

// JoinDistinctTrimEmpty Join去重加过滤空字符
func JoinDistinctTrimEmpty(sep string, elems ...string) string {
	return strings.Join(arrays.FilterEmptyString(arrays.Distinct(elems)), sep)
}

// Split 按分隔符分割字符串
func Split(s string, sep string) []string {
	if len(s) == 0 {
		return []string{}
	}
	return strings.Split(s, sep)
}

// SplitAny 按分隔符分割字符串
func SplitAny(s string, sep ...string) []string {
	if len(s) == 0 {
		return []string{}
	}
	result := []string{s}
	for _, item := range sep {
		var t []string
		for _, r := range result {
			for _, n := range Split(r, item) {
				if len(n) != 0 {
					t = append(t, n)
				}
			}
		}
		result = t
	}
	return result
}

// SplitInt 按分隔符分割字符串为整形
func SplitInt(s string, sep string) []int {
	return arrays.Map(Split(s, sep), func(item string) int {
		result, err := strconv.Atoi(item)
		if err != nil {
			return 0
		}
		return result
	})
}
