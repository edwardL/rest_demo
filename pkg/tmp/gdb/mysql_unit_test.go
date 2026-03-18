package gdb

import (
	"reflect"
	"strings"
	"testing"
)

func TestSqlGetSearchWhereCases(t *testing.T) {
	var s = NewSql("users")
	var op string
	var vals []any
	var err error

	op, vals, err = s.getSearchWhere("equal", 1)
	if err != nil || op != "= ?" || len(vals) != 1 || vals[0].(int) != 1 {
		t.Fatalf("equal 解析失败: op=%s vals=%v err=%v", op, vals, err)
	}

	op, vals, err = s.getSearchWhere("in", "1,2,3")
	if err != nil || op != "IN (?)" || len(vals) != 1 {
		t.Fatalf("in(string) 解析失败: op=%s vals=%v err=%v", op, vals, err)
	}

	op, vals, err = s.getSearchWhere("between", []any{1, 2})
	if err != nil || op != "BETWEEN ? AND ?" || len(vals) != 2 {
		t.Fatalf("between([]any) 解析失败: op=%s vals=%v err=%v", op, vals, err)
	}

	op, vals, err = s.getSearchWhere("between", "2024-01-01,2024-12-31")
	if err != nil || op != "BETWEEN ? AND ?" || len(vals) != 2 {
		t.Fatalf("between(string) 解析失败: op=%s vals=%v err=%v", op, vals, err)
	}

	op, vals, err = s.getSearchWhere("like", "tom")
	if err != nil || op != "LIKE ?" || len(vals) != 1 || vals[0].(string) != "%tom%" {
		t.Fatalf("like 解析失败: op=%s vals=%v err=%v", op, vals, err)
	}

	_, _, err = s.getSearchWhere("unknown_cmp", 1)
	if err == nil {
		t.Fatalf("不支持的比较符应返回错误")
	}
}

func TestSqlGenQueryCases(t *testing.T) {
	var s = NewSql("users")
	var query string
	var args []any
	var err error

	query, args, err = s.genQuery("id in (?)", []any{[]int{1, 2, 3}})
	if err != nil || len(args) != 3 {
		t.Fatalf("genQuery []int 失败: query=%s args=%v err=%v", query, args, err)
	}
	if !strings.Contains(query, "?") {
		t.Fatalf("genQuery []int 占位符异常: %s", query)
	}

	query, args, err = s.genQuery("(a,b) in (?)", []any{[][]any{{1, 2}, {3, 4}}})
	if err != nil || len(args) != 4 {
		t.Fatalf("genQuery [][]any 失败: query=%s args=%v err=%v", query, args, err)
	}

	query, args, err = s.genQuery("id in (?)", []any{[3]int{7, 8, 9}})
	if err != nil || len(args) != 3 {
		t.Fatalf("genQuery array 失败: query=%s args=%v err=%v", query, args, err)
	}
	if !reflect.DeepEqual(args, []any{7, 8, 9}) {
		t.Fatalf("genQuery array 参数错误: %v", args)
	}

	query, args, err = s.genQuery("ipv6 = ?", []any{RawAny([]byte{1, 2})})
	if err != nil || len(args) != 1 {
		t.Fatalf("genQuery RawArg 失败: query=%s args=%v err=%v", query, args, err)
	}

	_, _, err = s.genQuery("(a,b) in (?)", []any{[][]struct{}{}})
	if err == nil {
		t.Fatalf("空二维未知类型应返回错误")
	}
}

func TestSqlPpNumAndGenWhere(t *testing.T) {
	var s = NewSql("users")
	if s.ppNum("a=? and b=? and c=?") != 3 {
		t.Fatalf("ppNum 统计错误")
	}

	s.Where("id = ?", 1).WhereOr("name = ?", "tom")
	var where string
	var args []any
	where, args = s.GenWhere()
	if where == "" || len(args) != 2 {
		t.Fatalf("GenWhere 结果异常: where=%s args=%v", where, args)
	}
}
