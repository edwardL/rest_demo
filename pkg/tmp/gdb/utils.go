package gdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"reflect"
	"strings"
	"sync/atomic"
)

// MapToMapAny 任意map[string]T类型转 map[string]any
func MapToMapAny(d any) (map[string]any, error) {
	var m = make(map[string]any)
	switch dt := d.(type) {
	case map[string]any:
		return dt, nil
	case *map[string]any:
		return *dt, nil
	case map[string]RawBody:
		for k, v := range dt {
			m[k] = v
		}
		return m, nil
	case map[string]string:
		for k, v := range dt {
			m[k] = v
		}
		return m, nil
	}
	// 无匹配 使用反射统一处理所有 map[string]T 类型
	var v = reflect.ValueOf(d)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map && v.Type().Key().Kind() == reflect.String {
		// 检查是否为 map[string]T 格式
		var result = make(map[string]any, v.Len())
		var iter = v.MapRange()
		var key string
		var value any
		for iter.Next() {
			key = iter.Key().String()
			value = iter.Value().Interface()
			result[key] = value
		}
		return result, nil
	}
	return nil, errors.New("not a map[string]T")
}

// MapsToMapsAny 任意[]map[string]T类型转 []map[string]any
func MapsToMapsAny(d any) ([]map[string]any, error) {
	var ml = make([]map[string]any, 0)
	if d == nil {
		return ml, nil
	}
	var err = errors.New("not a []map[string]any|string")
	// 处理常用类型
	switch dt := d.(type) {
	case []map[string]any:
		return dt, nil
	case []map[string]RawBody:
		var result = make([]map[string]any, len(dt))
		for i, m := range dt {
			resultMap := make(map[string]any, len(m))
			for k, v := range m {
				resultMap[k] = v
			}
			result[i] = resultMap
		}
		return result, nil
	case []map[string]string:
		var result = make([]map[string]any, len(dt))
		for i, m := range dt {
			resultMap := make(map[string]any, len(m))
			for k, v := range m {
				resultMap[k] = v
			}
			result[i] = resultMap
		}
		return result, nil
	case []any:
		var res = make([]map[string]any, len(dt))
		var me error
		for i, m := range dt {
			var mm map[string]any
			mm, me = MapToMapAny(m)
			if me != nil {
				return nil, err
			}
			res[i] = mm
		}
		return res, nil
	}
	return nil, err
}

// StructDbField 获取结构体里的字段名
// appendAlias 是否追加别名 空字符串 自动识别模型别名 不传不拼接别名
func StructDbField(s any, appendAlias ...string) []string {
	var alias string
	var v reflect.Type
	if s != nil {
		v = reflect.TypeOf(s)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	} else {
		return nil
	}
	if len(appendAlias) > 0 {
		alias = appendAlias[0]
		if alias == "" {
			if v.Implements(modelAliasFaceType) {
				alias = (v.(ModelAliasFace)).TableAlias()
			}
		}
	}
	var fi = conv.GetStructFieldList(v)
	var fieldArr = make([]string, 0, len(fi))
	for _, f := range fi {
		if alias != "" {
			fieldArr = append(fieldArr, fmt.Sprintf("`%s`.`%s`", alias, f))
		} else {
			fieldArr = append(fieldArr, f)
		}
	}
	return fieldArr
}

// Dump 格式化打印
func Dump(d any) {
	DumpCtx(setLogCallDepthCtx(context.Background(), 4), d)
}

// DumpCtx 格式化打印
func DumpCtx(ctx context.Context, d any) {
	ctx = setLogCallDepthCtx(ctx, getLogCallDepthCtx(ctx)) // 修正打印的行号
	var jsonByte []byte
	switch v := d.(type) {
	case string:
		jsonByte = []byte(v)
	case []byte:
		jsonByte = v
	default:
		marshal, err := json.Marshal(d)
		if err != nil {
			defLog.CtxInfof(ctx, "\n%v", d)
			return
		}
		jsonByte = marshal
	}
	if len(jsonByte) == 0 {
		jsonByte = []byte("{}")
	}
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, jsonByte, "", "\t")
	if err != nil {
		defLog.CtxInfof(ctx, "\n%v", string(jsonByte))
		return
	}
	defLog.CtxInfof(ctx, "\n%v", string(prettyJSON.Bytes()))
}

func strToArr(str string, sep string, delEmpty bool) []string {
	arr := make([]string, 0)
	strArr := strings.Split(str, sep)
	if !delEmpty {
		return strArr
	}
	for _, v := range strArr {
		if v != "" {
			arr = append(arr, v)
		}
	}
	return arr
}

// replaceIndex 替换出现在指定位置的 ?
func replaceIndex(query string, pos int, new string) (q string, err error) {
	idx := strIndex(query, "?", pos)
	if idx == -1 {
		return "", errors.New("[ " + query + " ] 参数和占位符不匹配")
	}
	// 校验截取范围，避免索引越界
	if idx+1 > len(query) {
		return "", errors.New("占位符位置超出字符串长度")
	}
	return query[0:idx] + new + query[idx+1:], nil
}

// strIndex 字符串第几次出现的位置
func strIndex(s, substr string, pos int) int {
	if pos == 0 {
		pos = 1
	}
	if pos == 1 {
		return strings.Index(s, substr)
	}
	lastI := -1
	st := s
	for i := 1; i <= pos; i++ {
		idx := strings.Index(st, substr)
		if idx == -1 {
			return idx
		}
		st = st[idx+1:]
		lastI += idx + 1
	}
	return lastI
}

// genQueryList 构建数组查询 [1,2,3] => ?,?,?
func genQueryList[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](query string, arr []T, phIndex *atomic.Int32) (newQuery string, args []any, err error) {
	var argsLen = len(arr)
	if argsLen == 0 {
		return query, nil, nil
	}
	query, err = genPrePil(query, int(phIndex.Load()), argsLen)
	if err != nil {
		return query, nil, err
	}
	phIndex.Add(int32(argsLen))
	for _, av := range arr {
		args = append(args, av)
	}
	return query, args, nil
}

// genQueryGroupList 构建数组查询 [1,2,3],[1,2,3] => (?,?,?),(?,?,?)
func genQueryGroupList[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string](query string, arr [][]T, phIndex *atomic.Int32) (newQuery string, args []any, err error) {
	var argsLen = len(arr)
	var childLen = 0
	if len(arr) > 0 {
		childLen = len(arr[0])
	}
	var i int
	query, err = genPrePilGroup(query, int(phIndex.Load()), argsLen, childLen)
	if err != nil {
		return "", nil, err
	}
	phIndex.Add(int32(argsLen * childLen))
	for _, av := range arr {
		for i = 0; i < childLen; i++ {
			args = append(args, av[i])
		}
	}
	return query, args, nil
}

// genQueryGroupList 构建数组查询 [1,2,3],[1,2,3] => (?,?,?),(?,?,?)
func genQueryGroupAnyList(query string, arr [][]any, phIndex *atomic.Int32) (newQuery string, args []any, err error) {
	var argsLen = len(arr)
	var childLen = 0
	if len(arr) > 0 {
		childLen = len(arr[0])
	}
	query, err = genPrePilGroup(query, int(phIndex.Load()), argsLen, childLen)
	if err != nil {
		return "", args, err
	}
	phIndex.Add(int32(argsLen * childLen))
	for _, v := range arr {
		args = append(args, v[:childLen]...)
	}
	return query, args, nil
}

// genPrePil 构建预编译sql where id = ?
func genPrePil(query string, pos, len int) (q string, err error) {
	var zw = strings.Repeat(", ?", len)
	zw = strings.TrimLeft(zw, ", ")
	return replaceIndex(query, pos, zw)
}

// genPrePilGroup 构建预编译sql组 where (id,ts) = (?,?)
func genPrePilGroup(query string, pos, len, size int) (q string, err error) {
	var zw = strings.Repeat(", ?", size)
	zw = strings.TrimLeft(zw, ", ")
	zw = "(" + zw + ")"
	zw = strings.Repeat(", "+zw, len)
	zw = strings.TrimLeft(zw, ", ")
	return replaceIndex(query, pos, zw)
}

// arrToArrList 构建数组列表
func arrToArrList(a []any) ([][]any, error) {
	var res [][]any = make([][]any, len(a))
	for k, arg := range a {
		switch arr := arg.(type) {
		case []int:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int8:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int16:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []int64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint8:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint16:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []uint64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []string:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []float32:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []float64:
			for _, v := range arr {
				res[k] = append(res[k], v)
			}
		case []bool:
			for _, v := range arr {
				if v {
					res[k] = append(res[k], 1)
				} else {
					res[k] = append(res[k], 0)
				}
			}
		case []any:
			res[k] = arr
		default:
			return nil, errors.New("参数类型错误,不支持的嵌套切片类型")
		}
	}
	return res, nil
}

// 解析类型并调用TableName方法
func getTableNameRecursive(val reflect.Value) (string, error) {
	var tableName string
	switch val.Kind() {
	case reflect.Ptr:
		// 处理指针类型（如*Aaa、**Aaa）
		if val.IsNil() {
			// 空指针：尝试通过类型创建非空实例（避免解引用空指针）
			var ptrType = val.Type()
			var elemType = ptrType.Elem()
			// 创建指针指向的元素的实例（如*Aaa的元素是Aaa，创建Aaa{}）
			var newElem = reflect.New(elemType).Elem()
			return getTableNameRecursive(newElem)
		}
		// 非空指针：先检查自身是否实现接口，若实现则直接调用
		if val.Type().Implements(modelFaceType) {
			tableName = (val.Interface().(ModelFace)).TableName()
			return tableName, nil
		}
		// 若自身不实现，解引用后递归（如**Aaa -> *Aaa）
		return getTableNameRecursive(val.Elem())
	case reflect.Slice, reflect.Array:
		// 处理切片/数组（如[]Aaa、[]*Aaa、*[]Aaa等）
		var elemType = val.Type().Elem()
		// 创建元素类型的实例（用于检查和调用方法）
		var elemVal reflect.Value
		if elemType.Kind() == reflect.Ptr {
			// 元素是指针类型（如[]*Aaa）：创建指针实例（&Aaa{}）
			elemVal = reflect.New(elemType).Elem()
		} else {
			// 元素是值类型（如[]Aaa）：创建值实例（Aaa{}）
			elemVal = reflect.New(elemType).Elem()
		}
		// 递归检查元素类型
		return getTableNameRecursive(elemVal)
	default:
		// 处理值类型（如Aaa、Bbb）
		// 检查值类型是否实现接口（值接收者情况）
		if val.Type().Implements(modelFaceType) {
			tableName = (val.Interface().(ModelFace)).TableName()
			return tableName, nil
		}
		// 检查指针类型是否实现接口（指针接收者情况）
		var ptrVal = reflect.New(val.Type()) // 创建指针实例（&Aaa{}）
		if ptrVal.Type().Implements(modelFaceType) {
			tableName = (ptrVal.Interface().(ModelFace)).TableName()
			return tableName, nil
		}
		// 均不实现
		return "", fmt.Errorf("类型 %s 及其指针均不实现ModelFace", val.Type())
	}
}

// InArr 判断是否在数组里
func InArr[T comparable](v T, arr []T) bool {
	for _, val := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// As 减少外部 + 号拼接操作
// As("a","b") = a b
// As("a","AS b") = a AS b
func As(str ...string) string {
	return strings.Join(str, " ")
}

// EscapeLike 转义like 特殊字符
// EscapeLike("模糊查询的内容","拼接前面","拼接后面")
// 例如 EscapeLike("模糊查询的内容","","%") -> 模糊查询的内容%
// 例如 EscapeLike("模糊查询的内容","%","") -> %模糊查询的内容
// 默认转义字符为 前后都有%
func EscapeLike(str string, like ...string) string {
	var likeStart, likeEnd string
	if len(like) == 0 {
		likeStart, likeEnd = "%", "%"
	} else {
		likeStart = like[0]
		if len(like) > 1 {
			likeEnd = like[1]
		}
	}
	return likeStart + strings.NewReplacer(
		"\\", "\\\\", // 转义反斜杠
		"%", "\\%", // 转义百分号
		"_", "\\_", // 转义下划线
	).Replace(str) + likeEnd
}

// GetTableName 获取表名
func GetTableName(tb any) string {
	var err error
	var table string
	switch tbt := tb.(type) {
	case string:
		return tbt
	case ModelFace:
		var val = reflect.ValueOf(tbt)
		if !conv.IsNilValue(val) {
			return tbt.TableName()
		}
	}
	table, err = getTableNameRecursive(reflect.ValueOf(tb))
	if err != nil || table == "" {
		return ""
	}
	return table
}

// IsSqlChar mysql排序合法字符串
func IsSqlChar(b byte) bool {
	// 大小写字母
	if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
		return true
	}
	// 数字
	if b >= '0' && b <= '9' {
		return true
	}
	// 字符
	for _, fc := range mysqlFieldChat {
		if b == fc {
			return true
		}
	}
	return false
}

// ValidateSqlChar 校验字符串是否排序合法字符
func ValidateSqlChar(bs []byte) bool {
	for _, b := range bs {
		if !IsSqlChar(b) {
			return false
		}
	}
	return true
}

// 校验字符串是否为 DESC/ASC（不区分大小写，纯字符串处理兼容byte逻辑）
func validateSortDirection(s string) bool {
	var upperDir = strings.ToUpper(s)
	return upperDir == "DESC" || upperDir == "ASC"
}

// ValidateOrderParam 校验排序参数是否合法
func ValidateOrderParam(input string, fieldReplace map[string]string) (orderStr string, illegalField []string, err error) {
	var commaParts = conv.SplitTrim(input, ",")
	// 无有效内容，返回非法
	if len(commaParts) == 0 {
		return orderStr, illegalField, nil
	}
	var replaceFlag = len(fieldReplace) > 0
	var ok bool
	var orderField string
	var sortField []string
	var orderRes []string
	illegalField = make([]string, 0)
	var orderItem strings.Builder
	for _, targetStr := range commaParts {
		// 按空格分割（使用Fields，自动忽略连续空格/首尾空格，底层基于byte遍历）
		var spaceParts = strings.Fields(targetStr)
		var spaceLen = len(spaceParts)
		if spaceLen > 2 {
			return orderStr, illegalField, errors.New("排序参数异常:" + targetStr)
		}
		orderItem.Reset()
		if replaceFlag {
			if orderField, ok = fieldReplace[spaceParts[0]]; ok {
				if orderField != "" {
					sortField = append(sortField, orderField)
					orderItem.WriteString(orderField)
					if spaceLen > 1 {
						orderItem.WriteString(" ")
						// 仅DESC/ASC，不区分大小写
						if !validateSortDirection(spaceParts[1]) {
							return orderStr, illegalField, errors.New("排序规则只能为DESC/ASC")
						}
						orderItem.WriteString(spaceParts[1])
					}
					orderRes = append(orderRes, orderItem.String())
				}
			} else {
				illegalField = append(illegalField, spaceParts[0])
			}
		} else {
			// 根据空格分割后的长度进行校验
			switch spaceLen {
			case 1:
				sortField = append(sortField, spaceParts[0])
				// 校验仅包含字母、点、数字、下划线
				if !ValidateSqlChar([]byte(spaceParts[0])) {
					return orderStr, illegalField, errors.New("排序字段仅能包含字母、点、数字、下划线")
				}
			case 2:
				sortField = append(sortField, spaceParts[0])
				// 校验仅包含字母、点、数字、下划线
				if !ValidateSqlChar([]byte(spaceParts[0])) {
					return orderStr, illegalField, errors.New("排序字段仅能包含字母、点、数字、下划线")
				}
				// 仅DESC/ASC，不区分大小写
				if !validateSortDirection(spaceParts[1]) {
					return orderStr, illegalField, errors.New("排序规则只能为DESC/ASC")
				}
			default:
				return orderStr, illegalField, errors.New("排序参数非法：" + input)
			}
		}
	}
	if replaceFlag {
		return strings.Join(orderRes, ","), illegalField, nil
	}
	return input, illegalField, nil
}
