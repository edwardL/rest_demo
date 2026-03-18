package validator

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// 测试用例结构体
type TestCase struct {
	Name     string
	Expr     string
	Vars     VarValues
	Expected bool
	HasError bool
}

func initFunc() {
	// 条件验证函数映射表
	conf.conditionalFunctionMap = map[string]ConditionalFunction{
		// 检查$的值是否在指定范围内
		"dollarInRange": func(v any, vl VarValues, args ...any) (bool, error) {
			num, ok := v.(int) // v是$变量的值
			if !ok {
				return false, errors.New("dollarInRange要求$是整数")
			}
			if len(args) < 2 {
				return false, errors.New("dollarInRange需要两个参数: min和max")
			}

			min, ok1 := args[0].(int)
			max, ok2 := args[1].(int)
			if !ok1 || !ok2 {
				return false, errors.New("dollarInRange参数必须是整数")
			}

			return num >= min && num <= max, nil
		},

		// 检查$的值是否等于另一个变量的值
		"equalsOtherVar": func(v any, vl VarValues, args ...any) (bool, error) {
			if len(args) < 1 {
				return false, errors.New("equalsOtherVar需要一个变量参数")
			}

			// 从变量映射中获取其他变量的值
			otherVal := args[0].(int)
			// 比较$的值和其他变量的值
			return v == otherVal, nil
		},

		// 检查$的长度是否大于指定值，同时检查另一个变量是否有效
		"complexCheck": func(v any, vl VarValues, args ...any) (bool, error) {
			str, ok := v.(string) // $的值
			if !ok {
				return false, errors.New("complexCheck要求$是字符串")
			}

			if len(args) < 1 {
				return false, errors.New("complexCheck需要一个长度参数")
			}
			minLen, ok := args[0].(int)
			if !ok {
				return false, errors.New("complexCheck参数必须是整数")
			}

			// 检查另一个变量$status是否为"active"
			status, exists := vl["status"]
			if !exists {
				return false, errors.New("$status变量不存在")
			}
			statusStr, ok := status.(string)
			if !ok {
				return false, errors.New("$status必须是字符串")
			}

			// 综合判断
			return len(str) > minLen && statusStr == "active", nil
		},
	}
	// 比较函数映射表
	conf.compareFunctionMap = map[string]compareFunction{
		"len": func(v any, vl VarValues, args ...any) (any, error) {
			var vr = args[0]
			switch val := vr.(type) {
			case string:
				return len(val), nil
			case []int:
				return len(val), nil
			case []string:
				return len(val), nil
			default:
				return nil, errors.New("len函数不支持的类型")
			}
		},

		// 计算数值加倍的特殊函数
		"double": func(v any, vl VarValues, args ...any) (any, error) {
			num, ok := args[0].(int)
			if !ok {
				return nil, errors.New("double函数要求参数是整数")
			}
			return num * 2, nil
		},

		// 连接字符串的特殊函数
		"concat": func(v any, vl VarValues, args ...any) (any, error) {
			result := ""
			for _, arg := range args {
				argStr, ok := arg.(string)
				if !ok {
					return nil, errors.New("concat函数参数必须是字符串")
				}
				result += argStr
			}
			return result, nil
		},
	}
}

// 表达式测试
func TestParseIf(t *testing.T) {
	initFunc()
	// 定义所有测试用例
	testCases := []TestCase{
		// 1. 基础比较 - 数字
		{
			Name:     "数字大于比较",
			Expr:     "$ > 100",
			Vars:     VarValues{"$": 150},
			Expected: true,
		},
		{
			Name:     "数字小于等于比较",
			Expr:     "$a <= 50",
			Vars:     VarValues{"$": 0, "a": 60},
			Expected: false,
		},

		// 2. 基础比较 - 字符串
		{
			Name:     "字符串等于比较",
			Expr:     "$ == active",
			Vars:     VarValues{"$": "active"},
			Expected: true,
		},
		{
			Name:     "字符串不等于比较",
			Expr:     "$b != error",
			Vars:     VarValues{"$": 0, "b": "error"},
			Expected: false,
		},

		// 3. 列表判断 - in
		{
			Name:     "变量在字符串列表中",
			Expr:     "$ in (apple, banana, cherry)",
			Vars:     VarValues{"$": "banana"},
			Expected: true,
		},
		{
			Name:     "变量在数字列表中",
			Expr:     "$num in (10, 20, 30)",
			Vars:     VarValues{"$": 0, "num": 25},
			Expected: false,
		},

		// 4. 列表判断 - !in
		{
			Name:     "变量不在字符串列表中",
			Expr:     "$ !in (red, green, blue)",
			Vars:     VarValues{"$": "yellow"},
			Expected: true,
		},
		{
			Name:     "变量不在数字列表中",
			Expr:     "$count !in (5, 10, 15)",
			Vars:     VarValues{"$": 0, "count": 10},
			Expected: false,
		},

		// 5. 逻辑运算 - &&
		{
			Name:     "与运算都为真",
			Expr:     "$ > 5 && $a in (x, y, z)",
			Vars:     VarValues{"$": 10, "a": "y"},
			Expected: true,
		},
		{
			Name:     "与运算有一个为假",
			Expr:     "$ == test && $b !in (1,2,3)",
			Vars:     VarValues{"$": "test", "b": 2},
			Expected: false,
		},

		// 6. 逻辑运算 - ||
		{
			Name:     "或运算有一个为真",
			Expr:     "$ < 0 || $c == success",
			Vars:     VarValues{"$": 5, "c": "success"},
			Expected: true,
		},
		{
			Name:     "或运算都为假",
			Expr:     "$ in (a,b) || $d > 100",
			Vars:     VarValues{"$": "c", "d": 50},
			Expected: false,
		},

		// 7. 混合运算
		{
			Name:     "比较与列表混合",
			Expr:     "$ >= 100 && $status in (active, pending)",
			Vars:     VarValues{"$": 150, "status": "pending"},
			Expected: true,
		},
		// 8. 边界情况
		{
			Name:     "空列表（永远为假）",
			Expr:     "$ in ()",
			Vars:     VarValues{"$": "anything"},
			Expected: false,
		},
		{
			Name:     "变量与列表类型不匹配",
			Expr:     "$ in (10, 20, 30)",
			Vars:     VarValues{"$": "10"}, // 字符串vs数字
			Expected: false,
		},

		// 基础函数测试
		{
			Name:     "dollarInRange-在范围内",
			Expr:     "dollarInRange(10, 20)",
			Vars:     VarValues{"$": 15},
			Expected: true,
		},
		{
			Name:     "dollarInRange-超出范围",
			Expr:     "dollarInRange(5, 10)",
			Vars:     VarValues{"$": 12},
			Expected: false,
		},

		// 使用变量映射的函数测试
		{
			Name:     "equalsOtherVar-相等",
			Expr:     "equalsOtherVar($threshold)",
			Vars:     VarValues{"$": 100, "threshold": 100},
			Expected: true,
		},
		{
			Name:     "equalsOtherVar-不相等",
			Expr:     "equalsOtherVar($max)",
			Vars:     VarValues{"$": 50, "max": 100},
			Expected: false,
		},

		// 复杂函数测试
		{
			Name:     "complexCheck-满足条件",
			Expr:     "complexCheck(5)",
			Vars:     VarValues{"$": "hello world", "status": "active"},
			Expected: true,
		},
		{
			Name:     "complexCheck-不满足条件",
			Expr:     "complexCheck(100)",
			Vars:     VarValues{"$": "hello world", "status": "active"},
			Expected: false,
		},

		// 函数与其他表达式混合
		{
			Name:     "函数与比较混合",
			Expr:     "dollarInRange(0, 100) && $status == active",
			Vars:     VarValues{"$": 50, "status": "active"},
			Expected: true,
		},

		// 新增测试用例 - 更多边界情况和复杂场景
		{
			Name:     "字符串比较大于",
			Expr:     "$ > abc",
			Vars:     VarValues{"$": "def"},
			Expected: true,
		},
		{
			Name:     "字符串比较小于",
			Expr:     "$ < xyz",
			Vars:     VarValues{"$": "abc"},
			Expected: true,
		},
		{
			Name:     "多重逻辑运算",
			Expr:     "$ > 0 && $ < 100 || $ == 200",
			Vars:     VarValues{"$": 200},
			Expected: true,
		},

		// 错误情况测试
		{
			Name:     "无效变量名",
			Expr:     "$invalid > 10",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},
		{
			Name:     "缺少右值",
			Expr:     "$ > ",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},
		{
			Name:     "无效逻辑运算符",
			Expr:     "$ > 5 & $ < 10",
			Vars:     VarValues{"$": 7},
			HasError: true,
		},
		{
			Name:     "未定义函数",
			Expr:     "undefinedFunc(10)",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},
		{
			Name:     "缺少$变量",
			Expr:     "$a > 5",
			Vars:     VarValues{"$a": 10},
			HasError: true,
		},
		{
			Name:     "变量名不以$开头",
			Expr:     "var > 5",
			Vars:     VarValues{"$": 10, "var": 7},
			HasError: true,
		},
		{
			Name:     "类型不匹配比较",
			Expr:     "$ > 10",
			Vars:     VarValues{"$": "string"},
			HasError: true,
		},
		{
			Name:     "in列表语法错误-缺少括号",
			Expr:     "$ in (1,2,3",
			Vars:     VarValues{"$": 2},
			HasError: true,
		},
		{
			Name:     "!in列表语法错误-缺少括号",
			Expr:     "$ !in (1,2,3",
			Vars:     VarValues{"$": 2},
			HasError: true,
		},
		{
			Name:     "函数参数错误-dollarInRange缺少参数",
			Expr:     "dollarInRange(10)",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},
		{
			Name:     "函数参数错误-dollarInRange参数类型错误",
			Expr:     "dollarInRange(abc, def)",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},
		{
			Name:     "函数参数错误-equalsOtherVar缺少参数",
			Expr:     "equalsOtherVar()",
			Vars:     VarValues{"$": 5, "threshold": 10},
			HasError: true,
		},
		{
			Name:     "函数参数错误-complexCheck缺少参数",
			Expr:     "complexCheck()",
			Vars:     VarValues{"$": "hello", "status": "active"},
			HasError: true,
		},
		{
			Name:     "函数参数错误-complexCheck参数类型错误",
			Expr:     "complexCheck(abc)",
			Vars:     VarValues{"$": "hello", "status": "active"},
			HasError: true,
		},
		{
			Name:     "函数执行错误-complexCheck缺少$status变量",
			Expr:     "complexCheck(5)",
			Vars:     VarValues{"$": "hello"},
			HasError: true,
		},
		{
			Name:     "函数执行错误-dollarInRange变量类型错误",
			Expr:     "dollarInRange(5, 10)",
			Vars:     VarValues{"$": "not a number"},
			HasError: true,
		},
		{
			Name:     "空表达式",
			Expr:     "",
			Vars:     VarValues{"$": 5},
			HasError: true,
		},

		// len特殊函数测试
		{
			Name:     "len函数-字符串长度大于5",
			Expr:     "len($) > 5",
			Vars:     VarValues{"$": "hello world"},
			Expected: true,
		},
		{
			Name:     "len函数-数组长度等于3",
			Expr:     "len($arr) == 3",
			Vars:     VarValues{"$": 0, "arr": []int{1, 2, 3}},
			Expected: true,
		},

		// double特殊函数测试
		{
			Name:     "double函数-结果等于10",
			Expr:     "double($) == 10",
			Vars:     VarValues{"$": 5},
			Expected: true,
		},
		{
			Name:     "double函数-结果大于8",
			Expr:     "double($num) > 8",
			Vars:     VarValues{"$": 0, "num": 4},
			Expected: false,
		},

		// concat特殊函数测试
		{
			Name:     "concat函数-结果比较",
			Expr:     "concat($, _suffix) == test_suffix",
			Vars:     VarValues{"$": "test"},
			Expected: true,
		},

		// 特殊函数与普通函数混合
		{
			Name:     "特殊函数与普通函数混合",
			Expr:     "len($) > 3 && isValid($)",
			Vars:     VarValues{"$": "test"},
			HasError: true,
			Expected: true,
		},
	}
	var cond = NewConditional()
	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := cond.ParseIf("$", tc.Expr, tc.Vars)

			// 检查是否有错误
			if tc.HasError {
				if err == nil {
					t.Errorf("测试用例 '%s' 应该出现错误但没有错误", tc.Name)
				}
				return
			}

			if err != nil {
				t.Errorf("测试用例 '%s' 出现错误: %v", tc.Name, err)
				return
			}

			// 比较实际结果与预期结果
			if actual != tc.Expected {
				t.Errorf("测试用例 '%s' 失败\n表达式: %s\n变量: %v\n预期: %v, 实际: %v",
					tc.Name, tc.Expr, formatVars(tc.Vars), tc.Expected, actual)
			}
		})
	}
}

// 格式化变量输出
func formatVars(vars VarValues) string {
	var parts []string
	for k, v := range vars {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
