package validator

import "errors"

// 配置
type validatorConf struct {
	defErrCode             string                         // 默认错误码
	defErrMsg              string                         // 默认错误信息
	conditionalFunctionMap map[string]ConditionalFunction // 条件验证函数映射表 直接返回结果 func(当前值,全部值映射,条件参数)
	compareFunctionMap     map[string]compareFunction     // 比较函数映射表 返回的结果会进行后续比较 func(当前值,全部值映射,条件参数)
	validatorErrCodeMsg    map[string]string              // 错误码以及错误信息
}

// 默认配置
var conf = &validatorConf{
	defErrCode: "",
	defErrMsg:  "参数错误",
	conditionalFunctionMap: map[string]ConditionalFunction{
		"true": func(v any, vl VarValues, args ...any) (bool, error) {
			if len(args) != 1 {
				return false, errors.New("true函数只能用一个入参数")
			}
			return trueFunc(args[0])
		},
		"false": func(v any, vl VarValues, args ...any) (bool, error) {
			if len(args) != 1 {
				return false, errors.New("true函数只能用一个入参数")
			}
			var b, e = trueFunc(args[0])
			return !b, e
		},
	},
	compareFunctionMap: map[string]compareFunction{
		"len": func(v any, vl VarValues, args ...any) (any, error) {
			if len(args) != 1 {
				return false, errors.New("len函数只能用一个入参数")
			}
			return lenFunc(args[0])
		},
	},
	validatorErrCodeMsg: map[string]string{},
}
