package validator

import (
	"fmt"
	"testing"
)

type User2 struct {
	NameAa string `json:"name_aa" v:"required" vCode:"95AD" vMsg:"名字必填大于20字"`
	Age    int    `json:"age" v:"required|egt:0|elt:150" vCode:"95AE" vMsg:"年龄必填且应在0-150之间"`
	Email  string `json:"email" v:"required|len:gt:5" vCode:"|95AF" vMsg:"邮箱必填且长度应大于5"`
	Status string `json:"status" v:"required|enum:active,inactive,suspended" vCode:"95AG" vMsg:"状态必填且应为active,inactive,suspended之一"`
}

type User struct {
	User2
	Name     string   `json:"name" v:"required|len:gt:20|if:[len($) > 0]" vCode:"95AD" vMsg:"名字必填大于20字"`
	Email    string   `json:"email" v:"dp:name[required],[required]" vCode:"96AE" vMsg:"邮箱必填"`
	UserInfo *User    `json:"user_info"`
	Age      int      `json:"age" v:"required|egt:0|elt:150" vCode:"95AH" vMsg:"年龄必填且应在0-150之间且应小于父级年龄"`
	Score    int      `json:"score" v:"required|egt:0|elt:100" vCode:"95AI" vMsg:"分数应在0-100之间"`
	Tags     []string `json:"tags" v:"required|len:gt:0" vCode:"95AJ" vMsg:"标签必填且不能为空"`
}

func Test_Validator(t *testing.T) {
	// 测试验证失败的情况
	t.Run("ValidationFailed", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		err := v.Validate(User{
			Name:  "aaa",
			Email: "bbb",
			UserInfo: &User{
				Name: "aaa",
			},
		})
		if err == nil {
			t.Error("Expected validation to fail, but it passed")
		} else {
			fmt.Println("Validation errors:", err.Error())
		}
	})

	// 测试验证通过的情况
	t.Run("ValidationPassed", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "test",
				Age:    25,
				Email:  "test@example.com",
				Status: "active",
			},
			Name:  longName,
			Email: longName, // 使 email == name 成立
			UserInfo: &User{
				User2: User2{
					NameAa: "test2",
					Age:    30,
					Email:  "test2@example.com",
					Status: "inactive",
				},
				Name:  longName,
				Email: longName,
				Age:   30,
				Score: 95,
				Tags:  []string{"tag1", "tag2"},
			},
			Age:   25,
			Score: 80,
			Tags:  []string{"tag1", "tag2", "tag3"},
		})
		if err != nil {
			t.Errorf("Expected validation to pass, but got errors: %v", err)
		}
	})

	// 测试required规则
	t.Run("RequiredRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		err := v.Validate(User{
			Name:  "", // 违反required规则
			Email: "test@example.com",
		})
		if err == nil {
			t.Error("Expected validation to fail due to required rule, but it passed")
		}
	})

	// 测试len:gt规则
	t.Run("LengthRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		err := v.Validate(User{
			Name:  "short", // 不满足len:gt:20规则
			Email: "test@example.com",
		})
		if err == nil {
			t.Error("Expected validation to fail due to length rule, but it passed")
		}
	})

	// 测试条件验证规则 (if:[$ != aa && $name == aaa])
	t.Run("ConditionalRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		// 当Name为"aaa"时，会触发条件验证，但"aaa"长度不足20字符，所以会失败
		err := v.Validate(User{
			Name:  "aaa",
			Email: "test@example.com",
		})
		if err == nil {
			t.Error("Expected conditional validation to fail, but it passed")
		}
	})

	// 测试依赖字段规则 (dp:name[required|eq:aaa],[required|gt:20])
	t.Run("DependentFieldRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		// Name为"aaa"时，触发第一个条件，要求Email必须为"aaa"
		err := v.Validate(User{
			Name:  "aaa",
			Email: "bbb", // 不满足eq:aaa
		})
		if err == nil {
			t.Error("Expected dependent field validation to fail, but it passed")
		}
	})

	// 测试嵌套结构体验证
	t.Run("NestedStructValidation", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "test",
				Age:    25,
				Email:  "test@example.com",
				Status: "active",
			},
			Name:  longName,
			Email: longName,
			UserInfo: &User{
				Name:  "short", // UserInfo中的Name不满足长度要求
				Email: "test@example.com",
			},
			Age:   25,
			Score: 80,
			Tags:  []string{"tag1", "tag2"},
		})
		if err == nil {
			t.Error("Expected nested validation to fail, but it passed")
		}
	})

	// 测试嵌套required字段
	t.Run("NestedRequiredField", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "", // 违反required规则
				Age:    25,
				Email:  "test@example.com",
				Status: "active",
			},
			Name:  longName,
			Email: longName,
			Age:   25,
			Score: 80,
			Tags:  []string{"tag1", "tag2"},
		})
		if err == nil {
			t.Error("Expected nested required field validation to fail, but it passed")
		}
	})

	// 测试数值比较规则 (egt, elt)
	t.Run("NumericComparisonRules", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "test",
				Age:    -1, // 违反egt:0规则
				Email:  "test@example.com",
				Status: "active",
			},
			Name:  longName,
			Email: longName,
			Age:   151, // 违反elt:150规则
			Score: 80,
			Tags:  []string{"tag1", "tag2"},
		})
		if err == nil {
			t.Error("Expected numeric comparison validation to fail, but it passed")
		}
	})

	// 测试枚举规则
	t.Run("EnumRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "test",
				Age:    25,
				Email:  "test@example.com",
				Status: "invalid_status", // 不在枚举值中
			},
			Name:  longName,
			Email: longName,
			Age:   25,
			Score: 80,
			Tags:  []string{"tag1", "tag2"},
		})
		if err == nil {
			t.Error("Expected enum validation to fail, but it passed")
		}
	})

	// 测试切片长度规则
	t.Run("SliceLengthRule", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		longName := "this_is_a_very_long_name_over_20_chars"
		err := v.Validate(User{
			User2: User2{
				NameAa: "test",
				Age:    25,
				Email:  "test@example.com",
				Status: "active",
			},
			Name:  longName,
			Email: longName,
			Age:   25,
			Score: 80,
			Tags:  []string{}, // 违反len:gt:0规则
		})
		if err == nil {
			t.Error("Expected slice length validation to fail, but it passed")
		}
	})

	// 测试map校验
	t.Run("MapValidator", func(t *testing.T) {
		var v = NewValidator("9999").ValidAll(true)
		//longName := "this_is_a_very_long_name_over_20_chars"
		err := v.ValidateByRoutes(map[string]any{
			"name": "ddd",
			"age":  300,
		}, map[string]ValidateRules{
			"name": {
				Code:  "95AD",
				Msg:   "名字必填且大于20字",
				Rules: "required|len:gt:20",
			},
			"age": {
				Code:  "95AH|<|<",
				Msg:   "年龄必填且应在0-150之间|<|<",
				Rules: "required|egt:0|elt:150",
			},
		})
		fmt.Println(err.Errors())
		if err == nil {
			t.Error("Expected slice length validation to fail, but it passed")
		}
	})
}

func Test_Eq(t *testing.T) {
	type Aaa struct { // 扫描探头ID
		ScanProtocol []int8 `json:"scan_protocol" v:"required|len:eq:2"` // 扫描协议[[1,0]=IPv4,[0,1]=IPv6, [1,1]=IPv4+IPv6]
	}
	t.Run("len = 2", func(t *testing.T) {
		var err = Validate(Aaa{ScanProtocol: []int8{0, 0}})
		if err != nil {
			t.Error("Expected slice length validation to fail, but it passed")
		}
	})

	t.Run("len != 2", func(t *testing.T) {
		var err = Validate(Aaa{ScanProtocol: []int8{0, 0, 0}})
		if err == nil {
			t.Error("Expected slice length validation to fail, but it passed")
		}
		fmt.Println(err.Errors())
	})
}
