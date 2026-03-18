# 验证器使用文档

## 目录

- [验证器概述](#验证器概述)
- [验证规则](#验证规则)
    - [基础验证规则](#基础验证规则)
    - [复合验证规则](#复合验证规则)
- [标签定义](#标签定义)
- [使用示例](#使用示例)
- [错误处理](#错误处理)

## 验证器概述

验证器通过结构体标签定义验证规则，支持多种验证类型，包括必填验证、长度验证、数值比较、枚举验证、条件验证等。验证器能够处理嵌套结构体，并支持字段间的依赖验证。

## 验证规则

### 基础验证规则

| 规则         | 描述                  | 示例                                   |
|------------|---------------------|--------------------------------------|
| `required` | 必填字段验证，字段不能为空值      | `v:"required"`                       |
| `enum`     | 枚举值验证，字段值必须在指定列表中   | `v:"enum:active,inactive,suspended"` |
| `eq`       | 等于验证，字段值必须等于指定值     | `v:"eq:target_value"`                |
| `lt`       | 小于验证，字段值必须小于指定值     | `v:"lt:100"`                         |
| `elt`      | 小于等于验证，字段值必须小于等于指定值 | `v:"elt:100"`                        |
| `gt`       | 大于验证，字段值必须大于指定值     | `v:"gt:0"`                           |
| `egt`      | 大于等于验证，字段值必须大于等于指定值 | `v:"egt:0"`                          |

### 复合验证规则

| 规则    | 描述                                                         | 示例                                                 |
|-------|------------------------------------------------------------|----------------------------------------------------|
| `len` | 长度验证，支持字符串、切片、数组和映射等类型 长度必须大于20                            | `v:"len:gt:20"`                                    |
| `dp`  | 依赖字段验证，根据其他字段的值来决定验证规则  如果name不为空且等于aaa 则当前字段不能为空且大于20     | `v:"dp:name[required\|eq:aaa],[required \|gt:20]"` |
| `if`  | 条件表达式验证，根据条件表达式决定是否验证  \$表示当前字段的值 $name标识name字段的值 不支持() 加权 | `v:"if:[$ != aa && $name == aaa]"`                 |

### 函数验证规则

验证器支持通过自定义函数扩展验证功能，包括条件验证函数和比较函数：

| 函数类型   | 函数名     | 描述       | 使用示例                            |
|--------|---------|----------|---------------------------------|
| 条件验证函数 | `true`  | 判断参数是否为真 | `v:"if:true($field)"`           |
| 条件验证函数 | `false` | 判断参数是否为假 | `v:"if:false($field)"`          |
| 比较函数   | `len`   | 获取参数的长度  | `v:"if:$ == len($other_field)"` |

## 标签定义

验证器使用以下标签来定义验证规则：

- `v` - 验证规则标签
- `vCode` - 验证错误码标签 `<` 使用上一个错误码
- `vMsg` - 验证错误信息标签 `<` 使用上一个错误信息

### 标签分隔符

- `|` - 验证规则分隔符
- `:` - 规则参数分隔符
- `[]` - 依赖规则选项组分隔符
- `,` - 依赖规则选项组内规则分隔符

### 示例结构体定义

```go
type User struct {
    Name   string `json:"name" v:"required|len:gt:20" vCode:"001" vMsg:"姓名为必填项且长度需大于20"`
    Age    int    `json:"age" v:"required|egt:0|elt:150" vCode:"002" vMsg:"年龄为必填项且应在0-150之间"`
    Status string `json:"status" v:"required|enum:active,inactive,suspended" vCode:"003" vMsg:"状态为必填项且应为指定枚举值"`
}
```

## 使用示例

```go
import "nwgit.gzhhit.com/BD/hhitcommcode.git/utils/validator"

// 定义带验证规则的结构体
type User struct {
    Name   string `json:"name" v:"required|len:gt:20" vCode:"001" vMsg:"姓名为必填项且长度需大于20"`
    Age    int    `json:"age" v:"required|egt:0|elt:150" vCode:"002" vMsg:"年龄为必填项且应在0-150之间"`
}

// 创建验证器实例并执行验证
func validateUser(user User) error {
    v := validator.NewValidator("9999").ValidAll(true)
    err := v.Validate(user)
    return err
}

// 使用示例
user := User{
    Name: "this_is_a_very_long_name_over_20_chars",
    Age:  25,
}

if err := validateUser(user); err != nil {
    fmt.Println("验证失败:", err.Error())
} else {
    fmt.Println("验证通过")
}
```

## 错误处理

验证器在验证失败时会返回包含详细错误信息的错误对象，错误信息包括字段名、错误码和错误消息。

```go
type ValidationError struct {
    Field string `json:"field"`  // 字段名
    Code  string `json:"code"`   // 错误码
    Msg   string `json:"msg"`    // 错误消息
}
```

错误信息格式示例：

```
name(001): 姓名为必填项且长度需大于20; age(002): 年龄为必填项且应在0-150之间
```

当验证失败时，可以通过遍历 `ValidationErrors` 获取每个字段的详细错误信息。

## 扩展函数

验证器支持通过以下函数扩展验证功能：

### WithConditionalFunction

设置自定义条件验证函数，条件验证函数直接返回布尔结果：

```go
// 定义自定义条件验证函数
func customCondition(v any, vl VarValues, args ...any) (bool, error) {
    // 实现自定义逻辑
    return true, nil
}

// 注册函数
validator.WithConditionalFunction("myFunc", customCondition)
```

### WithCompareFunction

设置自定义比较函数，比较函数返回一个值用于后续比较：

```go
// 定义自定义比较函数
func customCompare(v any, vl VarValues, args ...any) (any, error) {
    // 实现自定义逻辑，返回可用于比较的值
    return len(args), nil
}

// 注册函数
validator.WithCompareFunction("myCompare", customCompare)
```

### WithErrCodeMsg

设置单个错误码对应的错误信息：

```go
validator.WithErrCodeMsg("E001", "自定义错误信息")
```

### WithErrCodesMsg

批量设置错误码对应的错误信息：

```go
errorMap := map[string]string{
    "E001": "错误信息1",
    "E002": "错误信息2",
    "E003": "错误信息3",
}
validator.WithErrCodesMsg(errorMap)
```
