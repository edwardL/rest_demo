package retval

import (
	hhtools "nwgit.gzhhit.com/BD/hhitcomm.git"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/log"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
)

type inter interface {
	// 分页总数字段允许的整数类型
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// SuccessPageView 构建“分页查询成功”的统一返回结果，并按请求视图白名单过滤字段。
//
// 输入参数：
//   - info: 当前请求上下文信息；当包含 ReqViewInfo.SearchViewInfo.FieldList 时，会按字段白名单过滤输出。
//   - totalNums: 分页总条数（支持各类有符号/无符号整数类型）。
//   - pageData: 当前页原始数据切片（结构体切片或可被 conv.StructToMaps 处理的数据）。
//
// 输出结果：
//   - 第一个返回值 uint8:
//   - 成功时固定为 1。
//   - 失败时固定为 1。
//   - 第二个返回值 any:
//   - 成功时：
//     map 结构（ResultData01）：
//     {
//     "result": true,
//     "code":   hhtools.CodeGetDataSuccess,
//     "msg":    hhtools.MsgGetDataSuccess,
//     "data": []map[string]any{
//     {
//     "page_data":  []map[string]any, // 过滤后的分页数据
//     "total_nums": totalNums,         // 总条数
//     },
//     },
//     }
//   - 失败时（例如数据转换失败）：
//     map 结构（ResultData01）：
//     {
//     "result": false,
//     "code":   hhtools.CodeReDataFormatError,
//     "msg":    hhtools.MsgReDataFormatError,
//     "data":   []map[string]any{},
//     }
func SuccessPageView[I inter, T any](info *types.SitkWebCliInfo, totalNums I, pageData []T) (uint8, any) {
	// dataMaps 分页数据 map 列表
	var dataMaps []map[string]any
	// err 数据转换错误
	var err error

	dataMaps, err = conv.StructToMaps(pageData, false)
	if err != nil {
		log.OutputErrorf(1, "分页data处理失败 %v", err)
		return Err1(hhtools.CodeReDataFormatError, hhtools.MsgReDataFormatError)
	}

	// 按视图白名单过滤字段
	dataMaps = filterMapListByView(info, dataMaps)

	// resData 统一分页返回结构
	var resData = []map[string]any{
		{
			"page_data":  dataMaps,
			"total_nums": totalNums,
		},
	}
	return Success1(resData)
}

// SuccessView 构建“列表/单对象查询成功”的统一返回结果，并按请求视图白名单过滤字段。
//
// 输入参数：
//   - info: 当前请求上下文信息；当包含 ReqViewInfo.SearchViewInfo.FieldList 时，会按字段白名单过滤输出。
//   - data: 原始业务数据（通常为结构体或可被 conv.StructToMap 处理的数据）。
//
// 输出结果：
//   - 第一个返回值 uint8:
//   - 成功时固定为 1。
//   - 失败时固定为 1。
//   - 第二个返回值 any:
//   - 成功时：
//     map 结构（ResultData01）：
//     {
//     "result": true,
//     "code":   hhtools.CodeGetDataSuccess,
//     "msg":    hhtools.MsgGetDataSuccess,
//     "data":   []map[string]any{...}, // 过滤后数据，统一以数组形式返回
//     }
//   - 失败时（例如数据转换失败）：
//     map 结构（ResultData01）：
//     {
//     "result": false,
//     "code":   hhtools.CodeReDataFormatError,
//     "msg":    hhtools.MsgReDataFormatError,
//     "data":   []map[string]any{},
//     }
func SuccessView[T any](info *types.SitkWebCliInfo, data T) (uint8, any) {
	// dataMap 单条数据 map
	var dataMap map[string]any
	// err 数据转换错误
	var err error
	dataMap, err = conv.StructToMap(data, false)
	if err != nil {
		log.OutputErrorf(1, "data处理失败 %v", err)
		return Err1(hhtools.CodeReDataFormatError, hhtools.MsgReDataFormatError)
	}

	// dataMaps 统一为数组返回
	var dataMaps = []map[string]any{dataMap}
	dataMaps = filterMapListByView(info, dataMaps)
	return Success1(dataMaps)
}

// filterMapListByView 按查询视图中的字段白名单过滤 map 列表中的字段。
//
// 输入参数：
//   - info: 请求上下文；当 info/ReqViewInfo 为空，或 FieldList 为空时，不做过滤。
//   - dataList: 待过滤的数据列表，每个元素为一条记录的字段 map。
//
// 输出结果：
//   - 返回过滤后的 []map[string]any。
//   - 当无视图限制时，返回原 dataList（原地引用）。
//   - 过滤过程为“原地删除”非白名单字段，调用方可感知到 dataList 的内容变化。
func filterMapListByView(info *types.SitkWebCliInfo, dataList []map[string]any) []map[string]any {
	if info == nil || info.ReqViewInfo == nil {
		return dataList
	}
	if len(info.ReqViewInfo.SearchViewInfo.FieldList) == 0 {
		return dataList
	}

	// allowFields 白名单字段集合
	var allowFields = make(map[string]struct{}, len(info.ReqViewInfo.SearchViewInfo.FieldList))
	var field string
	for field = range info.ReqViewInfo.SearchViewInfo.FieldList {
		allowFields[field] = struct{}{}
	}

	var i int
	for i = 0; i < len(dataList); i++ {
		for field = range dataList[i] {
			if _, ok := allowFields[field]; !ok {
				delete(dataList[i], field)
			}
		}
	}
	return dataList
}
