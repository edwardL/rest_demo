package gdb

import (
	"context"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"testing"
)

func TestDbChainPassthroughAndHelpers(t *testing.T) {
	var d = &Db{gs: NewSql("users")}
	d.log = defLog.Clone()
	var wrap = QueryWrapper()
	wrap.Where("d=?", 4)

	d.Joins("JOIN t2 ON t2.id = users.id").
		Join("t3", "t3.id=users.id").
		InnerJoin("t4", "t4.id=users.id").
		LeftJoin("t5", "t5.id=users.id").
		RightJoin("t6", "t6.id=users.id").
		Where("id = ?", 1).
		WhereOr("name = ?", "tom").
		WhereCond(CondAnd, "ts > ?", 1).
		WhereBlock("status = ?", 1).
		WhereBlockOr("status = ?", 2).
		WhereGroup(func(db *Db) error { db.Where("a=?", 1); return nil }).
		WhereGroupOr(func(db *Db) error { db.Where("b=?", 2); return nil }).
		WhereGroupCond(CondAnd, func(db *Db) error { db.Where("c=?", 3); return nil }).
		Wrapper(wrap).
		Table("users").
		As("u").
		Field("id").
		Fields("name").
		FieldsByData(map[string]any{"name": ""}, nil).
		FieldsByDataAlias(map[string]any{"name": ""}, nil, "u").
		OmitFields("x").
		ReplaceFields(map[string]string{"name": "user_name"}).
		Limit("?", 10).
		Limits(0, 10).
		Offset(0).
		Page(1, 10).
		PageReset().
		Group("id").
		Groups("id").
		Order("id desc").
		Orders("id desc").
		OrderByFilter("id desc", map[string]string{"id": "id"}, true).
		WhereSearch([]types.SearchItem{{FieldName: "id", FieldCmpOp: "=", FieldValue: 1}}, map[string]string{"id": "id"}, true).
		WhereSearchParams(NewSearchParamsHandle(&types.SearchParams{SortParams: "id desc", CurrPage: 1, PageNums: 10})).
		Union(&Db{gs: NewSql("u1")}).
		UnionAll(&Db{gs: NewSql("u2")}).
		Raw("SELECT 1").
		ConvFieldsType(func(key string, val any, m *ReadOnlyMap) any { return val }).
		Ctx(context.Background()).
		WriteLog(true).
		WriteHhDbLog(true).
		WriteErrSql(false).
		WriteCompSql(false).
		LogLevel(InfoLogLevel).
		LogCallDepth(2).
		EmptyError().
		Tx(nil)

	if d.GetGSql() == nil {
		t.Fatalf("GetGSql 为空")
	}
	if d.GetJoinLen() < 0 || d.GetWhereLen() < 0 {
		t.Fatalf("GetJoinLen/GetWhereLen 返回异常")
	}
	if d.GetTx() != nil {
		t.Fatalf("GetTx 异常")
	}

	var cloned = d.Clone()
	if cloned == nil {
		t.Fatalf("Clone 失败")
	}
}
