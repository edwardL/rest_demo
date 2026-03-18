package retval

import (
	hhtools "nwgit.gzhhit.com/BD/hhitcomm.git"
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
	"testing"
)

func TestSuccessSeries(t *testing.T) {
	var tp uint8
	var resp any

	tp, resp = Success1()
	if tp != 1 {
		t.Fatalf("Success1 类型错误: %d", tp)
	}
	var s1 = resp.(ftypes.ResultData01)
	if !s1.Result || s1.Code != hhtools.CodeGetDataSuccess || s1.Msg != hhtools.MsgGetDataSuccess {
		t.Fatalf("Success1 基础字段错误: %#v", s1)
	}

	tp, resp = Success2("abc")
	if tp != 2 {
		t.Fatalf("Success2 类型错误: %d", tp)
	}
	var s2 = resp.(ftypes.ResultData02)
	if s2.Data != "abc" {
		t.Fatalf("Success2 Data 错误: %#v", s2)
	}

	tp, resp = Success3(map[string]string{"k": "v"})
	if tp != 3 {
		t.Fatalf("Success3 类型错误: %d", tp)
	}
	var s3 = resp.(ftypes.ResultData03)
	if s3.Data["k"] != "v" {
		t.Fatalf("Success3 Data 错误: %#v", s3)
	}

	tp, resp = Success4([]any{1, "x"})
	if tp != 4 {
		t.Fatalf("Success4 类型错误: %d", tp)
	}
	var s4 = resp.(ftypes.ResultData04)
	if len(s4.Data) != 2 {
		t.Fatalf("Success4 Data 错误: %#v", s4)
	}

	tp, resp = Success5(map[string]any{"id": 1})
	if tp != 5 {
		t.Fatalf("Success5 类型错误: %d", tp)
	}
	var s5 = resp.(ftypes.ResultData05)
	var m = s5.Data.(map[string]any)
	if m["id"].(int) != 1 {
		t.Fatalf("Success5 Data 错误: %#v", s5)
	}
}

func TestErrSeries(t *testing.T) {
	var tp uint8
	var resp any

	tp, resp = Err1("E1", "m1")
	if tp != 1 {
		t.Fatalf("Err1 类型错误: %d", tp)
	}
	var e1 = resp.(ftypes.ResultData01)
	if e1.Result || e1.Code != "E1" || e1.Msg != "m1" {
		t.Fatalf("Err1 基础字段错误: %#v", e1)
	}

	tp, resp = Err2("E2", "m2", "d2")
	if tp != 2 {
		t.Fatalf("Err2 类型错误: %d", tp)
	}
	var e2 = resp.(ftypes.ResultData02)
	if e2.Result || e2.Data != "d2" {
		t.Fatalf("Err2 字段错误: %#v", e2)
	}

	tp, resp = Err3("E3", "m3", map[string]string{"k": "v"})
	if tp != 3 {
		t.Fatalf("Err3 类型错误: %d", tp)
	}
	var e3 = resp.(ftypes.ResultData03)
	if e3.Result || e3.Data["k"] != "v" {
		t.Fatalf("Err3 字段错误: %#v", e3)
	}

	tp, resp = Err4("E4", "m4", []interface{}{1, "x"})
	if tp != 4 {
		t.Fatalf("Err4 类型错误: %d", tp)
	}
	var e4 = resp.(ftypes.ResultData04)
	if e4.Result || len(e4.Data) != 2 {
		t.Fatalf("Err4 字段错误: %#v", e4)
	}

	tp, resp = Err5("E5", "m5", map[string]any{"id": 1})
	if tp != 5 {
		t.Fatalf("Err5 类型错误: %d", tp)
	}
	var e5 = resp.(ftypes.ResultData05)
	var m = e5.Data.(map[string]any)
	if e5.Result || m["id"].(int) != 1 {
		t.Fatalf("Err5 字段错误: %#v", e5)
	}
}
