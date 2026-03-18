package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ConditionalFunction 普通函数类型
type ConditionalFunction func(v any, vl VarValues, args ...any) (bool, error)

// 比较函数类型：返回计算结果和可能的错误，结果可参与后续表达式
type compareFunction func(v any, vl VarValues, args ...any) (any, error)

// VarValues 变量值映射
type VarValues map[string]any

// Token 定义
type Token struct {
	Type  string   // "var"、"num"、"str"、"op"、"logic"、"in"、"not_in"、"func"、"special_func"
	Value string   // 变量名、运算符值或函数名
	Args  []string // 列表参数或函数参数
}

// Conditional 提交校验
type Conditional struct {
	indexKey string
}

// NewConditional 创建条件表达式校验器实例。
func NewConditional() *Conditional {
	return &Conditional{}
}

// ParseIf 解析表达式
func (c *Conditional) ParseIf(indexKey string, expr string, vars VarValues) (bool, error) {
	expr = strings.TrimSpace(expr)
	expr = strings.TrimPrefix(expr, "if ")
	c.indexKey = indexKey
	//if err := c.validateVariables(vars); err != nil {
	//	return false, err
	//}

	var varNames []string
	for name := range vars {
		varNames = append(varNames, name)
	}

	var tokens, err = c.tokenize(expr, varNames, conf.conditionalFunctionMap, conf.compareFunctionMap)
	if err != nil {
		return false, err
	}

	return c.evaluateTokens(tokens, vars, conf.conditionalFunctionMap, conf.compareFunctionMap)
}

// 验证变量是否符合规则
//func (c *Conditional) validateVariables(vars VarValues) error {
//	if len(vars) == 0 {
//		return errors.New("至少需要包含$变量")
//	}
//	if _, hasDollar := vars["$"]; !hasDollar {
//		return errors.New("必须包含$变量")
//	}
//	for name := range vars {
//		if name != "$" && !strings.HasPrefix(name, "$") {
//			return fmt.Errorf("变量 %s 必须以$开头", name)
//		}
//	}
//	return nil
//}

// 词法分析器：区分普通函数和比较函数
func (c *Conditional) tokenize(expr string, varNames []string, condFunc map[string]ConditionalFunction, compFunc map[string]compareFunction) ([]Token, error) {
	var tokens []Token
	var i = 0
	var start int
	var n = len(expr)
	var cf uint8
	var funcName, varName string
	for i < n {
		if expr[i] == ' ' || expr[i] == '\t' {
			i++
			continue
		}

		cf = expr[i]

		// 处理运算符
		if cf == '>' || cf == '<' || cf == '=' || cf == '!' {
			if i+1 < n && expr[i+1] == '=' {
				tokens = append(tokens, Token{Type: "op", Value: expr[i : i+2]})
				i += 2
			} else {
				tokens = append(tokens, Token{Type: "op", Value: string(cf)})
				i++
			}
			continue
		}

		// 处理逻辑运算符
		if cf == '&' || cf == '|' {
			if i+1 < n && expr[i+1] == cf {
				tokens = append(tokens, Token{Type: "logic", Value: expr[i : i+2]})
				i += 2
			} else {
				return nil, errors.New("无效的逻辑运算符: " + string(cf))
			}
			continue
		}

		// 处理函数调用：区分比较函数和普通函数
		if (cf >= 'a' && cf <= 'z') || (cf >= 'A' && cf <= 'Z') {
			start = i
			// 解析函数名
			for i < n {
				curr := expr[i]
				if (curr >= 'a' && curr <= 'z') ||
					(curr >= 'A' && curr <= 'Z') ||
					(curr >= '0' && curr <= '9') ||
					curr == '_' {
					i++
				} else {
					break
				}
			}
			funcName = expr[start:i]

			i = c.skipWhitespace(expr, i)
			if i < n && expr[i] == '(' {
				// 检查是否是比较函数
				if _, isSpecial := compFunc[funcName]; isSpecial {
					i++ // 跳过(
					i = c.skipWhitespace(expr, i)

					// 解析比较函数参数
					args, err := c.parseFunctionArgs(expr, &i, varNames)
					if err != nil {
						return nil, err
					}

					// 添加比较函数token
					tokens = append(tokens, Token{
						Type:  "special_func",
						Value: funcName,
						Args:  args,
					})
					continue
				} else {
					// 处理普通函数
					i++ // 跳过(
					i = c.skipWhitespace(expr, i)

					// 解析函数参数列表
					args, err := c.parseFunctionArgs(expr, &i, varNames)
					if err != nil {
						return nil, err
					}

					// 检查普通函数是否存在
					if _, exists := condFunc[funcName]; !exists {
						return nil, fmt.Errorf("未定义的函数: %s", funcName)
					}

					// 添加普通函数token
					tokens = append(tokens, Token{
						Type:  "func",
						Value: funcName,
						Args:  args,
					})
					continue
				}
			} else {
				// 不是函数调用，视为字符串常量
				tokens = append(tokens, Token{Type: "str", Value: funcName})
				continue
			}
		}

		// 处理变量
		if cf == '$' {
			start = i
			i++

			for i < n {
				curr := expr[i]
				if (curr >= 'a' && curr <= 'z') ||
					(curr >= 'A' && curr <= 'Z') ||
					(curr >= '0' && curr <= '9') ||
					curr == '_' {
					i++
				} else {
					break
				}
			}
			varName = expr[start:i]

			if !c.isValidVariable(varName, varNames) {
				return nil, fmt.Errorf("不允许的变量: %s", varName)
			}

			i = c.skipWhitespace(expr, i)
			if i+1 < n && expr[i] == '!' && strings.HasPrefix(expr[i:], "!in ") {
				i += 3
				i = c.skipWhitespace(expr, i)

				if i >= n || expr[i] != '(' {
					return nil, errors.New("!in表达式缺少'('")
				}
				i++

				args, err := c.parseListArgs(expr, &i)
				if err != nil {
					return nil, err
				}

				tokens = append(tokens, Token{
					Type:  "not_in",
					Value: varName,
					Args:  args,
				})
				continue
			} else if i+2 < n && strings.HasPrefix(expr[i:], "in ") {
				i += 2
				i = c.skipWhitespace(expr, i)

				if i >= n || expr[i] != '(' {
					return nil, errors.New("in表达式缺少'('")
				}
				i++

				args, err := c.parseListArgs(expr, &i)
				if err != nil {
					return nil, err
				}

				tokens = append(tokens, Token{
					Type:  "in",
					Value: varName,
					Args:  args,
				})
				continue
			} else {
				tokens = append(tokens, Token{Type: "var", Value: varName})
				continue
			}
		}

		// 处理数字
		if cf >= '0' && cf <= '9' {
			start = i
			for i < n && expr[i] >= '0' && expr[i] <= '9' {
				i++
			}
			numVal := expr[start:i]
			tokens = append(tokens, Token{Type: "num", Value: numVal})
			continue
		}

		// 处理字符串常量
		start = i
		for i < n {
			curr := expr[i]
			if curr == ' ' || curr == '\t' ||
				curr == '>' || curr == '<' || curr == '=' || curr == '!' ||
				curr == '&' || curr == '|' || curr == '(' || curr == ')' ||
				curr == ',' {
				break
			}
			i++
		}
		if i > start {
			strVal := expr[start:i]
			tokens = append(tokens, Token{Type: "str", Value: strVal})
			continue
		}

		return nil, errors.New("无效的字符: " + string(cf))
	}

	return tokens, nil
}

// 解析函数参数列表
func (c *Conditional) parseFunctionArgs(expr string, i *int, varNames []string) ([]string, error) {
	var args []string
	var n = len(expr)
	var start int
	var varName string
	var numVal, strVal string
	for *i < n {
		*i = c.skipWhitespace(expr, *i)
		if *i >= n {
			break
		}

		// 检查是否是闭合括号
		if expr[*i] == ')' {
			*i++ // 跳过)
			return args, nil
		}

		// 解析单个参数（变量或常量）
		start = *i
		if expr[*i] == '$' {
			// 变量参数
			*i++
			for *i < n {
				curr := expr[*i]
				if (curr >= 'a' && curr <= 'z') ||
					(curr >= 'A' && curr <= 'Z') ||
					(curr >= '0' && curr <= '9') ||
					curr == '_' {
					*i++
				} else {
					break
				}
			}
			varName = expr[start:*i]
			if !c.isValidVariable(varName, varNames) {
				return nil, fmt.Errorf("函数参数包含无效变量: %s", varName)
			}
			args = append(args, varName)
		} else if expr[*i] >= '0' && expr[*i] <= '9' {
			// 数字常量参数
			for *i < n && expr[*i] >= '0' && expr[*i] <= '9' {
				*i++
			}
			numVal = expr[start:*i]
			args = append(args, numVal)
		} else {
			// 字符串常量参数
			for *i < n {
				curr := expr[*i]
				if curr == ' ' || curr == ',' || curr == ')' {
					break
				}
				*i++
			}
			strVal = expr[start:*i]
			args = append(args, strVal)
		}

		// 处理参数分隔符
		*i = c.skipWhitespace(expr, *i)
		if *i < n && expr[*i] == ',' {
			*i++ // 跳过逗号
		} else if *i < n && expr[*i] != ')' {
			return nil, errors.New("函数参数列表格式错误，缺少逗号分隔")
		}
	}
	return nil, errors.New("函数调用缺少闭合的')'")
}

// 辅助函数：跳过空白字符
func (c *Conditional) skipWhitespace(expr string, i int) int {
	n := len(expr)
	for i < n && (expr[i] == ' ' || expr[i] == '\t') {
		i++
	}
	return i
}

// 解析列表参数
func (c *Conditional) parseListArgs(expr string, i *int) ([]string, error) {
	var args []string
	var start = *i
	var n = len(expr)
	var arg string
	for *i < n {
		if expr[*i] == ')' {
			if *i > start {
				arg = strings.TrimSpace(expr[start:*i])
				args = append(args, arg)
			}
			*i++
			return args, nil
		} else if expr[*i] == ',' {
			arg = strings.TrimSpace(expr[start:*i])
			args = append(args, arg)
			*i++
			start = *i
			*i = c.skipWhitespace(expr, *i)
		} else {
			*i++
		}
	}

	return nil, errors.New("列表表达式缺少闭合的')'")
}

// 检查变量是否有效
func (c *Conditional) isValidVariable(varName string, allowed []string) bool {
	if varName == "$" {
		varName = c.indexKey
	} else {
		varName = varName[1:]
	}
	for _, name := range allowed {
		if varName == name {
			return true
		}
	}
	return false
}

// 计算token列表的结果
func (c *Conditional) evaluateTokens(tokens []Token, vars VarValues, condFunc map[string]ConditionalFunction, compFunc map[string]compareFunction) (bool, error) {
	var atomResults []bool
	var logicOps []string
	var currentAtoms = make([]Token, 0)
	var res bool
	var err error
	for _, t := range tokens {
		if t.Type == "logic" {
			res, err = c.evaluateAtom(currentAtoms, vars, condFunc, compFunc)
			if err != nil {
				return false, err
			}
			atomResults = append(atomResults, res)
			logicOps = append(logicOps, t.Value)
			currentAtoms = []Token{}
		} else {
			currentAtoms = append(currentAtoms, t)
		}
	}

	if len(currentAtoms) > 0 {
		res, err = c.evaluateAtom(currentAtoms, vars, condFunc, compFunc)
		if err != nil {
			return false, err
		}
		atomResults = append(atomResults, res)
	}

	if len(atomResults) == 0 {
		return false, errors.New("表达式为空")
	}
	var finalResult = atomResults[0]
	for i, op := range logicOps {
		if i+1 >= len(atomResults) {
			break
		}
		switch op {
		case "&&":
			finalResult = finalResult && atomResults[i+1]
		case "||":
			finalResult = finalResult || atomResults[i+1]
		}
	}

	return finalResult, nil
}

// 计算原子表达式（支持比较函数结果参与逻辑）
func (c *Conditional) evaluateAtom(atom []Token, vars VarValues, condFunc map[string]ConditionalFunction, compFunc map[string]compareFunction) (bool, error) {
	// 处理包含比较函数的表达式（如 len($) > 5）
	if len(atom) == 3 {
		var left = atom[0]
		var op = atom[1]
		var right = atom[2]
		var rightVal any

		// 左侧可能是比较函数或变量
		var leftVal any
		var err error

		if left.Type == "special_func" {
			// 计算比较函数结果
			leftVal, err = c.evaluateCompFunction(left, vars, compFunc)
			if err != nil {
				return false, err
			}
		} else if left.Type == "var" {
			// 变量值
			var val, ok = vars[c.getArgValueKey(left.Value)]
			if !ok {
				return false, fmt.Errorf("未定义的变量: %s", left.Value)
			}
			leftVal = val
		} else {
			return false, errors.New("比较表达式左值必须是变量或比较函数")
		}

		// 验证运算符
		if op.Type != "op" {
			return false, errors.New("比较表达式中间必须是运算符")
		}

		// 解析右值
		rightVal, err = c.parseRightValue(right)
		if err != nil {
			return false, err
		}

		// 比较结果
		return c.compareValues(leftVal, op.Value, rightVal)
	}

	// 处理普通函数
	if len(atom) == 1 {
		var t = atom[0]
		if t.Type == "func" {
			// 获取函数
			var fn, exists = condFunc[t.Value]
			if !exists {
				return false, fmt.Errorf("未定义的函数: %s", t.Value)
			}

			// 获取$变量的值（第一个参数）
			var dollarVal, ok = vars[c.indexKey]
			if !ok {
				return false, errors.New("未找到$变量")
			}

			// 解析从规则传递的参数
			var ruleArgs []any
			for _, arg := range t.Args {
				argVal, err := c.getArgValue(arg, vars)
				if err != nil {
					return false, err
				}
				ruleArgs = append(ruleArgs, argVal)
			}

			// 调用普通函数
			return fn(dollarVal, vars, ruleArgs...)
		}

		// 处理in和!in
		if t.Type == "in" || t.Type == "not_in" {
			var varVal, ok = vars[c.getArgValueKey(t.Value)]
			if !ok {
				return false, fmt.Errorf("未定义的变量: %s", t.Value)
			}
			var inList = false
			var varStr, isStr = varVal.(string)
			var varNum, isNum = varVal.(int)
			var argNum int
			var err error
			for _, arg := range t.Args {
				argNum, err = strconv.Atoi(arg)
				if err == nil {
					if isNum && varNum == argNum {
						inList = true
						break
					}
				} else {
					if isStr && varStr == arg {
						inList = true
						break
					}
				}
			}

			if t.Type == "in" {
				return inList, nil
			} else {
				return !inList, nil
			}
		}
	}

	return false, errors.New("无效的原子表达式格式")
}

// 计算比较函数的值
func (c *Conditional) evaluateCompFunction(token Token, vars VarValues, compFunc map[string]compareFunction) (any, error) {
	// 获取比较函数
	var fn, exists = compFunc[token.Value]
	if !exists {
		return nil, fmt.Errorf("未定义的比较函数: %s", token.Value)
	}

	// 获取$变量的值（第一个参数）
	dollarVal, ok := vars[c.indexKey]
	if !ok {
		return nil, errors.New("未找到$变量")
	}

	// 解析函数参数
	var args []any
	for _, arg := range token.Args {
		argVal, err := c.getArgValue(arg, vars)
		if err != nil {
			return nil, err
		}
		args = append(args, argVal)
	}

	// 调用比较函数并返回结果
	return fn(dollarVal, vars, args...)
}

// 获取参数值（可能是变量或常量）
func (c *Conditional) getArgValue(arg string, vars VarValues) (any, error) {
	if strings.HasPrefix(arg, "$") {
		// 变量参数
		val, ok := vars[c.getArgValueKey(arg)]
		if !ok {
			return nil, fmt.Errorf("未定义的变量: %s", arg)
		}
		return val, nil
	}

	// 尝试解析为数字
	num, err := strconv.Atoi(arg)
	if err == nil {
		return num, nil
	}

	// 否则视为字符串
	return arg, nil
}

// getArgValueKey 转换变量名字
func (c *Conditional) getArgValueKey(arg string) string {
	if arg == "$" {
		return c.indexKey
	}
	return arg[1:]
}

// 解析右值（数字或字符串）
func (c *Conditional) parseRightValue(right Token) (any, error) {
	if right.Type == "num" {
		return strconv.Atoi(right.Value)
	} else if right.Type == "str" {
		return right.Value, nil
	} else if right.Type == "special_func" {
		return nil, errors.New("右值不支持比较函数")
	}
	return nil, errors.New("无效的右值类型: " + right.Type)
}

// 比较两个值
func (c *Conditional) compareValues(a any, op string, b any) (bool, error) {
	if av, ok := a.(reflect.Value); ok {
		a = av.Interface()
	}
	if bv, ok := b.(reflect.Value); ok {
		b = bv.Interface()
	}
	var aType = fmt.Sprintf("%T", a)
	var bType = fmt.Sprintf("%T", b)
	if aType != bType {
		return false, fmt.Errorf("类型不匹配: %s 和 %s", aType, bType)
	}

	switch op {
	case "==":
		return a == b, nil
	case "!=":
		return a != b, nil
	case ">":
		switch aVal := a.(type) {
		case int:
			return aVal > b.(int), nil
		case string:
			return aVal > b.(string), nil
		}
	case "<":
		switch aVal := a.(type) {
		case int:
			return aVal < b.(int), nil
		case string:
			return aVal < b.(string), nil
		}
	case ">=":
		switch aVal := a.(type) {
		case int:
			return aVal >= b.(int), nil
		case string:
			return aVal >= b.(string), nil
		}
	case "<=":
		switch aVal := a.(type) {
		case int:
			return aVal <= b.(int), nil
		case string:
			return aVal <= b.(string), nil
		}
	}
	return false, errors.New("不支持的运算符: " + op)
}
