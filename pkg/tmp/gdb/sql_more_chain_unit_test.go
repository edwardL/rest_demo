package gdb

import (
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"strings"
	"testing"
)

func TestSqlJoinAndWhereControls(t *testing.T) {
	var s = NewSql("users")
	s.Join("t1", "users.id=t1.uid")
	s.InnerJoin("t2", "users.id=t2.uid")
	s.LeftJoin("t3", "users.id=t3.uid")
	s.RightJoin("t4", "users.id=t4.uid")
	if s.GetJoinLen() != 4 {
		t.Fatalf("Join 数量错误: %d", s.GetJoinLen())
	}

	s.Where("id = ?", 1).WhereOr("name = ?", "tom")
	if s.GetWhereLen() != 2 {
		t.Fatalf("Where 数量错误: %d", s.GetWhereLen())
	}

	s.WhereReset()
	if s.GetWhereLen() != 0 {
		t.Fatalf("WhereReset 未生效")
	}

	s.WhereGroup(func(gs *Sql) (err error) {
		gs.Where("a=?", 1)
		return nil
	})
	s.WhereGroupOr(func(gs *Sql) (err error) {
		gs.Where("b=?", 2)
		return nil
	})
	if s.GetWhereLen() != 2 {
		t.Fatalf("WhereGroup/WhereGroupOr 数量错误: %d", s.GetWhereLen())
	}
}

func TestSqlFieldGroupOrderAndFilter(t *testing.T) {
	var s = NewSql("users")
	s.FieldsByData(map[string]any{"name": "", "status": 1}, map[string]string{"status": ""})
	if !strings.Contains(s.GetFields().Query.String(), "name") {
		t.Fatalf("FieldsByData 构建异常: %s", s.GetFields().Query.String())
	}

	s.FieldsByDataAlias(map[string]any{"name": ""}, nil, "u")
	if !strings.Contains(s.GetFields().Query.String(), "u.") {
		t.Fatalf("FieldsByDataAlias 构建异常: %s", s.GetFields().Query.String())
	}

	s.OmitFields([]string{"x", "y"})
	s.ReplaceFields(map[string]string{"name": "user_name"})
	if s.GetFieldOmit()["x"] != "x" || s.GetFieldReplace()["name"] != "user_name" {
		t.Fatalf("Omit/ReplaceFields 未生效")
	}

	s.Group([]string{"name"})
	if !strings.Contains(s.GetGroup().Query.String(), "GROUP BY") {
		t.Fatalf("Group 构建异常: %s", s.GetGroup().Query.String())
	}

	s.Order([]string{"name DESC"})
	if !strings.Contains(s.GetOrder().Query.String(), "ORDER BY") {
		t.Fatalf("Order 构建异常: %s", s.GetOrder().Query.String())
	}

	s.OrderByFilter("name desc,id asc", map[string]string{"name": "u.name"}, true)
	if !strings.Contains(s.GetOrder().Query.String(), "ORDER BY") {
		t.Fatalf("OrderByFilter 构建异常: %s", s.GetOrder().Query.String())
	}
}

func TestSqlWhereSearchAndCte(t *testing.T) {
	var s = NewSql("users")
	var items = []types.SearchItem{
		{FieldName: "name", FieldCmpOp: "like", FieldValue: "tom"},
		{FieldName: "age", FieldCmpOp: ">", FieldValue: 18},
	}
	s.WhereSearch(items, map[string]string{"name": "u.name", "age": "u.age"})
	if s.Err() != nil || s.GetWhereLen() == 0 {
		t.Fatalf("WhereSearch 构建失败: err=%v len=%d", s.Err(), s.GetWhereLen())
	}

	var view = types.DataViewSearchInfo{FieldList: map[string]types.DataViewField{
		"name": {FieldAlias: "name"},
		"age":  {FieldAlias: "age"},
	}}
	s.WhereSearchView(view, items, map[string]string{"name": "name", "age": "age"})
	if s.Err() != nil {
		t.Fatalf("WhereSearchView 构建失败: %v", s.Err())
	}

	var q = NewCteQuery().SetCet("WITH RECURSIVE")
	q.SetCteTable("u1", NewSql("users").Where("id> ?", 1))
	s.CteQuery(q)
	var res *Result
	var err error
	res, err = s.Select()
	if err != nil {
		t.Fatalf("CteQuery Select 构建失败: %v", err)
	}
	if !strings.Contains(strings.ToUpper(res.Sql.String()), "WITH") {
		t.Fatalf("CTE SQL 缺少 WITH: %s", res.Sql.String())
	}
}
