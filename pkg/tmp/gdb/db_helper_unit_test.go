package gdb

import "testing"

func TestDbHelperWithoutDbInit(t *testing.T) {
	var affected int64
	var insertId int64
	var err error

	affected, insertId, err = Exec("THIS IS INVALID SQL")
	if err == nil {
		t.Fatalf("Exec 非法 SQL 应返回错误")
	}
	if affected != 0 || insertId != 0 {
		t.Fatalf("未初始化数据库时返回值应为0")
	}

	var one *genericUser
	one, err = QueryOne[genericUser]("SELECT xxxxxx FROM table_not_exists")
	if err == nil {
		t.Fatalf("QueryOne 非法 SQL 应返回错误")
	}
	if one != nil {
		t.Fatalf("未初始化数据库时 QueryOne 返回值应为空")
	}
}
