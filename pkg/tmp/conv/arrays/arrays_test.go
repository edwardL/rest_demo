package arrays

import (
	"errors"
	"testing"
)

func TestForEach(t *testing.T) {
	var arr = []int{1, 2, 3}
	var sum int
	ForEach(arr, func(i int) {
		sum += i
	})
	if sum != 6 {
		t.Fatalf("ForEach 求和失败: %d", sum)
	}
}

func TestSplit(t *testing.T) {
	var arr = []int{1, 2, 3, 4, 5, 6, 7}
	var result = Split(arr, 3)
	if len(result) != 3 {
		t.Fatalf("Split 长度错误: %d", len(result))
	}
	if len(result[0]) != 3 || len(result[1]) != 3 || len(result[2]) != 1 {
		t.Fatalf("Split 子数组长度错误: %v", result)
	}

	result = Split(arr, 0)
	if len(result) != 1 {
		t.Fatalf("Split size=0 长度错误: %d", len(result))
	}

	result = Split(arr, -1)
	if len(result) != 1 {
		t.Fatalf("Split size<0 长度错误: %d", len(result))
	}
}

func TestFilter(t *testing.T) {
	var arr = []int{1, 2, 3, 4, 5}
	var result = Filter(arr, func(i int) bool {
		return i > 3
	})
	if len(result) != 2 || result[0] != 4 || result[1] != 5 {
		t.Fatalf("Filter 结果错误: %v", result)
	}
}

func TestFilterEmptyString(t *testing.T) {
	var arr = []string{"a", "", "b", "", "c"}
	var result = FilterEmptyString(arr)
	if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Fatalf("FilterEmptyString 结果错误: %v", result)
	}
}

func TestMap(t *testing.T) {
	var arr = []int{1, 2, 3}
	var result = Map(arr, func(i int) string {
		return string(rune('A' + i - 1))
	})
	if len(result) != 3 || result[0] != "A" || result[1] != "B" || result[2] != "C" {
		t.Fatalf("Map 结果错误: %v", result)
	}
}

func TestMapError(t *testing.T) {
	var arr = []int{1, 2, 3}
	var result []string
	var err error
	result, err = MapError(arr, func(i int) (string, error) {
		return string(rune('A' + i - 1)), nil
	})
	if err != nil || len(result) != 3 {
		t.Fatalf("MapError 正常场景错误: %v %v", result, err)
	}

	_, err = MapError(arr, func(i int) (string, error) {
		if i == 2 {
			return "", errors.New("test error")
		}
		return "", nil
	})
	if err == nil {
		t.Fatalf("MapError 错误场景应返回错误")
	}
}

func TestDistinct(t *testing.T) {
	var arr = []int{1, 2, 3, 2, 1, 4, 3}
	var result = Distinct(arr)
	if len(result) != 4 {
		t.Fatalf("Distinct 长度错误: %d", len(result))
	}
	if !Contains(result, 1) || !Contains(result, 2) || !Contains(result, 3) || !Contains(result, 4) {
		t.Fatalf("Distinct 结果错误: %v", result)
	}
}

func TestContains(t *testing.T) {
	var arr = []string{"a", "b", "c"}
	if !Contains(arr, "a") {
		t.Fatalf("Contains 应包含 a")
	}
	if Contains(arr, "d") {
		t.Fatalf("Contains 不应包含 d")
	}
}

func TestContainsIf(t *testing.T) {
	var arr = []int{1, 2, 3, 4, 5}
	if !ContainsIf(arr, func(i int) bool { return i > 4 }) {
		t.Fatalf("ContainsIf 应找到大于4的元素")
	}
	if ContainsIf(arr, func(i int) bool { return i > 10 }) {
		t.Fatalf("ContainsIf 不应找到大于10的元素")
	}
}

func TestGroup(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}
	var arr = []user{
		{Name: "a", Age: 10},
		{Name: "b", Age: 20},
		{Name: "c", Age: 10},
	}
	var result = Group(arr, func(u user) int {
		return u.Age
	})
	if len(result) != 2 {
		t.Fatalf("Group 分组数量错误: %d", len(result))
	}
	if len(result[10]) != 2 || len(result[20]) != 1 {
		t.Fatalf("Group 分组元素数量错误: %v", result)
	}
}
