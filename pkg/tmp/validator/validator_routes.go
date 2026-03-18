package validator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type RoutesType string

const (
	RtRequired RoutesType = "required" // 必填 0值也不行
	RtEmpty    RoutesType = "empty"    // 空值
	RtEnum     RoutesType = "enum"     // 枚举
	RtEq       RoutesType = "eq"       // 小于
	RtLt       RoutesType = "lt"       // 小于
	RtElt      RoutesType = "elt"      // 小于等于
	RtGt       RoutesType = "gt"       // 大于
	RtEgt      RoutesType = "egt"      // 大于等于
	RtLen      RoutesType = "len"      // 长度 复合条件
	RtDp       RoutesType = "dp"       // 依赖字段 复合条件
	RtIf       RoutesType = "if"       // 表达式(只能使用同一个结构体下的字段) 复合条件
)

type ValidFunc func(v reflect.Value, fn string, info map[string]any, option string) (bool, error)

// 集成条件验证器
var routesValidator = map[RoutesType]ValidFunc{
	RtRequired: requiredValidator,
	RtEmpty:    emptyValidator,
	RtEnum:     enumValidator,
	RtEq:       eqValidator,
	RtLt:       ltValidator,
	RtElt:      eltValidator,
	RtGt:       gtValidator,
	RtEgt:      egtValidator,
}

// 复合条件验证器
var compoundRoutesValidator = map[RoutesType]ValidFunc{
	RtLen: lenValidator,
	RtDp:  dpValidator,
	RtIf:  ifValidator,
}

// GetValidFunc 获取验证器
func GetValidFunc(rt RoutesType) (ValidFunc, bool) {
	if vf, ok := routesValidator[rt]; ok {
		return vf, true
	}
	if vf, ok := compoundRoutesValidator[rt]; ok {
		return vf, true
	}
	return nil, false
}

// requiredValidator 必填验证器
func requiredValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if option != "" { // 多字段 有一个不为空则通过
		var fieldList = strings.Split(option, ",")
		fieldList = append(fieldList, fn)
		for _, field := range fieldList {
			if _, ok := info[field]; ok && !isZeroValue(info[field]) {
				return true, nil
			}
		}
	}
	return !isZeroValue(v), nil
}

// emptyValidator 必填验证器
func emptyValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	return isZeroValue(v), nil
}

// eqValidator 等于验证器
func eqValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if isZeroValue(v) {
		return true, nil // 空值不验证
	}
	var vStr = toString(v)
	return vStr == option, nil
}

// ltValidator 小于验证器
func ltValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	//if isZeroValue(v) {
	//	return true, nil // 空值不验证
	//}

	optionVal, err := strconv.ParseFloat(option, 64)
	if err != nil {
		return false, err // 配置错误，验证失败
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() < int64(optionVal), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() < uint64(optionVal), nil
	case reflect.Float32, reflect.Float64:
		return v.Float() < optionVal, nil
	default:
		return false, nil // 类型不支持
	}
}

// eltValidator 小于等于验证器
func eltValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	//if isZeroValue(v) {
	//	return true, nil // 空值不验证
	//}

	optionVal, err := strconv.ParseFloat(option, 64)
	if err != nil {
		return false, err // 配置错误，验证失败
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() <= int64(optionVal), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() <= uint64(optionVal), nil
	case reflect.Float32, reflect.Float64:
		return v.Float() <= optionVal, nil
	default:
		return false, nil // 类型不支持
	}
}

// gtValidator 大于验证器
func gtValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	//if isZeroValue(v) {
	//	return true, nil // 空值不验证
	//}

	optionVal, err := strconv.ParseFloat(option, 64)
	if err != nil {
		return false, err // 配置错误，验证失败
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > int64(optionVal), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() > uint64(optionVal), nil
	case reflect.Float32, reflect.Float64:
		return v.Float() > optionVal, nil
	default:
		return false, nil // 类型不支持
	}
}

// egtValidator 大于等于验证器
func egtValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	//if isZeroValue(v) {
	//	return true, nil // 空值不验证
	//}

	optionVal, err := strconv.ParseFloat(option, 64)
	if err != nil {
		return false, err // 配置错误，验证失败
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() >= int64(optionVal), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() >= uint64(optionVal), nil
	case reflect.Float32, reflect.Float64:
		return v.Float() >= optionVal, nil
	default:
		return false, nil // 类型不支持
	}
}

// lenValidator 长度验证器
func lenValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if isZeroValue(v) {
		return true, nil // 空值不验证
	}
	var sourceLen = 0
	switch v.Kind() {
	case reflect.String:
		sourceLen = len([]rune(v.String()))
	case reflect.Slice, reflect.Map, reflect.Array:
		sourceLen = v.Len()
	default:
		return false, nil // 类型不支持
	}
	var r = strings.Split(option, ",")
	var rv []string
	var vidFunc ValidFunc
	var ok bool
	var err error
	// 处理多个字条件
	for _, role := range r {
		rv = strings.Split(role, ":")
		if len(rv) == 2 {
			if vidFunc, ok = routesValidator[RoutesType(rv[0])]; ok {
				if ok, err = vidFunc(reflect.ValueOf(sourceLen), fn, info, rv[1]); !ok || err != nil {
					return ok, err
				}
			} else {
				return false, errors.New("子条件配置错误 错误条件->" + rv[0])
			}
		} else {
			return false, errors.New("子条件配置错误 格式应该为非符合类型 条件:值 当前->" + role)
		}
	}
	return true, nil
}

// enumValidator 枚举验证器
func enumValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if isZeroValue(v) {
		return true, nil // 空值不验证
	}

	// 将选项字符串按逗号分割成枚举值列表
	var enumValues = strings.Split(option, ",")

	// 获取字段值的字符串表示
	var valueStr string
	switch v.Kind() {
	case reflect.String:
		valueStr = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valueStr = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valueStr = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		valueStr = strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		valueStr = strconv.FormatBool(v.Bool())
	default:
		return false, nil // 不支持的类型
	}

	// 检查值是否在枚举列表中
	for _, enumValue := range enumValues {
		if strings.TrimSpace(enumValue) == valueStr {
			return true, nil
		}
	}

	return false, nil
}

// dpValidator 条件验证器
func dpValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if isZeroValue(v) {
		return true, nil // 空值不验证
	}
	var dep = string([]byte{defTagOptions.GroupOptSepRight, defTagOptions.GroupOptElseSep})
	var ifRole = strings.Split(option, dep)
	if len(ifRole) != 2 {
		return false, errors.New("条件配置错误 错误条件->" + option)
	}
	var fieldName = option[0:strings.Index(option, string(defTagOptions.GroupOptSepLeft))]
	var pCond = getCond(option)
	var childCond = getCond(option[strings.Index(option, dep):])
	var pVal reflect.Value
	var pAny any
	var ok bool
	if pAny, ok = info[fieldName]; ok {
		pVal = pAny.(reflect.Value)
		for pVal.Kind() == reflect.Ptr {
			pVal = pVal.Elem()
		}
	} else {
		return false, errors.New("字段配置错误 错误字段->" + fieldName)
	}

	// 判断前置条件是否符合
	var res, err = validatorField(pVal, fieldName, info, pCond)
	if err != nil {
		return false, errors.New("前置条件配置错误 错误条件->" + pCond)
	}
	if !res {
		return false, nil
	}

	// 满足前置条件
	return validatorField(v, fn, info, childCond)
}

// ifValidator 条件验证器
func ifValidator(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	if isZeroValue(v) {
		return true, nil // 空值不验证
	}
	option = getCond(option)
	var ifRes, err = NewConditional().ParseIf(fn, option, info)
	if err != nil {
		return false, err
	}
	return ifRes, nil
}

// validatorField 单字段校验
func validatorField(v reflect.Value, fn string, info map[string]any, option string) (bool, error) {
	// 判断前置条件是否符合
	var pRouteList = split(option, defTagOptions.ValidSep)
	var rv []string
	var vidFunc ValidFunc
	var ok bool
	var err error
	for _, route := range pRouteList {
		var r = strings.Split(route, ",")
		// 处理多个字条件
		for _, role := range r {
			rv = strings.Split(role, ":")
			if vidFunc, ok = routesValidator[RoutesType(rv[0])]; ok {
				if ok, err = vidFunc(v, fn, info, strings.Join(rv[1:], ":")); !ok || err != nil {
					return ok, err
				}
			} else {
				return false, errors.New("前置条件配置错误 错误条件->" + rv[0])
			}
		}
	}
	return true, nil
}

// isZeroValue 检查值是否为零值
func isZeroValue(val any) bool {
	var v reflect.Value
	if val == nil {
		return true
	}
	if rv, ok := val.(reflect.Value); ok {
		v = rv
	} else {
		v = reflect.ValueOf(val)
	}
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	case reflect.Invalid:
		return true
	default:
		// 未知类型默认为空
		return true
	}
}
