package gdb

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestErrorFunctions(t *testing.T) {
	var e = NewDbErr("E1", "m1")
	if e.Error() != "E1m1" {
		t.Fatalf("DbError.Error 异常")
	}

	var ce = ConvDbErr(e)
	if ce.Code != "E1" {
		t.Fatalf("ConvDbErr(DbError) 异常")
	}

	ce = ConvDbErr(errors.New("x"))
	if ce.Msg != "x" {
		t.Fatalf("ConvDbErr(error) 异常")
	}

	if !IsRecordNotFound(ErrRecordNotFound) {
		t.Fatalf("IsRecordNotFound 应为 true")
	}
	if IsRecordNotFound(errors.New("a")) {
		t.Fatalf("IsRecordNotFound 普通错误应为 false")
	}
}

func TestToMapsAndMysqlGetters(t *testing.T) {
	var ctx = context.Background()
	var in = []map[string]int{{"a": 1}, {"b": 2}}
	var out []map[string]any
	var isStruct bool
	var err error
	out, isStruct, err = toMaps(ctx, in, true)
	if err != nil || len(out) != 2 {
		t.Fatalf("toMaps 基础转换异常: %v %v %v", out, isStruct, err)
	}

	var s = NewSql("users")
	s.Ctx(ctx)
	s.As("u")
	s.Fields("id,name")
	s.Raw("SELECT 1")
	s.Where("id = ?", 1)
	s.Join("t2", "users.id=t2.id")
	s.Limit("?", 10)
	s.Offset(2)
	s.Group("id")
	s.Order("id desc")
	s.ReplaceFields(map[string]string{"id": "uid"})
	s.OmitFields("x")
	s.IgnoreDuplicate(map[string]bool{"name": true})
	s.SetDuplicate(map[string]string{"name": "VALUES(name)"})

	if s.GetCtx() == nil || s.GetTableAlias() != "u" {
		t.Fatalf("GetCtx/GetTableAlias 异常")
	}
	if s.GetRawSql() == nil || s.GetFields() == nil || s.GetWhereCtrl() == nil {
		t.Fatalf("基础 getter 异常")
	}
	if s.GetJoinCtrl() == nil || s.GetLimitCtrl() == nil || s.GetOffsetCtrl() == nil {
		t.Fatalf("join/limit/offset getter 异常")
	}
	if s.GetGroup() == nil || s.GetOrder() == nil {
		t.Fatalf("group/order getter 异常")
	}
	if s.GetFieldReplace()["id"] != "uid" {
		t.Fatalf("GetFieldReplace 异常")
	}
	if s.GetFieldOmit()["x"] != "x" {
		t.Fatalf("GetFieldOmit 异常")
	}
	if !s.GetIgnoreDuplicate()["name"] {
		t.Fatalf("GetIgnoreDuplicate 异常")
	}
	if s.GetSetDuplicate()["name"] == "" {
		t.Fatalf("GetSetDuplicate 异常")
	}
}

func TestOrderedMapExtraAndJSONTypeUnmarshal(t *testing.T) {
	var om = NewOrderedMap[string, int]()
	om.SetKey([]string{"k1", "k2"})
	om.SetValues(map[string]int{"k1": 1, "k2": 2})
	if !reflect.DeepEqual(om.Values(), []int{1, 2}) {
		t.Fatalf("Values 顺序异常: %v", om.Values())
	}

	var sum int
	om.All(func(k string, v int) bool {
		sum += v
		return true
	})
	if sum != 3 {
		t.Fatalf("All 遍历异常")
	}

	var pairCount int
	om.AllPairs(func(value KeyValue[string, int]) bool {
		pairCount++
		return true
	})
	if pairCount != 2 {
		t.Fatalf("AllPairs 遍历异常")
	}

	type payload struct {
		A int `json:"a"`
	}
	var jt JSONType[payload]
	var err error
	err = jt.UnmarshalJSON([]byte(`{"a":7}`))
	if err != nil || jt.Data.A != 7 {
		t.Fatalf("JSONType.UnmarshalJSON 异常: %v %#v", err, jt.Data)
	}
}

func TestDbTxAndDbSetterMethods(t *testing.T) {
	var d = &Db{gs: NewSql("users")}
	d.log = defLog.Clone()
	d.WriteLog(true, true).WriteHhDbLog(true).WriteErrSql(false).WriteCompSql(false)
	if !d.writeLog || !d.writeHhDbLog || d.writeErrSql || d.writeCompSql {
		t.Fatalf("Db 写日志开关方法异常")
	}
	d.LogLevel(DebugLogLevel)
	d.EmptyError(errors.New("e"))
	if d.emptyError == nil {
		t.Fatalf("EmptyError 设置异常")
	}

	var tx = &DbTx{Db: d}
	if tx.GetTx() != nil {
		t.Fatalf("GetTx 默认应为空")
	}
	var err error
	err = tx.Transaction(func(db *DbTx) error { return nil })
	if err == nil {
		t.Fatalf("DbTx.Transaction 应返回不支持嵌套错误")
	}
}

func TestWithLogAndSetDefLogTrace(t *testing.T) {
	var c = &Conf{}
	var lg = &gdbLog{level: InfoLogLevel}
	WithLog(lg)(c)
	if c.Log == nil {
		t.Fatalf("WithLog 未生效")
	}

	SetDefLog(lg)
	SetTraceFunc(func(ctx context.Context) string { return "t" })
	SetLevel(WarnLogLevel)
	if lg.level != WarnLogLevel {
		t.Fatalf("SetLevel 未生效")
	}

	var now = time.Now().Unix()
	var bts []byte
	var err error
	bts, err = json.Marshal(now)
	if err != nil || len(bts) == 0 {
		t.Fatalf("基础 marshal 校验失败")
	}
}
