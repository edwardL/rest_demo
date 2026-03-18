package validator

import (
	"errors"
)

// 增加验证场景 适配多规则

type ValidScene string
type Field = string

var validatorScenes = map[ValidScene]map[Field]ValidateRules{}

// RegisterScenes 注册场景
func RegisterScenes(scene ValidScene, routes map[Field]ValidateRules) {
	validatorScenes[scene] = routes
}

// ValidateByScene s = 结构体或者map[string]any
func ValidateByScene(s any, scenes ...ValidScene) *Error {
	var err = NewErr()
	for _, scene := range scenes {
		if r, ok := validatorScenes[scene]; ok {
			var ve = ValidateByRoute(s, r)
			if ve.Errors() != nil {
				err.AppendVidErr(ve.ValidationErrors...)
			}
		} else {
			return NewErr().SetSysErr(errors.New("未找到验证场景！"))
		}
	}
	if err.Errors() == nil {
		return nil
	}
	return err
}

// ValidatesByScene s = []结构体或者[]map[string]any
func ValidatesByScene[T any](s []T, scenes ...ValidScene) *Error {
	var err = NewErr()
	for _, scene := range scenes {
		if r, ok := validatorScenes[scene]; ok {
			var ve = ValidateByRoutes(s, r)
			if ve.Errors() != nil {
				err.AppendVidErr(ve.ValidationErrors...)
			}
		} else {
			return NewErr().SetSysErr(errors.New("未找到验证场景！"))
		}
	}
	if err.Errors() == nil {
		return nil
	}
	return err
}
