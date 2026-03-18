package log

import (
	"context"
	"io"
	"log"
	"strings"
)

// 配置方法

// SetLogFlag 设置标准库 logger 的输出标志位。
func SetLogFlag(flag int) {
	logFlag = flag
}

// SetLogPrefix 设置各日志级别的前缀样式。
func SetLogPrefix(p Prefix) {
	prefix = p
}

// SetLevel 设置当前日志输出级别。
func SetLevel(lv Level) {
	logLevel = lv
}

// SetLogLevel 通过字符串设置日志输出级别。
func SetLogLevel(lv string) {
	switch strings.ToLower(lv) {
	case "debug":
		SetLevel(LevelDebug)
	case "info":
		SetLevel(LevelInfo)
	case "error":
		SetLevel(LevelError)
	case "warn":
		SetLevel(LevelWarn)
	case "close":
		SetLevel(LevelClose)
	}
}

// SetDebugWrite 设置调试级别日志输出目标。
func SetDebugWrite(writers ...io.Writer) {
	LogDebug = log.New(io.MultiWriter(writers...), prefix.Debug, logFlag)
}

// SetInfoWrite 设置信息级别日志输出目标。
func SetInfoWrite(writers ...io.Writer) {
	LogInfo = log.New(io.MultiWriter(writers...), prefix.Info, logFlag)
}

// SetWarnWrite 设置告警级别日志输出目标。
func SetWarnWrite(writers ...io.Writer) {
	LogWarn = log.New(io.MultiWriter(writers...), prefix.Warn, logFlag)
}

// SetErrorWrite 设置错误级别日志输出目标。
func SetErrorWrite(writers ...io.Writer) {
	LogError = log.New(io.MultiWriter(writers...), prefix.Error, logFlag)
}

// SetCtxFunc 设置上下文追踪信息提取函数。
func SetCtxFunc(f CtxFunc) {
	ctxFunc = f
}

// GetTraceCtx 返回携带 traceId 的上下文。
func GetTraceCtx(ctx context.Context, traceId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxTraceIdKey, traceId)
}
