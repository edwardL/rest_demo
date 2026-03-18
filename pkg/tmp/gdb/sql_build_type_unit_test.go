package gdb

import (
	"strings"
	"testing"
)

func TestRawAndRawAny(t *testing.T) {
	var r = Raw("age = age + 1")
	if string(r) != "age = age + 1" {
		t.Fatalf("Raw 结果错误")
	}

	var ra = RawAny([]byte{1, 2, 3})
	if ra.val == nil {
		t.Fatalf("RawAny 结果错误")
	}
}

func TestResultCompSql(t *testing.T) {
	var b = strings.Builder{}
	b.WriteString("SELECT * FROM t WHERE a=? AND b=? AND c=? AND d=? AND e=?")
	var r = Result{
		Sql:  &b,
		Args: []any{1, true, "a'b", []byte("xx"), nil},
	}
	var sql = r.CompSql()
	if !strings.Contains(sql, "a=1") {
		t.Fatalf("数字参数替换失败: %s", sql)
	}
	if !strings.Contains(sql, "b=1") {
		t.Fatalf("布尔参数替换失败: %s", sql)
	}
	if !strings.Contains(sql, "c='a\\'b'") {
		t.Fatalf("字符串转义失败: %s", sql)
	}
	if !strings.Contains(sql, "d='xx'") {
		t.Fatalf("[]byte 参数替换失败: %s", sql)
	}
	if !strings.Contains(sql, "e=NULL") {
		t.Fatalf("nil 参数替换失败: %s", sql)
	}
}
