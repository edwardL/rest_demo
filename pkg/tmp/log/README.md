# utils/log（模型友好版）

`utils/log` 是轻量日志包，提供级别控制、格式化输出、上下文 trace 注入。

## 1) 快速使用

```go
package main

import (
	"context"
	hlog "nwgit.gzhhit.com/BD/hhitcommcode.git/utils/log"
)

func main() {
	hlog.Init()
	hlog.SetLogLevel("info")
	hlog.SetCtxFunc(func(ctx context.Context) string {
		var traceId, _ = ctx.Value("trace_id").(string)
		return traceId
	})

	var ctx = context.WithValue(context.Background(), "trace_id", "trace-001")
	hlog.Info("service started")
	hlog.CtxInfof(ctx, "query user id=%d", 100)
}
```

## 2) 全部导出 API（完整）

### 常量

- `LevelDebug`
- `LevelInfo`
- `LevelWarn`
- `LevelError`
- `LevelClose`

### 变量

- `LogDebug *log.Logger`
- `LogInfo *log.Logger`
- `LogWarn *log.Logger`
- `LogError *log.Logger`

### 类型

- `type Level int`
- `type CtxFunc func(ctx context.Context) string`
- `type Prefix struct { Info, Debug, Warn, Error string }`
- `type Writer struct { Level string }`

### 初始化与配置

- `Init()`
- `SetLogFlag(flag int)`
- `SetLogPrefix(p Prefix)`
- `SetLevel(lv Level)`
- `SetLogLevel(lv string)`
- `SetCtxFunc(f CtxFunc)`
- `GetTraceCtx(ctx context.Context, traceId string) context.Context`

### Writer 配置

- `SetDebugWrite(writers ...io.Writer)`
- `SetInfoWrite(writers ...io.Writer)`
- `SetWarnWrite(writers ...io.Writer)`
- `SetErrorWrite(writers ...io.Writer)`

### 低层输出（可控调用深度）

- `OutputInfo(callDepth int, v ...any)`
- `OutputWarn(callDepth int, v ...any)`
- `OutputError(callDepth int, v ...any)`
- `OutputDebug(callDepth int, v ...any)`
- `OutputInfof(callDepth int, format string, v ...any)`
- `OutputWarnf(callDepth int, format string, v ...any)`
- `OutputErrorf(callDepth int, format string, v ...any)`
- `OutputDebugf(callDepth int, format string, v ...any)`

### 常用输出

- `Info(v ...any)`
- `Warn(v ...any)`
- `Error(v ...any)`
- `Debug(v ...any)`
- `Infof(format string, v ...any)`
- `Warnf(format string, v ...any)`
- `Errorf(format string, v ...any)`
- `Debugf(format string, v ...any)`

### 上下文输出

- `CtxInfo(ctx context.Context, v ...any)`
- `CtxWarn(ctx context.Context, v ...any)`
- `CtxError(ctx context.Context, v ...any)`
- `CtxDebug(ctx context.Context, v ...any)`
- `CtxInfof(ctx context.Context, format string, v ...any)`
- `CtxWarnf(ctx context.Context, format string, v ...any)`
- `CtxErrorf(ctx context.Context, format string, v ...any)`
- `CtxDebugf(ctx context.Context, format string, v ...any)`

## 3) 建议实践

- 进程启动时统一执行 `Init + SetLogLevel + SetCtxFunc`
- 业务日志用 `Info/Warn/Error`；框架封装可用 `Output*` 控制调用深度
- 单测将 writer 指向 `bytes.Buffer` 断言输出

## 4) 测试命令

```bash
go test ./utils/log -v
go test ./utils/log -run '^TestSetLogLevel$' -v
```
