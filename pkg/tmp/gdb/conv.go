package gdb

import (
	"context"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
)

// toMap any 转 map[string]any
func toMap(ctx context.Context, d any, dbConvInitPtr bool) (m map[string]any, isStruct bool, err error) {
	m, err = MapToMapAny(d)
	if err == nil {
		return m, false, nil
	}
	// 不是map类型那就是结构体
	err = nil // 是结构体 清空错误
	// 内部会对d进行类型判断 无需在调用前重复判断
	m, err = conv.StructToMap(d, dbConvInitPtr)
	if err == nil {
		return m, true, nil
	}
	if err.Error() != conv.NotStructErrMsg {
		defLog.CtxWarn(ctx, "StructToMap 失败 使用默认转换：", err)
	}
	err = nil
	err = conv.AToB(d, &m)
	return m, true, err
}

// toMaps any 转 []map[string]any
func toMaps(ctx context.Context, d any, dbConvInitPtr bool) (m []map[string]any, isStruct bool, err error) {
	m, err = MapsToMapsAny(d)
	if err == nil {
		return m, false, nil
	}
	m, err = conv.StructToMaps(d, dbConvInitPtr)
	if err == nil {
		return m, true, nil
	}
	if err.Error() != conv.NotStructErrMsg {
		defLog.CtxWarn(ctx, "StructToMaps 失败 使用默认转换：", err)
	}
	err = nil
	err = conv.AToB(d, &m)
	return m, true, err
}
