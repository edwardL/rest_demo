# utils/conv/arrays（模型友好版）

`utils/conv/arrays` 提供泛型数组/切片操作工具函数。

## 1) 快速使用

```go
package main

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv/arrays"
)

func main() {
	// ForEach 遍历
	arrays.ForEach([]int{1, 2, 3}, func(i int) {
		fmt.Println(i)
	})

	// Split 分段
	var chunks = arrays.Split([]int{1, 2, 3, 4, 5}, 2)
	fmt.Println(chunks) // [[1 2] [3 4] [5]]

	// Filter 过滤
	var evens = arrays.Filter([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 == 0
	})
	fmt.Println(evens) // [2 4]

	// Map 映射
	var strs = arrays.Map([]int{1, 2, 3}, func(i int) string {
		return fmt.Sprintf("num%d", i)
	})
	fmt.Println(strs) // [num1 num2 num3]

	// Distinct 去重
	var uniq = arrays.Distinct([]int{1, 2, 2, 3, 1})
	fmt.Println(uniq) // [1 2 3]

	// Contains 判断存在
	if arrays.Contains([]string{"a", "b"}, "a") {
		fmt.Println("包含 a")
	}

	// Group 分组
	var users = []struct {
		Name string
		Age  int
	}{{"tom", 20}, {"jack", 20}, {"alice", 30}}
	var groups = arrays.Group(users, func(u struct {
		Name string
		Age  int
	}) int {
		return u.Age
	})
	fmt.Println(groups[20]) // [{tom 20} {jack 20}]
}
```

## 2) 全部导出 API（完整）

### 函数

- `ForEach[T any](arr []T, f func(T))`
    - 遍历数组，对每个元素执行 f
- `Split[T any](arr []T, size int) [][]T`
    - 将数组按 size 分段，size < 1 时返回整个数组
- `Filter[T any](arr []T, f func(T) bool) []T`
    - 过滤数组，保留满足 f 的元素
- `FilterEmptyString(arr []string) []string`
    - 过滤空字符串
- `Map[T any, R any](arr []T, f func(T) R) []R`
    - 将数组映射为另一个类型数组
- `MapError[T any, R any](arr []T, f func(T) (R, error)) ([]R, error)`
    - 映射数组，支持错误中断
- `Distinct[T comparable](arr []T) []T`
    - 去重，返回唯一元素数组
- `Contains[T comparable](arr []T, i T) bool`
    - 判断数组是否包含元素 i
- `ContainsIf[T any](arr []T, f func(i T) bool) bool`
    - 判断数组是否包含满足 f 的元素
- `Group[K comparable, T any](arr []T, f func(T) K) map[K][]T`
    - 按 f 返回的 key 分组

## 3) 测试命令

```bash
go test ./utils/conv/arrays -v
go test ./utils/conv/arrays -run '^TestSplit$' -v
```
