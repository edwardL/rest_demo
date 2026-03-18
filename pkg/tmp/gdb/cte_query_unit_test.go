package gdb

import "testing"

func TestCteQueryBasic(t *testing.T) {
	var cte = NewCteQuery().SetCet("WITH")
	if cte.Cte != "WITH" {
		t.Fatalf("SetCet 失败")
	}

	cte.SetCteTable("u", NewSql("users"))
	if len(cte.VirtualTable) != 1 || cte.VirtualTable[0].TableAlias != "u" {
		t.Fatalf("SetCteTable(*Sql) 失败: %+v", cte.VirtualTable)
	}

	cte.SetCteTable("d", New("demo"))
	if len(cte.VirtualTable) != 2 || cte.VirtualTable[1].TableAlias != "d" {
		t.Fatalf("SetCteTable(DbFace) 失败: %+v", cte.VirtualTable)
	}
}

func TestCteQueryInvalidType(t *testing.T) {
	var cte = NewCteQuery()
	cte.SetCteTable("x", 123)
	if cte.err == nil {
		t.Fatalf("非法类型应产生错误")
	}

	var before = len(cte.VirtualTable)
	cte.SetCteTable("y", NewSql("users"))
	if len(cte.VirtualTable) != before {
		t.Fatalf("错误状态下不应继续写入虚拟表")
	}
}
