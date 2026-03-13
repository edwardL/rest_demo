package errcode

import "reflect"

type ErrCode int

const (
	ErrUserLoginDisabled ErrCode = 4009 //登录重试太多次了，30s之后在尝试吧
)

func (i ErrCode) Error() string {
	return reflect.ValueOf(i).String()
}
