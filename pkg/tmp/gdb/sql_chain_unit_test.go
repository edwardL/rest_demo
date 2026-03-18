package gdb

import (
	"strings"
	"testing"
)

func TestSqlChainSelectBuild(t *testing.T) {
	var res *Result
	var err error
	res, err = NewSql("users u").
		Fields("u.id, u.name").
		Where("u.id IN (?)", []int{1, 2, 3}).
		WhereGroup(func(tx *Sql) (err error) {
			tx.Where("u.name = ?", "tom").WhereOr("u.name = ?", "jack")
			return nil
		}).
		Order("u.id DESC").
		Page(1, 10).
		Select()
	if err != nil {
		t.Fatalf("Select 构建失败: %v", err)
	}
	var sql = res.CompSql()
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "FROM users u") {
		t.Fatalf("Select SQL 异常: %s", sql)
	}
	if !strings.Contains(sql, "ORDER BY") || !strings.Contains(sql, "LIMIT") {
		t.Fatalf("排序分页缺失: %s", sql)
	}
}

func TestSqlUpdateDeleteExistsBuild(t *testing.T) {
	var res *Result
	var err error

	res, err = NewSql("users").Where("id = ?", 1).Update("name", "new")
	if err != nil || !strings.Contains(res.CompSql(), "UPDATE users") {
		t.Fatalf("Update SQL 异常: %v %v", res, err)
	}

	res, err = NewSql("users").Where("id = ?", 1).Delete()
	if err != nil || !strings.Contains(res.CompSql(), "DELETE") {
		t.Fatalf("Delete SQL 异常: %v %v", res, err)
	}

	res, err = NewSql("users").Where("id = ?", 1).Exists()
	if err != nil || !strings.Contains(res.CompSql(), "SELECT EXISTS") {
		t.Fatalf("Exists SQL 异常: %v %v", res, err)
	}
}

func TestSqlJoinUnionBuild(t *testing.T) {
	var q1 = NewSql("users").Where("id > ?", 1)
	var q2 = NewSql("users").Where("id < ?", 10)
	var res *Result
	var err error

	res, err = NewSql("users u").
		LeftJoin("profile p", "u.id = p.user_id").
		Where("u.id = ?", 1).
		Select()
	if err != nil || !strings.Contains(res.CompSql(), "LEFT JOIN") {
		t.Fatalf("Join SQL 异常: %v %v", res, err)
	}

	res, err = NewSql("users").Where("id > ?", 0).Union(q1, q2).Select()
	if err != nil || !strings.Contains(strings.ToUpper(res.CompSql()), "UNION") {
		t.Fatalf("Union SQL 异常: %v %v", res, err)
	}
}
