package retval

import (
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
)

// Err1 返回错误1
func Err1(code, msg string, data ...[]map[string]interface{}) (uint8, any) {
	var retData = ftypes.ResultData01{
		Result: false,
		Code:   code,
		Data:   make([]map[string]interface{}, 0),
		Msg:    msg,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 1, retData
}

// Err2 返回错误2
func Err2(code, msg string, data ...string) (uint8, any) {
	var retData = ftypes.ResultData02{
		Result: false,
		Code:   code,
		Data:   "",
		Msg:    msg,
	}
	if len(data) > 0 {
		retData.Data = data[0]
	}
	return 2, retData
}

// Err3 返回错误3
func Err3(code, msg string, data ...map[string]string) (uint8, any) {
	var retData = ftypes.ResultData03{
		Result: false,
		Code:   code,
		Data:   make(map[string]string),
		Msg:    msg,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 3, retData
}

// Err4 返回错误4
func Err4(code, msg string, data ...[]interface{}) (uint8, any) {
	var retData = ftypes.ResultData04{
		Result: false,
		Code:   code,
		Data:   make([]interface{}, 0),
		Msg:    msg,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 4, retData
}

// Err5 返回错误5
func Err5(code, msg string, data ...any) (uint8, any) {
	var retData = ftypes.ResultData05{
		Result: false,
		Code:   code,
		Data:   struct{}{},
		Msg:    msg,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 5, retData
}
