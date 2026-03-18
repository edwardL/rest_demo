package gdb

import (
	"context"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"strings"
)

var (
	// mysqlKeywords sql查询关键词会使用 `包裹`
	mysqlKeywords = map[string]struct{}{
		"add":    struct{}{},
		"alter":  struct{}{},
		"case":   struct{}{},
		"create": struct{}{},
		"delete": struct{}{},
		"drop":   struct{}{},
		"from":   struct{}{},
		"group":  struct{}{},
		"insert": struct{}{},
		"select": struct{}{},
		"update": struct{}{},
		"where":  struct{}{},
		"status": struct{}{},
		"name":   struct{}{},
		"port":   struct{}{},
		"desc":   struct{}{},
		"key":    struct{}{},
		"index":  struct{}{},
		"skip":   struct{}{},
		"order":  struct{}{},
		"tenant": struct{}{},
		"user":   struct{}{},
		"count":  struct{}{},
	}
	// mysqlFieldChat 字段允许的字符
	mysqlFieldChat = []byte{'.', '`', '_'}
	// zeroDelField // 0值忽略的字段
	zeroValIgnoreField = []string{"id", "ts", "create_time", "update_time"}
)

type Cond = string

const (
	CondAnd Cond = "AND"
	CondOr  Cond = "OR"
)

const (
	// FieldRemove val 填这个会忽略这个字段
	FieldRemove = "_field_remove_"
)

// RawBody 原样拼接
type RawBody string

// Raw 保持原样拼接 适配场景 age = age+1
func Raw(q string) RawBody {
	return RawBody(q)
}

// RawArg 原样参数
type RawArg struct {
	val any
}

// RawAny 保持原样参数 适配场景 ipv6var > []byte
func RawAny(q any) RawArg {
	return RawArg{val: q}
}

// Result 返回结果
type Result struct {
	Sql  *strings.Builder
	Args []any
}

// CompSql 获取完整SQL
func (r Result) CompSql() string {
	if r.Args == nil {
		return r.Sql.String()
	}
	var err error
	var compSql = new(strings.Builder)
	var compSqlArr = strings.Split(r.Sql.String(), "?")
	var argsStr string
	var val any
	var argsMaxKey = len(r.Args) - 1
	for k, sqlItem := range compSqlArr {
		compSql.WriteString(sqlItem)
		argsStr = ""
		if k > argsMaxKey {
			continue
		}
		val = r.Args[k]
		if val == nil {
			argsStr = "NULL"
		} else {
			switch vt := val.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				argsStr = conv.ToString(vt)
			case bool:
				if vt {
					argsStr = "1"
				} else {
					argsStr = "0"
				}
			case []byte:
				argsStr = "'" + string(vt) + "'"
			default:
				argsStr = "'" + strings.ReplaceAll(conv.ToString(vt), "'", "\\'") + "'"
			}
		}
		compSql.WriteString(argsStr)
	}
	if err != nil {
		defLog.CtxError(context.Background(), "创建完整sql错误：", err)
	}
	return compSql.String()
}
