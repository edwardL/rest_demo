package log

import (
	"log"
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
	logFlag       int = log.Ldate | log.Ltime | log.Lshortfile
	ctxTraceIdKey     = struct{}{}

	logLevel Level  = LevelDebug
	prefix   Prefix = Prefix{
		Info:  colorBlue + "[INFO] " + colorReset + " ",
		Debug: colorPurple + "[DEBUG]" + colorReset + " ",
		Warn:  colorYellow + "[WARN] " + colorReset + " ",
		Error: colorRed + "[ERROR]" + colorReset + " ",
	}
)

type Prefix struct {
	Info  string
	Debug string
	Warn  string
	Error string
}

type Writer struct {
	Level string
}

// Write 实现 io.Writer 接口
func (w Writer) Write(b []byte) (n int, err error) {
	return len(b), nil
}

// Init 初始化 默认日志输出其和
func Init() {

}

func init() {
	Init()
}
