package validator

import (
	"errors"
	"reflect"
	"strings"
)

// ValidateRules 校验规则
type ValidateRules struct {
	Rules  string `json:"rules"` // 规则
	Option string `json:"value"` // 参数
	Code   string `json:"code"`
	Msg    string `json:"msg"`
}

// GetCode 获取错误码
func (vr ValidateRules) GetCode(defCode string) string {
	if vr.Code != "" {
		return vr.Code
	}
	return defCode
}

// GetMsg 获取错误信息
func (vr ValidateRules) GetMsg() string {
	if vr.Msg != "" {
		return vr.Msg
	}
	if msg, ok := conf.validatorErrCodeMsg[vr.Code]; ok {
		return msg
	}
	if vr.Option != "" {
		return vr.Rules + ":" + vr.Option
	}
	return vr.Rules
}

// ValidateRulesContainer 校验规则
type ValidateRulesContainer struct {
	childStruct any             // 子结构体
	Field       string          `json:"field"`
	Rules       []ValidateRules `json:"rules"` // 规则
}

// ValidateRulesList 校验规则列表
type ValidateRulesList []ValidateRulesContainer

type Validator struct {
	prefix        string // 父字段
	validAll      bool   // 校验全部
	defErrCode    string // 默认错误码
	validTag      *TagOptions
	validateRules ValidateRulesList
	fieldMap      map[string]any
}

// NewValidator 创建一个验证器
func NewValidator(defErrCode ...string) *Validator {
	var defCode string = conf.defErrCode
	if len(defErrCode) > 0 {
		defCode = defErrCode[0]
	}
	return &Validator{
		validTag:      &defTagOptions,
		defErrCode:    defCode,
		validateRules: make([]ValidateRulesContainer, 0),
		fieldMap:      make(map[string]any),
	}
}

// SetValidTag 设置标签参数
func (vid *Validator) SetValidTag(vt *TagOptions) {
	vid.validTag = vt
}

// ValidAll 是否校验全部
func (vid *Validator) ValidAll(va bool) *Validator {
	vid.validAll = va
	return vid
}

// SetPrefix 设置标签参数
func (vid *Validator) SetPrefix(p string) *Validator {
	vid.prefix = p
	return vid
}

// Validate 结构体校验
func (vid *Validator) Validate(s any) *Error {
	var val = reflect.ValueOf(s)
	var typ = reflect.TypeOf(s)
	var err error
	var validationErrors = NewErr()
	// 如果是指针，获取指向的元素
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 确保是结构体
	if val.Kind() != reflect.Struct {
		return validationErrors.SetSysErr(errors.New("参数必须是结构体")).Errors()
	}

	// 解析规则
	err = vid.ParseStruct(typ, val)
	if err != nil {
		return validationErrors.SetSysErr(err).Errors()
	}

	return vid.ValidateData(val)
}

// ValidateData 数据校验 请先解析规则
func (vid *Validator) ValidateData(s any) *Error {
	var val reflect.Value
	if rv, ok := s.(reflect.Value); ok {
		val = rv
	} else {
		val = reflect.ValueOf(s)
	}
	var err error
	var validationErrors = NewErr()
	// 如果是指针，获取指向的元素
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保是结构体
	if val.Kind() != reflect.Struct {
		return validationErrors.SetSysErr(errors.New("参数必须是结构体")).Errors()
	}

	// 无规则直接返回
	if len(vid.validateRules) == 0 {
		return nil
	}

	var value reflect.Value
	var ok bool
	// 开始校验
	for _, v := range vid.validateRules {
		if v.childStruct != nil { // 嵌套结构体校验
			value = v.childStruct.(reflect.Value)
			for value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
			if !value.IsValid() {
				continue
			}
			if isZeroValue(value) {
				value = reflect.New(value.Type()).Elem()
			}
			var childValidationErrors = NewValidator(vid.defErrCode).SetPrefix(vid.genField(vid.prefix, v.Field)).Validate(value.Interface())
			if childValidationErrors != nil && childValidationErrors.Errors() != nil {
				validationErrors.AppendVidErr(childValidationErrors.ValidationErrors...)
				if !vid.validAll {
					return validationErrors.Errors()
				}
			}
		}

		if _, ok = vid.fieldMap[v.Field]; !ok {
			continue
		}
		value = vid.fieldMap[v.Field].(reflect.Value)

		// 校验规则
		for _, vr := range v.Rules {
			validatorFunc, exists := GetValidFunc(RoutesType(vr.Rules))
			if !exists {
				return validationErrors.SetSysErr(errors.New("规则不存在" + vr.Rules)).Errors()
			}
			ok, err = validatorFunc(value, v.Field, vid.fieldMap, vr.Option)
			if err != nil {
				return validationErrors.SetSysErr(err).Errors()
			}
			if !ok {
				validationErrors.AppendVidErr(ValidationError{
					Field: vid.genField(vid.prefix, v.Field),
					Code:  vr.GetCode(vid.defErrCode),
					Msg:   vr.GetMsg(),
				})
				if !vid.validAll {
					return validationErrors.Errors()
				}
			}
		}
	}
	return validationErrors.Errors()
}

// ValidateByRoutes 数据校验
func (vid *Validator) ValidateByRoutes(s any, route map[string]ValidateRules) *Error {
	var err error
	var validationErrors = NewErr()
	if s == nil {
		return validationErrors.SetSysErr(errors.New("参数不能为空")).Errors()
	}
	// 解析规则
	err = vid.parseRoutes(route)
	if err != nil {
		return validationErrors.SetSysErr(err).Errors()
	}

	// 无规则直接返回
	if len(vid.validateRules) == 0 {
		return nil
	}

	switch d := s.(type) {
	case map[string]any:
		vid.fieldMap = d
	case *map[string]any:
		vid.fieldMap = *d
	default:
		var t = reflect.TypeOf(s)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return validationErrors.SetSysErr(errors.New("参数必须是结构体或map[string]any")).Errors()
		}
		var e = vid.structToMap(s)
		if e != nil {
			return validationErrors.SetSysErr(e).Errors()
		}
	}

	var value reflect.Value
	var vAny any
	var ok bool
	// 开始校验
	for _, v := range vid.validateRules {

		vAny, ok = vid.fieldMap[v.Field]
		if !ok {
			continue
		}
		if vAny == nil { // 空值
			vAny = ""
		}
		value = reflect.ValueOf(vAny)

		// 校验规则
		for _, vr := range v.Rules {
			validatorFunc, exists := GetValidFunc(RoutesType(vr.Rules))
			if !exists {
				return validationErrors.SetSysErr(errors.New("规则不存在" + vr.Rules)).Errors()
			}
			ok, err = validatorFunc(value, v.Field, vid.fieldMap, vr.Option)
			if err != nil {
				return validationErrors.SetSysErr(err).Errors()
			}
			if !ok {
				validationErrors.AppendVidErr(ValidationError{
					Field: vid.genField(vid.prefix, v.Field),
					Code:  vr.GetCode(vid.defErrCode),
					Msg:   vr.GetMsg(),
				})
				if !vid.validAll {
					return validationErrors.Errors()
				}
			}
		}
	}
	return validationErrors.Errors()
}

// structToMap 结构体转map
func (vid *Validator) structToMap(s any) error {
	var typ = reflect.TypeOf(s)
	var val = reflect.ValueOf(s)
	// 如果是指针，获取指向的元素
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	var field reflect.StructField
	var jsonTag, fieldName string
	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		field = typ.Field(i)
		// 获取字段名（JSON标签）
		jsonTag = field.Tag.Get("json")
		fieldName = field.Name
		if jsonTag != "" {
			fieldName = jsonTag
		}
		vid.fieldMap[fieldName] = val.Field(i)
	}
	return nil
}

// RegRoutes 注册验证规则 map[字段名][]ValidateRules
func (vid *Validator) RegRoutes(vr map[string][]ValidateRules) {
	for k, v := range vr {
		vid.AppendRoutes(k, v)
	}
}

// AppendRoutes 注册验证规则
func (vid *Validator) AppendRoutes(fieldName string, vr []ValidateRules) {
	vid.validateRules = append(vid.validateRules, ValidateRulesContainer{
		Field: fieldName,
		Rules: vr,
	})
}

// ParseStruct 解析结构体验证规则
func (vid *Validator) ParseStruct(s any, v any) error {
	var typ reflect.Type
	var val reflect.Value
	if t, ok := s.(reflect.Type); ok {
		typ = t
		val = v.(reflect.Value)
	} else {
		typ = reflect.TypeOf(s)
		val = reflect.ValueOf(v)
	}

	// 如果是指针，获取指向的元素
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	// 确保是结构体
	if typ.Kind() != reflect.Struct {
		return errors.New("参数必须是结构体")
	}

	var field reflect.StructField
	var vTag, jsonTag, fieldName, vCode, vMsg string
	var vTagArr, vCodeArr, vMsgArr []string
	var routesList []ValidateRules
	var fTyp reflect.Type
	var rl []string
	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		field = typ.Field(i)
		fTyp = field.Type
		if fTyp.Kind() == reflect.Ptr {
			fTyp = fTyp.Elem()
		}

		if fTyp.Kind() == reflect.Struct { // 匿名字段
			var value reflect.Value = val.Field(i)
			for value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
			vid.validateRules = append(vid.validateRules, ValidateRulesContainer{
				childStruct: value,
				Field:       fieldName,
			})
			continue
		}

		// 获取字段名（JSON标签）
		jsonTag = field.Tag.Get("json")
		fieldName = field.Name
		if jsonTag != "" {
			fieldName = jsonTag
		}

		// 获取标签
		vTag = field.Tag.Get(vid.validTag.Valid)
		if vTag == "" { // 无需校验
			continue
		}
		routesList = make([]ValidateRules, 0)

		// 获取错误码和错误消息
		vCode = field.Tag.Get(vid.validTag.ValidCode)
		vMsg = field.Tag.Get(vid.validTag.ValidMsg)

		vTagArr = split(vTag, vid.validTag.ValidSep)
		vCodeArr = split(vCode, vid.validTag.ValidSep)
		vMsgArr = split(vMsg, vid.validTag.ValidSep)
		for j, v := range vTagArr {
			rl = split(v, vid.validTag.ValidOptSep)
			routesList = append(routesList, ValidateRules{
				Rules:  vid.getArrVal(&rl, 0),
				Option: vid.getArrValAll(rl, 1),
				Code:   vid.getArrVal(&vCodeArr, j),
				Msg:    vid.getArrVal(&vMsgArr, j),
			})
		}
		vid.validateRules = append(vid.validateRules, ValidateRulesContainer{
			Field: fieldName,
			Rules: routesList,
		})
		vid.fieldMap[fieldName] = val.Field(i)
	}

	return nil
}

// parseStruct 解析验证规则
func (vid *Validator) parseRoutes(routes map[string]ValidateRules) error {
	var vTag, fieldName, vCode, vMsg string
	var vTagArr, vCodeArr, vMsgArr, rl []string
	var routesList []ValidateRules
	var vRoute ValidateRules
	// 遍历所有字段
	for fieldName, vRoute = range routes {
		routesList = make([]ValidateRules, 0)

		// 获取错误码和错误消息
		vCode = vRoute.Code
		vMsg = vRoute.Msg
		vTag = vRoute.Rules
		vTagArr = split(vTag, vid.validTag.ValidSep)
		vCodeArr = split(vCode, vid.validTag.ValidSep)
		vMsgArr = split(vMsg, vid.validTag.ValidSep)
		for j, v := range vTagArr {
			rl = split(v, vid.validTag.ValidOptSep)
			routesList = append(routesList, ValidateRules{
				Rules:  vid.getArrVal(&rl, 0),
				Option: vid.getArrValAll(rl, 1),
				Code:   vid.getArrVal(&vCodeArr, j),
				Msg:    vid.getArrVal(&vMsgArr, j),
			})
		}
		vid.validateRules = append(vid.validateRules, ValidateRulesContainer{
			Field: fieldName,
			Rules: routesList,
		})
	}
	return nil
}

// getArrVal 数组取值
func (vid *Validator) getArrVal(arr *[]string, index int) string {
	if index >= len(*arr) {
		return ""
	}
	var arrVal = (*arr)[index]
	if arrVal == "<" {
		(*arr)[index] = (*arr)[index-1]
	}
	return (*arr)[index]
}

// getArrVal 数组取值
func (vid *Validator) getArrValAll(arr []string, startIndex int) string {
	if startIndex >= len(arr) {
		return ""
	}
	return strings.Join(arr[startIndex:], string(vid.validTag.ValidOptSep))
}

// genField 构建值
func (vid *Validator) genField(prefix, field string) string {
	if prefix == "" {
		return field
	}
	return prefix + "." + field
}
