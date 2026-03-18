package log

import (
	"context"
	"log"
	hhlog "nwgit.gzhhit.com/BD/hhitlog.git"
	"os"
	"strings"
)

var (
	LogDebug *log.Logger
	LogInfo  *log.Logger
	LogWarn  *log.Logger
	LogError *log.Logger
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelClose

	colorReset  = "\033[0m"
	colorRed    = "\033[31m" // Error 红色
	colorYellow = "\033[33m" // Warning 黄色
	colorBlue   = "\033[34m" // Info 蓝色
	colorPurple = "\033[35m" // Debug 紫色
)

var (
	logFlag       int     = log.Ldate | log.Ltime | log.Lshortfile
	ctxTraceIdKey         = struct{}{}
	ctxFunc       CtxFunc = func(ctx context.Context) string {
		var v = ctx.Value(ctxTraceIdKey)
		if v != nil {
			if sv, ok := v.(string); ok {
				return sv
			}
		}
		return ""
	}
	logLevel Level  = LevelDebug
	prefix   Prefix = Prefix{
		Info:  colorBlue + "[INFO] " + colorReset + " ",
		Debug: colorPurple + "[DEBUG]" + colorReset + " ",
		Warn:  colorYellow + "[WARN] " + colorReset + " ",
		Error: colorRed + "[ERROR]" + colorReset + " ",
	}
)

type CtxFunc func(ctx context.Context) string
type Prefix struct {
	Info  string
	Debug string
	Warn  string
	Error string
}
type Writer struct {
	Level string
}

// Write 实现 io.Writer 接口并将日志转发到 hhitlog。
func (w Writer) Write(b []byte) (n int, err error) {
	hhlog.HLogOutPut(&hhlog.SLogStruct{
		LogLevel:  w.Level,
		LogOutPut: "file",
		LogInfo:   strings.Trim(string(b), "\n")[27:],
	})
	return len(b), nil
}

// Init 初始化默认日志输出器和级别输出通道。
func Init() {
	SetDebugWrite(os.Stdout, Writer{Level: "debug"})
	SetInfoWrite(os.Stdout, Writer{Level: "info"})
	SetWarnWrite(os.Stdout, Writer{Level: "warn"})
	SetErrorWrite(os.Stderr, Writer{Level: "error"})
}

func init() {
	Init()
}
