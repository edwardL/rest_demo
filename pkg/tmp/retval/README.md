# utils/retval（模型友好版）

`utils/retval` 用于快速构造统一返回结构，提供 5 组成功/失败构造函数以及视图感知的成功返回函数。

## 1) 设计说明

- 每个函数返回 `(uint8, any)`
- `uint8` 表示返回结构版本（1~5）
- `any` 是具体的 `ResultData0X` 结构
- `Success*` 统一使用成功码与成功消息
- `Err*` 由调用方传入 `code` 与 `msg`
- `SuccessPageView/SuccessView` 支持视图字段过滤

## 2) 全部导出 API（完整）

### 成功返回

- `Success1(data ...[]map[string]interface{}) (uint8, any)`
- `Success2(data ...string) (uint8, any)`
- `Success3(data ...map[string]string) (uint8, any)`
- `Success4(data ...[]any) (uint8, any)`
- `Success5(data ...any) (uint8, any)`

### 视图感知成功返回（支持字段白名单过滤）

- `SuccessPageView[I inter, T any](info *types.SitkWebCliInfo, totalNums I, pageData []T) (uint8, any)`
    - 分页视图返回，自动按 `info.ReqViewInfo.SearchViewInfo.FieldList` 过滤字段
    - `totalNums` 支持整数类型（int/int8/int16/int32/int64/uint 系列）
    - 返回结构：`{"page_data": [...], "total_nums": ...}`
- `SuccessView[T any](info *types.SitkWebCliInfo, data T) (uint8, any)`
    - 单条/列表视图返回，自动按视图字段过滤

### 错误返回

- `Err1(code, msg string, data ...[]map[string]interface{}) (uint8, any)`
- `Err2(code, msg string, data ...string) (uint8, any)`
- `Err3(code, msg string, data ...map[string]string) (uint8, any)`
- `Err4(code, msg string, data ...[]interface{}) (uint8, any)`
- `Err5(code, msg string, data ...any) (uint8, any)`

## 3) 用法示例

### 基础成功/错误返回

```go
package main

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/retval"
)

func main() {
	var t1, okRes = retval.Success5(map[string]any{"id": 1})
	var t2, errRes = retval.Err5("E_USER_NOT_FOUND", "用户不存在")

	fmt.Println(t1, okRes)
	fmt.Println(t2, errRes)
}
```

### 视图感知返回（带字段过滤）

```go
package main

import (
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/retval"
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// 构造视图信息，指定允许返回的字段
	var info = &ftypes.SitkWebCliInfo{
		ReqViewInfo: &ftypes.SitkWebViewInfo{
			SearchViewInfo: ftypes.DataViewSearchInfo{
				FieldList: map[string]ftypes.DataViewField{
					"id":   {FieldAlias: "id"},
					"name": {FieldAlias: "name"},
				},
			},
		},
	}

	var users = []User{{Id: 1, Name: "tom", Age: 20}, {Id: 2, Name: "jack", Age: 30}}

	// 分页返回（会自动过滤，只保留 id 和 name 字段）
	var t1, resp1 = retval.SuccessPageView(info, 2, users)
	// resp1.Data: [{"page_data": [{"id": 1, "name": "tom"}, {"id": 2, "name": "jack"}], "total_nums": 2}]

	// 单条返回
	var t2, resp2 = retval.SuccessView(info, users[0])
	// resp2.Data: [{"id": 1, "name": "tom"}]

	_, _ = t1, t2
}
```

## 4) 选择建议

- 需要列表对象：用 `Success1/Err1`
- 需要简单字符串：用 `Success2/Err2`
- 需要 `map[string]string`：用 `Success3/Err3`
- 需要通用数组：用 `Success4/Err4`
- 需要任意结构体或任意对象：用 `Success5/Err5`
- 需要分页视图返回并按视图过滤字段：用 `SuccessPageView`
- 需要单条/列表视图返回并按视图过滤字段：用 `SuccessView`

## 5) 测试命令

```bash
go test ./utils/retval -v
go test ./utils/retval -run '^TestSuccessPageView' -v
```
