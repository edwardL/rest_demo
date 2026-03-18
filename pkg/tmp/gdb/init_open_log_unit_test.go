package gdb

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestInitOptionsApply(t *testing.T) {
	var c = &Conf{DriveMap: make(map[string]NewWrapperFunc)}

	WithLogLevel(DebugLogLevel)(c)
	if c.LogLevel != DebugLogLevel {
		t.Fatalf("WithLogLevel 未生效")
	}

	WithWriteHhDbLog(true)(c)
	WithWriteLog(false)(c)
	WithWriteErrSql(false)(c)
	WithWriteCompSql(false)(c)
	if !c.WriteHhDbLog || c.WriteLog || c.WriteErrSql || c.WriteCompSql {
		t.Fatalf("日志开关 Option 未生效: %+v", c)
	}

	var traceFn = func(ctx context.Context) string { return "trace" }
	WithTraceIdFunc(traceFn)(c)
	if c.TraceIdFunc == nil || c.TraceIdFunc(context.Background()) != "trace" {
		t.Fatalf("WithTraceIdFunc 未生效")
	}

	WithAppendDbKeywords([]string{"kw_test_1"})(c)
	if len(c.DbKeywords) != 1 || c.DbKeywords[0] != "kw_test_1" {
		t.Fatalf("WithAppendDbKeywords 未生效")
	}

	WithAppendDbFieldChar([]byte{'$'})(c)
	if len(c.DbFieldChar) != 1 || c.DbFieldChar[0] != '$' {
		t.Fatalf("WithAppendDbFieldChar 未生效")
	}

	WithAppendZeroValIgnoreField([]string{"x_zero"})(c)
	if len(c.ZeroValIgnoreField) != 1 || c.ZeroValIgnoreField[0] != "x_zero" {
		t.Fatalf("WithAppendZeroValIgnoreField 未生效")
	}

	var notFound = NewDbErr("NF", "not found")
	WithEmptyError(notFound)(c)
	if c.EmptyError == nil || !c.EmptyError.isNotFound {
		t.Fatalf("WithEmptyError 未生效")
	}

	WithDbConvInitPtr(true)(c)
	if !c.DbConvInitPtr {
		t.Fatalf("WithDbConvInitPtr 未生效")
	}

	WithTimeLocation(time.UTC)(c)
	if c.TimeLocation != time.UTC {
		t.Fatalf("WithTimeLocation 未生效")
	}

	WithDriveType("mysql_x")(c)
	if c.DriveType != "mysql_x" {
		t.Fatalf("WithDriveType 未生效")
	}

	var newFun NewWrapperFunc = func(s *Sql) SqlWrapperFace { return s }
	WithDriveMap("x", newFun)(c)
	if c.DriveMap["x"] == nil {
		t.Fatalf("WithDriveMap 未生效")
	}
}

func TestInitApplyToGlobal(t *testing.T) {
	Init(
		WithAppendDbKeywords([]string{"kw_test_global_1"}),
		WithAppendDbFieldChar([]byte{'@'}),
		WithAppendZeroValIgnoreField([]string{"zero_global_1"}),
	)

	var ok bool
	_, ok = mysqlKeywords["kw_test_global_1"]
	if !ok {
		t.Fatalf("Init 未追加 mysqlKeywords")
	}
	if !InArr(byte('@'), mysqlFieldChat) {
		t.Fatalf("Init 未追加 mysqlFieldChat")
	}
	if !InArr("zero_global_1", zeroValIgnoreField) {
		t.Fatalf("Init 未追加 zeroValIgnoreField")
	}
}

func TestOpenGenWhereAndWrapper(t *testing.T) {
	var db = &Db{gs: NewSql("users")}
	db.Where("id = ?", 1)
	var where string
	var args []any
	where, args = db.GenWhere()
	if where == "" || len(args) != 1 {
		t.Fatalf("Db GenWhere 异常: where=%s args=%v", where, args)
	}

	var qw = db.QueryWrapper()
	if qw == nil || qw.Db != db {
		t.Fatalf("Db QueryWrapper 异常")
	}

	var sqlObj = NewSql("users")
	sqlObj.Where("name = ?", "tom")
	where, args = sqlObj.GenWhere()
	if where == "" || len(args) != 1 {
		t.Fatalf("Sql GenWhere 异常: where=%s args=%v", where, args)
	}

	var qw2 = sqlObj.QueryWrapper()
	if qw2 == nil || qw2.gs != sqlObj {
		t.Fatalf("Sql QueryWrapper 异常")
	}
}

func TestLogCallDepthCtx(t *testing.T) {
	var ctx = setLogCallDepthCtx(nil, 5)
	if getLogCallDepthCtx(ctx) != 5 {
		t.Fatalf("set/getLogCallDepthCtx 异常")
	}

	ctx = appendLogCallDepthCtx(ctx, 2)
	if getLogCallDepthCtx(ctx) != 7 {
		t.Fatalf("appendLogCallDepthCtx 异常")
	}

	if getLogCallDepthCtx(nil) != 3 {
		t.Fatalf("nil ctx 默认 callDepth 异常")
	}
}

func TestGdbLogAndWriter(t *testing.T) {
	var debugBuf = &bytes.Buffer{}
	var infoBuf = &bytes.Buffer{}
	var warnBuf = &bytes.Buffer{}
	var errBuf = &bytes.Buffer{}
	SetDebugLogWriter(debugBuf)
	SetInfoLogWriter(infoBuf)
	SetWarnLogWriter(warnBuf)
	SetErrorLogWriter(errBuf)

	var key = struct{}{}
	var lg = &gdbLog{level: DebugLogLevel}
	lg.SetTraceFunc(func(ctx context.Context) string {
		var v = ctx.Value(key)
		if v == nil {
			return ""
		}
		return v.(string)
	})

	var ctx = context.WithValue(context.Background(), key, "T1")
	ctx = setLogCallDepthCtx(ctx, 1)
	lg.CtxDebugf(ctx, " hello %d", 1)
	lg.CtxInfo(ctx, "info")
	lg.CtxWarn(ctx, "warn")
	lg.CtxError(ctx, "err")

	if !strings.Contains(debugBuf.String(), "T1 hello 1") {
		t.Fatalf("CtxDebugf 输出异常: %s", debugBuf.String())
	}
	if !strings.Contains(infoBuf.String(), "T1 info") {
		t.Fatalf("CtxInfo 输出异常: %s", infoBuf.String())
	}
	if !strings.Contains(warnBuf.String(), "T1 warn") {
		t.Fatalf("CtxWarn 输出异常: %s", warnBuf.String())
	}
	if !strings.Contains(errBuf.String(), "T1 err") {
		t.Fatalf("CtxError 输出异常: %s", errBuf.String())
	}

	var clone = lg.Clone()
	if clone == nil {
		t.Fatalf("Clone 为空")
	}

	var w = writer{Level: "info"}
	var n int
	var err error
	n, err = w.Write([]byte("0123456789012345678901234567line\n"))
	if err != nil || n != 0 {
		t.Fatalf("writer.Write 返回异常: n=%d err=%v", n, err)
	}
}
