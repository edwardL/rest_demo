package gdbtmp

import (
	"context"
	"reflect"
)

// MapAssignmentFunc 结构体转map 赋值方法 用于自定义类型处理
type MapAssignmentFunc func(field reflect.StructField, fieldValue reflect.Value, itemMap map[string]any, tag GdbTag) (next bool)

// StructAssignmentFunc map转结构体 赋值方法 用于自定义类型处理
type StructAssignmentFunc func(f reflect.Value, val any) (next bool, err error)
type Conf struct {
	Log                GdbLog                           // 日志
	LogLevel           LogLevel                         // 日志级别
	WriteHhDbLog       bool                             // 是否打印框架日志
	WriteLog           bool                             // 是否打印日志
	WriteErrSql        bool                             // 是否打印错误sql 避免打印敏感信息
	WriteCompSql       bool                             // 是否打印完整sql 避免打印敏感信息
	DbKeywords         []string                         // 忽略关键字
	ZeroValIgnoreField []string                         // 零值忽略字段
	TraceIdFunc        func(ctx context.Context) string // 上下文处理方法
	MapAssignment      MapAssignmentFunc                // 结构体转map 赋值方法
	StructAssignment   StructAssignmentFunc             // map转结构体 赋值方法
}

var conf = &Conf{
	Log:          nil,
	LogLevel:     InfoLogLevel,
	WriteHhDbLog: false,
	WriteLog:     true,
	WriteErrSql:  true,
	WriteCompSql: true,
	TraceIdFunc:  nil,
}

// Option 选项函数类型：接收Server指针，用于配置参数
type Option func(c *Conf)

func Init(opts ...Option) {
	for _, opt := range opts {
		opt(conf)
	}
	if conf.Log != nil {
		SetDefLog(conf.Log)
	}
	SetLevel(conf.LogLevel)
	if conf.TraceIdFunc != nil {
		SetTraceFunc(conf.TraceIdFunc)
	}
	if conf.DbKeywords != nil {
		for _, k := range conf.DbKeywords {
			mysqlKeywords[k] = struct{}{}
		}
	}
	if conf.ZeroValIgnoreField != nil {
		zeroValIgnoreField = append(zeroValIgnoreField, conf.ZeroValIgnoreField...)
	}
}
