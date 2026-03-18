package log

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func resetLoggerForTest() (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	var debugBuf = &bytes.Buffer{}
	var infoBuf = &bytes.Buffer{}
	var warnBuf = &bytes.Buffer{}
	var errBuf = &bytes.Buffer{}

	SetLogFlag(0)
	SetLogPrefix(Prefix{
		Info:  "[INFO] ",
		Debug: "[DEBUG] ",
		Warn:  "[WARN] ",
		Error: "[ERROR] ",
	})
	SetDebugWrite(debugBuf)
	SetInfoWrite(infoBuf)
	SetWarnWrite(warnBuf)
	SetErrorWrite(errBuf)
	SetCtxFunc(func(ctx context.Context) string {
		var v = ctx.Value("trace_id")
		if v == nil {
			return ""
		}
		var traceId, _ = v.(string)
		return traceId
	})

	return debugBuf, infoBuf, warnBuf, errBuf
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel("debug")
	if logLevel != LevelDebug {
		t.Fatalf("SetLogLevel debug failed")
	}

	SetLogLevel("info")
	if logLevel != LevelInfo {
		t.Fatalf("SetLogLevel info failed")
	}

	SetLogLevel("warn")
	if logLevel != LevelWarn {
		t.Fatalf("SetLogLevel warn failed")
	}

	SetLogLevel("error")
	if logLevel != LevelError {
		t.Fatalf("SetLogLevel error failed")
	}

	SetLogLevel("close")
	if logLevel != LevelClose {
		t.Fatalf("SetLogLevel close failed")
	}
}

func TestGetTraceCtx(t *testing.T) {
	var ctx = GetTraceCtx(nil, "trace-001")
	if ctx == nil {
		t.Fatalf("GetTraceCtx should not return nil")
	}
	var traceVal = ctx.Value(ctxTraceIdKey)
	if traceVal == nil {
		t.Fatalf("GetTraceCtx should set trace value")
	}
	if traceVal.(string) != "trace-001" {
		t.Fatalf("GetTraceCtx trace mismatch")
	}
}

func TestOutputWithLevelControl(t *testing.T) {
	var debugBuf, infoBuf, warnBuf, errBuf = resetLoggerForTest()
	_ = debugBuf

	SetLevel(LevelWarn)
	Info("info-message")
	Warn("warn-message")
	Error("error-message")
	Debug("debug-message")

	if infoBuf.Len() != 0 {
		t.Fatalf("info should be filtered at warn level")
	}
	if !strings.Contains(warnBuf.String(), "warn-message") {
		t.Fatalf("warn output mismatch: %s", warnBuf.String())
	}
	if !strings.Contains(errBuf.String(), "error-message") {
		t.Fatalf("error output mismatch: %s", errBuf.String())
	}
	if debugBuf.Len() != 0 {
		t.Fatalf("debug should be filtered at warn level")
	}
}

func TestOutputAndFormatFunctions(t *testing.T) {
	var _, infoBuf, warnBuf, errBuf = resetLoggerForTest()
	SetLevel(LevelDebug)

	Infof("id=%d", 10)
	Warnf("name=%s", "tom")
	Errorf("ok=%t", true)

	if !strings.Contains(infoBuf.String(), "id=10") {
		t.Fatalf("Infof output mismatch: %s", infoBuf.String())
	}
	if !strings.Contains(warnBuf.String(), "name=tom") {
		t.Fatalf("Warnf output mismatch: %s", warnBuf.String())
	}
	if !strings.Contains(errBuf.String(), "ok=true") {
		t.Fatalf("Errorf output mismatch: %s", errBuf.String())
	}
}

func TestCtxOutputFunctions(t *testing.T) {
	var _, infoBuf, warnBuf, _ = resetLoggerForTest()
	SetLevel(LevelDebug)

	var ctx = context.WithValue(context.Background(), "trace_id", "trace-ctx")
	CtxInfo(ctx, "hello")
	CtxWarnf(ctx, "age=%d", 18)

	if !strings.Contains(infoBuf.String(), "trace-ctx") || !strings.Contains(infoBuf.String(), "hello") {
		t.Fatalf("CtxInfo output mismatch: %s", infoBuf.String())
	}
	if !strings.Contains(warnBuf.String(), "trace-ctx age=18") {
		t.Fatalf("CtxWarnf output mismatch: %s", warnBuf.String())
	}
}

func TestBuildCtxFormat(t *testing.T) {
	resetLoggerForTest()
	var ctx = context.WithValue(context.Background(), "trace_id", "T-1")
	var got = buildCtxFormat(ctx, "msg=%s")
	if got != "T-1 msg=%s" {
		t.Fatalf("buildCtxFormat mismatch: %s", got)
	}

	var got2 = buildCtxFormat(context.Background(), "msg")
	if got2 != "msg" {
		t.Fatalf("buildCtxFormat without trace mismatch: %s", got2)
	}
}
