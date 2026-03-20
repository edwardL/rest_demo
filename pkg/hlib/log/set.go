package log

import (
	"context"
	"io"
	"log"
	"strings"
)

func SetLogFlag(flag int) {
	logFlag = flag
}

func SetLogPrefix(p Prefix) {
	prefix = p
}

func SetLevel(lv Level) {
	logLevel = lv
}

func SetLogLevel(lv string) {
	switch strings.ToLower(lv) {
	case "debug":
		SetLevel(LevelDebug)
	case "info":
		SetLevel(LevelInfo)
	case "warn":
		SetLevel(LevelWarn)
	case "error":
		SetLevel(LevelError)
	case "close":
		SetLevel(LevelClose)
	}
}

// SetDebugWrite 设置调试级别日志输出目标
func SetDebugWrite(writers ...io.Writer) {
	LogDebug = log.New(io.MultiWriter(writers...), prefix.Debug, logFlag)
}

// SetInfoWrite 设置信息级别日志输出目标
func SetInfoWrite(writers ...io.Writer) {
	LogInfo = log.New(io.MultiWriter(writers...), prefix.Info, logFlag)
}

func SetWarnWrite(writers ...io.Writer) {
	LogWarn = log.New(io.MultiWriter(writers...), prefix.Warn, logFlag)
}

func SetErrorWrite(writers ...io.Writer) {
	LogError = log.New(io.MultiWriter(writers...), prefix.Error, logFlag)
}

func GetTraceCtx(ctx context.Context, traceId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxTraceIdKey, traceId)
}
