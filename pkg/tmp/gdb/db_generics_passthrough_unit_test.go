package gdb

import (
	"context"
	"testing"
)

func TestDbTChainPassthroughs(t *testing.T) {
	var d = &DbT[genericUser]{Db: &Db{gs: NewSql("generic_users")}}
	d.Db.log = defLog.Clone()
	d.Joins("JOIN t2 ON t2.id = generic_users.id").
		Join("t3", "t3.id = generic_users.id").
		InnerJoin("t4", "t4.id = generic_users.id").
		LeftJoin("t5", "t5.id = generic_users.id").
		RightJoin("t6", "t6.id = generic_users.id").
		Where("id = ?", 1).
		WhereOr("name = ?", "tom").
		WhereCond(CondAnd, "ts > ?", 1).
		WhereBlock("name = ?", "tom").
		WhereBlockOr("name = ?", "jack").
		WhereGroup(func(db *Db) error { db.Where("x = ?", 1); return nil }).
		WhereGroupOr(func(db *Db) error { db.Where("y = ?", 2); return nil }).
		WhereGroupCond(CondAnd, func(db *Db) error { db.Where("z = ?", 3); return nil }).
		Field("id").
		Fields("name").
		FieldsByData(genericUser{}, nil).
		FieldsByDataAlias(genericUser{}, nil, "u").
		OmitFields("x").
		ReplaceFields(map[string]string{"id": "uid"}).
		Limit("?", 10).
		Limits(0, 10).
		Offset(0).
		Page(1, 10).
		Group("id").
		Groups("id").
		Order("id desc").
		Orders("id desc").
		OrderByFilter("id desc", map[string]string{"id": "id"}, true).
		Union(New("u1")).
		UnionAll(New("u2")).
		Raw("SELECT 1").
		ConvFieldsType(func(key string, val any, m *ReadOnlyMap) any { return val }).
		Ctx(context.Background()).
		WriteLog(true).
		WriteHhDbLog(true).
		WriteErrSql(false).
		WriteCompSql(false).
		LogLevel(InfoLogLevel).
		LogCallDepth(1).
		EmptyError()

	if d.GetGSql() == nil {
		t.Fatalf("链式透传后 GetGSql 为空")
	}
}
