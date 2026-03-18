package log

import (
	"context"
	"fmt"
	"strings"
)

// OutputInfo 按信息级别输出日志。
func OutputInfo(callDepth int, v ...any) {
	if logLevel > LevelInfo {
		return
	}
	callDepth += 2
	_ = LogInfo.Output(callDepth, string(fmt.Appendln([]byte{}, v...)))
}

// OutputError 按错误级别输出日志。
func OutputError(callDepth int, v ...any) {
	if logLevel > LevelError {
		return
	}
	callDepth += 2
	_ = LogError.Output(callDepth, string(fmt.Appendln([]byte{}, v...)))
}

// OutputWarn 按告警级别输出日志。
func OutputWarn(callDepth int, v ...any) {
	if logLevel > LevelWarn {
		return
	}
	callDepth += 2
	_ = LogWarn.Output(callDepth, string(fmt.Appendln([]byte{}, v...)))
}

// OutputDebug 按调试级别输出日志。
func OutputDebug(callDepth int, v ...any) {
	if logLevel > LevelDebug {
		return
	}
	callDepth += 2
	_ = LogDebug.Output(callDepth, string(fmt.Appendln([]byte{}, v...)))
}

// OutputInfof 按信息级别输出格式化日志。
func OutputInfof(callDepth int, format string, v ...any) {
	if logLevel > LevelInfo {
		return
	}
	callDepth += 2
	_ = LogInfo.Output(callDepth, string(fmt.Appendf([]byte{}, format, v...)))
}

// OutputErrorf 按错误级别输出格式化日志。
func OutputErrorf(callDepth int, format string, v ...any) {
	if logLevel > LevelError {
		return
	}
	callDepth += 2
	_ = LogError.Output(callDepth, string(fmt.Appendf([]byte{}, format, v...)))
}

// OutputWarnf 按告警级别输出格式化日志。
func OutputWarnf(callDepth int, format string, v ...any) {
	if logLevel > LevelWarn {
		return
	}
	callDepth += 2
	_ = LogWarn.Output(callDepth, string(fmt.Appendf([]byte{}, format, v...)))
}

// OutputDebugf 按调试级别输出格式化日志。
func OutputDebugf(callDepth int, format string, v ...any) {
	if logLevel > LevelDebug {
		return
	}
	callDepth += 2
	_ = LogDebug.Output(callDepth, string(fmt.Appendf([]byte{}, format, v...)))
}

// Info 输出信息级别日志。
func Info(v ...any) {
	OutputInfo(1, v...)
}

// Error 输出错误级别日志。
func Error(v ...any) {
	OutputError(1, v...)
}

// Warn 输出告警级别日志。
func Warn(v ...any) {
	OutputWarn(1, v...)
}

// Debug 输出调试级别日志。
func Debug(v ...any) {
	OutputDebug(1, v...)
}

// CtxInfo 输出携带上下文标识的信息日志。
func CtxInfo(ctx context.Context, v ...any) {
	if ctxFunc != nil {
		v = append([]any{ctxFunc(ctx)}, v...)
	}
	OutputInfo(1, v...)
}

// CtxError 输出携带上下文标识的错误日志。
func CtxError(ctx context.Context, v ...any) {
	if ctxFunc != nil {
		v = append([]any{ctxFunc(ctx)}, v...)
	}
	OutputError(1, v...)
}

// CtxWarn 输出携带上下文标识的告警日志。
func CtxWarn(ctx context.Context, v ...any) {
	if ctxFunc != nil {
		v = append([]any{ctxFunc(ctx)}, v...)
	}
	OutputWarn(1, v...)
}

// CtxDebug 输出携带上下文标识的调试日志。
func CtxDebug(ctx context.Context, v ...any) {
	if ctxFunc != nil {
		v = append([]any{ctxFunc(ctx)}, v...)
	}
	OutputDebug(1, v...)
}

// Infof 输出信息级别的格式化日志。
func Infof(format string, v ...any) {
	OutputInfof(1, format, v...)
}

// Errorf 输出错误级别的格式化日志。
func Errorf(format string, v ...any) {
	OutputErrorf(1, format, v...)
}

// Warnf 输出告警级别的格式化日志。
func Warnf(format string, v ...any) {
	OutputWarnf(1, format, v...)
}

// Debugf 输出调试级别的格式化日志。
func Debugf(format string, v ...any) {
	OutputDebugf(1, format, v...)
}

// 构建带上下文ID的格式字符串
func buildCtxFormat(ctx context.Context, format string) string {
	if ctxFunc != nil {
		traceId := ctxFunc(ctx)
		if traceId != "" {
			var sb strings.Builder
			sb.Grow(len(traceId) + 1 + len(format))
			sb.WriteString(traceId)
			sb.WriteString(" ")
			sb.WriteString(format)
			return sb.String()
		}
	}
	return format
}

// CtxInfof 输出携带上下文标识的信息格式化日志。
func CtxInfof(ctx context.Context, format string, v ...any) {
	OutputInfof(1, buildCtxFormat(ctx, format), v...)
}

// CtxErrorf 输出携带上下文标识的错误格式化日志。
func CtxErrorf(ctx context.Context, format string, v ...any) {
	OutputErrorf(1, buildCtxFormat(ctx, format), v...)
}

// CtxWarnf 输出携带上下文标识的告警格式化日志。
func CtxWarnf(ctx context.Context, format string, v ...any) {
	OutputWarnf(1, buildCtxFormat(ctx, format), v...)
}

// CtxDebugf 输出携带上下文标识的调试格式化日志。
func CtxDebugf(ctx context.Context, format string, v ...any) {
	OutputDebugf(1, buildCtxFormat(ctx, format), v...)
}
