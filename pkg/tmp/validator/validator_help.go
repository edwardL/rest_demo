package validator

import (
	"fmt"
	fframe "nwgit.gzhhit.com/BD/hhitframe.git/frame"
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
)

// Validate 结构体验证
func Validate(s any) *Error {
	return NewValidator().ValidAll(true).Validate(s)
}

// ValidateByRoute 自定义规则验证 s = 结构体或者map[string]any
func ValidateByRoute(s any, route map[string]ValidateRules) *Error {
	return NewValidator().ValidAll(true).ValidateByRoutes(s, route)
}

// ValidateByRoutes 自定义规则数组验证 s = []结构体或者[]map[string]any
func ValidateByRoutes[T any](s []T, route map[string]ValidateRules) *Error {
	var vali = NewValidator().ValidAll(true)
	var err *Error = NewErr()
	var eItem *Error
	for _, v := range s {
		eItem = vali.ValidateByRoutes(v, route)
		if eItem != nil {
			err.AppendVidErr(eItem.ValidationErrors...)
		}
	}
	return err.Errors()
}

// Validates 结构数组体验证
func Validates[T any](s []T) *Error {
	var vali = NewValidator().ValidAll(true)
	var err *Error = NewErr()
	var eItem *Error
	for _, v := range s {
		eItem = vali.Validate(v)
		if eItem != nil {
			err.AppendVidErr(eItem.ValidationErrors...)
		}
	}
	return err.Errors()
}

// NewTsChecker 构建一个空的ts检查器
func NewTsChecker() TsChecker {
	return map[int]bool{}
}

// TsChecker ts检查器
type TsChecker map[int]bool

// Check 检查ts是否有效
func (tc *TsChecker) Check(info *ftypes.SitkWebUserInfo, ts int) bool {
	if *tc == nil {
		*tc = make(map[int]bool)
	}
	check, ok := (*tc)[ts]
	if ok {
		return check
	}
	if ts != info.TID && !fframe.DeterInTsLst(info, fmt.Sprint(ts)) {
		(*tc)[ts] = false
		return false
	}
	(*tc)[ts] = true
	return true
}
