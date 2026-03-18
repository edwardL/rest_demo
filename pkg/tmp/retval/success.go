package retval

import (
	hhtools "nwgit.gzhhit.com/BD/hhitcomm.git"
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
)

// Success1 返回成功1
func Success1(data ...[]map[string]interface{}) (uint8, any) {
	var retData = ftypes.ResultData01{
		Result: true,
		Code:   hhtools.CodeGetDataSuccess,
		Data:   make([]map[string]interface{}, 0),
		Msg:    hhtools.MsgGetDataSuccess,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 1, retData
}

// Success2 返回成功2
func Success2(data ...string) (uint8, any) {
	var retData = ftypes.ResultData02{
		Result: true,
		Code:   hhtools.CodeGetDataSuccess,
		Data:   "",
		Msg:    hhtools.MsgGetDataSuccess,
	}
	if len(data) > 0 {
		retData.Data = data[0]
	}
	return 2, retData
}

// Success3 返回成功3
func Success3(data ...map[string]string) (uint8, any) {
	var retData = ftypes.ResultData03{
		Result: true,
		Code:   hhtools.CodeGetDataSuccess,
		Data:   make(map[string]string),
		Msg:    hhtools.MsgGetDataSuccess,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 3, retData
}

// Success4 返回成功4
func Success4(data ...[]any) (uint8, any) {
	var retData = ftypes.ResultData04{
		Result: true,
		Code:   hhtools.CodeGetDataSuccess,
		Data:   make([]any, 0),
		Msg:    hhtools.MsgGetDataSuccess,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 4, retData
}

// Success5 返回成功5
func Success5(data ...any) (uint8, any) {
	var retData = ftypes.ResultData05{
		Result: true,
		Code:   hhtools.CodeGetDataSuccess,
		Data:   struct{}{},
		Msg:    hhtools.MsgGetDataSuccess,
	}
	if len(data) > 0 && data[0] != nil {
		retData.Data = data[0]
	}
	return 5, retData
}
