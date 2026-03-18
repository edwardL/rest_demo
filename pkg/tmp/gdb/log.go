package gdb

import (
	"context"
	"fmt"
	"io"
	"log"
	hhlog "nwgit.gzhhit.com/BD/hhitlog.git"
	"os"
	"strings"
)

type LogLevel int

const (
	DebugLogLevel LogLevel = iota
	InfoLogLevel
	WarnLogLevel
	ErrorLogLevel
)

// 初始化内部日志
func init() {
	logDebug = log.New(io.MultiWriter(os.Stdout, writer{Level: "debug"}), "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	logInfo = log.New(io.MultiWriter(os.Stdout, writer{Level: "info"}), "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile)
	logWarn = log.New(io.MultiWriter(os.Stdout, writer{Level: "warn"}), "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(io.MultiWriter(os.Stderr, writer{Level: "error"}), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

type logCallDepth struct {
}

var logCallDepthKey = logCallDepth{}

func setLogCallDepthCtx(ctx context.Context, callDepth int) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, logCallDepthKey, callDepth)
}

func appendLogCallDepthCtx(ctx context.Context, callDepth int) context.Context {
	return setLogCallDepthCtx(ctx, getLogCallDepthCtx(ctx)+callDepth)
}

func getLogCallDepthCtx(ctx context.Context) int {
	if ctx == nil {
		return 3
	}
	if v := ctx.Value(logCallDepthKey); v != nil {
		if callDepth, ok := v.(int); ok {
			return callDepth
		}
	}
	return 3
}

// GdbLog 抽离依赖，支持注入
type GdbLog interface {
	CtxInfo(ctx context.Context, v ...any)
	CtxInfof(ctx context.Context, format string, v ...any)
	CtxDebug(ctx context.Context, v ...any)
	CtxDebugf(ctx context.Context, format string, v ...any)
	CtxWarn(ctx context.Context, v ...any)
	CtxWarnf(ctx context.Context, format string, v ...any)
	CtxError(ctx context.Context, v ...any)
	CtxErrorf(ctx context.Context, format string, v ...any)
	SetLevel(level LogLevel)
	SetTraceFunc(func(ctx context.Context) string)
	Clone() GdbLog
}

var defLog GdbLog = &gdbLog{
	level: InfoLogLevel,
}

// SetDefLog 设置数据库日志
func SetDefLog(l GdbLog) {
	defLog = l
}

// SetLevel 设置数据库日志级别
func SetLevel(level LogLevel) {
	defLog.SetLevel(level)
}

// SetTraceFunc 设置跟踪id提前函数
func SetTraceFunc(f func(ctx context.Context) string) {
	defLog.SetTraceFunc(f)
}

// SetDebugLogWriter 设置日志输出
func SetDebugLogWriter(w ...io.Writer) {
	logDebug = log.New(io.MultiWriter(w...), "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetInfoLogWriter 设置日志输出
func SetInfoLogWriter(w ...io.Writer) {
	logInfo = log.New(io.MultiWriter(w...), "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetWarnLogWriter 设置日志输出
func SetWarnLogWriter(w ...io.Writer) {
	logWarn = log.New(io.MultiWriter(w...), "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetErrorLogWriter 设置日志输出
func SetErrorLogWriter(w ...io.Writer) {
	logError = log.New(io.MultiWriter(w...), "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

type gdbLog struct {
	level     LogLevel
	traceFunc func(ctx context.Context) string
}

func (l *gdbLog) CtxInfo(ctx context.Context, v ...any) {
	if l.level > InfoLogLevel {
		return
	}
	_ = logInfo.Output(getLogCallDepthCtx(ctx), string(fmt.Appendln([]byte{}, append([]any{l.genCtx(ctx)}, v...)...)))
}
func (l *gdbLog) CtxInfof(ctx context.Context, format string, v ...any) {
	if l.level > InfoLogLevel {
		return
	}
	_ = logInfo.Output(getLogCallDepthCtx(ctx), string(fmt.Appendf([]byte{}, l.genCtx(ctx)+format, v...)))
}
func (l *gdbLog) CtxDebug(ctx context.Context, v ...any) {
	if l.level > DebugLogLevel {
		return
	}
	logDebug.Output(getLogCallDepthCtx(ctx), string(fmt.Appendln([]byte{}, append([]any{l.genCtx(ctx)}, v...)...)))
}
func (l *gdbLog) CtxDebugf(ctx context.Context, format string, v ...any) {
	if l.level > DebugLogLevel {
		return
	}
	logDebug.Output(getLogCallDepthCtx(ctx), string(fmt.Appendf([]byte{}, l.genCtx(ctx)+format, v...)))
}
func (l *gdbLog) CtxWarn(ctx context.Context, v ...any) {
	if l.level > WarnLogLevel {
		return
	}
	logWarn.Output(getLogCallDepthCtx(ctx), string(fmt.Appendln([]byte{}, append([]any{l.genCtx(ctx)}, v...)...)))
}
func (l *gdbLog) CtxWarnf(ctx context.Context, format string, v ...any) {
	if l.level > WarnLogLevel {
		return
	}
	logWarn.Output(getLogCallDepthCtx(ctx), string(fmt.Appendf([]byte{}, l.genCtx(ctx)+format, v...)))
}
func (l *gdbLog) CtxError(ctx context.Context, v ...any) {
	if l.level > ErrorLogLevel {
		return
	}
	logError.Output(getLogCallDepthCtx(ctx), string(fmt.Appendln([]byte{}, append([]any{l.genCtx(ctx)}, v...)...)))
}
func (l *gdbLog) CtxErrorf(ctx context.Context, format string, v ...any) {
	if l.level > ErrorLogLevel {
		return
	}
	logError.Output(getLogCallDepthCtx(ctx), string(fmt.Appendf([]byte{}, l.genCtx(ctx)+format, v...)))
}

func (l *gdbLog) Clone() GdbLog {
	var newLog = new(gdbLog)
	*newLog = *l
	return newLog
}

func (l *gdbLog) genCtx(ctx context.Context) string {
	var str = ""
	if ctx == nil {
		return str
	}
	if l.traceFunc != nil {
		str = l.traceFunc(ctx)
	}
	return str
}

func (l *gdbLog) SetLevel(level LogLevel) {
	l.level = level
}

func (l *gdbLog) SetTraceFunc(f func(ctx context.Context) string) {
	l.traceFunc = f
}

var (
	logDebug *log.Logger
	logInfo  *log.Logger
	logWarn  *log.Logger
	logError *log.Logger
)

type writer struct {
	Level string
}

func (w writer) Write(b []byte) (n int, err error) {
	hhlog.HLogOutPut(&hhlog.SLogStruct{
		LogLevel:  w.Level,
		LogOutPut: "file",
		LogInfo:   strings.Trim(string(b), "\n")[27:],
	})
	return 0, nil
}
