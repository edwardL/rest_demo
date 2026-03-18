package gdb

import (
	"context"
	"time"
)

type Conf struct {
	Log                GdbLog                           // 日志
	LogLevel           LogLevel                         // 日志级别
	WriteHhDbLog       bool                             // 是否打印框架日志
	WriteLog           bool                             // 是否打印日志
	WriteErrSql        bool                             // 是否打印错误sql 避免打印敏感信息
	WriteCompSql       bool                             // 是否打印完整sql 避免打印敏感信息
	EmptyError         *DbError                         // 空错误
	DbKeywords         []string                         // 忽略关键字
	DbFieldChar        []byte                           // 允许的字符
	ZeroValIgnoreField []string                         // 零值忽略字段
	TraceIdFunc        func(ctx context.Context) string // 上下文处理方法
	DbConvInitPtr      bool                             // 初始化指针
	TimeLocation       *time.Location                   // 时区
	DriveMap           map[string]NewWrapperFunc        // 创建构建sql
	DriveType          string                           // 驱动类型 默认mysql
}

var conf = &Conf{
	Log:          nil,
	LogLevel:     InfoLogLevel,
	WriteHhDbLog: false,
	WriteLog:     true,
	WriteErrSql:  true,
	WriteCompSql: true,
	TraceIdFunc:  nil,
	TimeLocation: time.Local,
	DriveMap:     make(map[string]NewWrapperFunc),
	DriveType:    "mysql",
}

// Option 选项函数类型：接收Server指针，用于配置参数
type Option func(c *Conf)

// Init 按选项初始化 gdb 全局配置。
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
	if conf.DbFieldChar != nil {
		mysqlFieldChat = append(mysqlFieldChat, conf.DbFieldChar...)
	}
	if conf.ZeroValIgnoreField != nil {
		zeroValIgnoreField = append(zeroValIgnoreField, conf.ZeroValIgnoreField...)
	}
	if conf.EmptyError != nil {
		ErrRecordNotFound = *conf.EmptyError
	}
}
