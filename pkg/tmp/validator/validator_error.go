package validator

import (
	"fmt"
	"strings"
)

// NewErr 获取错误
func NewErr() *Error {
	return &Error{}
}

// Error 包含所有校验错误
type Error struct {
	Err error // 非参数错误
	ValidationError
	ValidationErrors []ValidationError
}

// ValidationError 表示单个字段的校验错误
type ValidationError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
}

// Error 实现error接口
func (ve *Error) Error() string {
	if len(ve.ValidationErrors) == 0 {
		return ""
	}
	var errs []string
	if ve.Err != nil {
		errs = append(errs, ve.Err.Error())
	}
	for _, err := range ve.ValidationErrors {
		errs = append(errs, fmt.Sprintf("%s(%s): %s", err.Field, err.Code, err.Msg))
	}
	return strings.Join(errs, "; ")
}

// SetSysErr 设置系统错误
func (ve *Error) SetSysErr(err error) *Error {
	ve.Err = err
	return ve
}

// AppendVidErr 追加vid错误
func (ve *Error) AppendVidErr(vidErr ...ValidationError) {
	ve.ValidationErrors = append(ve.ValidationErrors, vidErr...)
	if ve.Code == "" && ve.Msg == "" && len(ve.ValidationErrors) > 0 {
		ve.ValidationError = ve.ValidationErrors[0]
	}
}

// Errors 获取错误
func (ve *Error) Errors() *Error {
	if ve.Err == nil && len(ve.ValidationErrors) == 0 {
		return nil
	}
	return ve
}
