package gdb

import "errors"

// DbError 自定义错误
type DbError struct {
	isNotFound bool
	Code       string `json:"code"`
	Msg        string `json:"msg"`
}

func (e DbError) Error() string {
	return e.Code + e.Msg
}

func NewDbErr(code, msg string) DbError {
	return DbError{
		isNotFound: false,
		Code:       code,
		Msg:        msg,
	}
}

func ConvDbErr(err error) DbError {
	var dbErr DbError
	if errors.As(err, &dbErr) {
		return dbErr
	}
	return NewDbErr("", err.Error())
}

// ErrRecordNotFound  空数据错误
var ErrRecordNotFound = DbError{
	isNotFound: true,
	Code:       "",
	Msg:        "Record Not Found！",
}

// IsRecordNotFound 判断是否为空数据错误
func IsRecordNotFound(err error) bool {
	var dbErr DbError
	if errors.As(err, &dbErr) {
		return dbErr.isNotFound
	}
	return false
}
