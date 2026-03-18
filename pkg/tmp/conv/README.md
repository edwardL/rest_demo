# utils/conv（模型友好版）

`utils/conv` 是通用转换工具包，覆盖：

- 标量类型转换（`string/int/uint/float/bool/time`）
- 结构体与 `map` 双向映射（支持 `gdb` tag）
- 切片字符串解析与目标类型转换
- 反射辅助判断（是否数组/切片/结构体/空值）

## 1) 快速使用

```go
package main

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
)

type User struct {
	Id   int    `json:"id" gdb:"id"`
	Name string `json:"name" gdb:"name"`
}

func main() {
	var i64, err = conv.ToInt64("123")
	if err != nil {
		panic(err)
	}

	var b, err2 = conv.ToBool("yes")
	if err2 != nil {
		panic(err2)
	}

	var m = map[string]any{"id": "1", "name": "alice"}
	var u User
	var err3 = conv.MapToStruct(m, &u, true)
	if err3 != nil {
		panic(err3)
	}

	fmt.Println(i64, b, u)
}
```

## 2) 关键参数约定

- `dbConvInitPtr=true`：结构体字段为指针时，转换会主动初始化指针
- `dbConvInitPtr=false`：零值会尽量保留为空指针
- 时间字段支持秒/毫秒/微秒/纳秒时间戳和常见格式字符串

## 3) 全部导出 API（完整）

### 常量与变量

- `TagTypeKey`
- `TagTypeTimeUnixMilli`
- `TagTypeTimeUnix`
- `TimeType`
- `TimePtrType`
- `NotStructErrMsg`

### 类型

- `type ValType reflect.Kind`
- `type StructTag struct { Name string; Type string }`

### ValType 可选值

- `TypeBool`
- `TypeInt`
- `TypeInt8`
- `TypeInt16`
- `TypeInt32`
- `TypeInt64`
- `TypeUint`
- `TypeUint8`
- `TypeUint16`
- `TypeUint32`
- `TypeUint64`
- `TypeFloat32`
- `TypeFloat64`
- `TypeString`

### 函数

- `AToB(a, b any) error`
- `AssignId(s any, id any) error`
- `GetStructFieldList(t reflect.Type) []string`
- `IsArray(v any) bool`
- `IsArrayOrSlice(v any) bool`
- `IsEmpty(v any) bool`
- `IsNilValue(v reflect.Value) bool`
- `IsSlice(v any) bool`
- `IsStruct(v any) bool`
- `MapToStruct(a map[string]any, s any, dbConvInitPtr ...bool) error`
- `MapsToStruct(mList []map[string]any, dest any, dbConvInitPtr ...bool) error`
- `ParseValToStrSlice(val string) []string`
- `SliceToSlice(sList [][]string, s any) error`
- `SplitTrim(s, sep string) []string`
- `StrToTargetType(strVal string, targetType reflect.Type) (reflect.Value, error)`
- `StructToMap(value any, dbConvInitPtr ...bool) (map[string]any, error)`
- `StructToMaps(value any, dbConvInitPtr ...bool) ([]map[string]any, error)`
- `ToBool(v any) (bool, error)`
- `ToFloat64E(i any) (float64, error)`
- `ToInt(i any) (int, error)`
- `ToInt32(i any) (int32, error)`
- `ToInt64(i any) (int64, error)`
- `ToPtr[T any](v T) *T`
- `ToString(i any) string`
- `ToTime(v any, timeLocation ...*time.Location) (time.Time, error)`
- `ToUint64(v any) (uint64, error)`
- `ToValType(input any, targetType ValType) any`

## 4) 场景示例

### 字符串数组入参转目标切片

```go
var source = [][]string{{"id"}, {"1"}, {"2"}, {"3"}}
var ids []int
var err = conv.SliceToSlice(source, &ids)
if err != nil {
	panic(err)
}
```

### 时间字段自动转换

```go
var t1, err = conv.ToTime("1700000000")
var t2, err2 = conv.ToTime("2025-01-02 03:04:05")
_, _, _, _ = t1, t2, err, err2
```

## 5) 测试命令

```bash
go test ./utils/conv -v
go test ./utils/conv -run '^TestToIntSeries$' -v
```
