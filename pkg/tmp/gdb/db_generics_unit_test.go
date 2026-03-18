package gdb

import (
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"testing"
)

type genericUser struct {
	Id   int    `json:"id" gdb:"id"`
	Name string `json:"name" gdb:"name"`
}

func (genericUser) TableName() string {
	return "generic_users"
}

func TestModelAndGenericBasic(t *testing.T) {
	var m = Model[genericUser]()
	if m == nil || m.GetGSql() == nil {
		t.Fatalf("Model 初始化失败")
	}
	if m.GetGSql().GetTable() != "generic_users" {
		t.Fatalf("Model 表推断失败: %s", m.GetGSql().GetTable())
	}

	var err error
	err = m.Scan()
	if err == nil {
		t.Fatalf("Scan 应返回不支持错误")
	}
	err = m.ScanAndCount()
	if err == nil {
		t.Fatalf("ScanAndCount 应返回不支持错误")
	}
}

func TestDbTWhereSearchParamsAndPage(t *testing.T) {
	var sp = &types.SearchParams{
		CurrPage: 2,
		PageNums: 15,
	}
	var h = NewSearchParamsHandle(sp)
	var m = Model[genericUser]().WhereSearchParams(h)
	if m.pageInfo == nil || m.pageInfo.CurrPage != 2 || m.pageInfo.PageNums != 15 {
		t.Fatalf("WhereSearchParams 分页信息设置失败: %+v", m.pageInfo)
	}

	m.PageReset()
	if m.pageInfo != nil {
		t.Fatalf("PageReset 未重置 pageInfo")
	}
}

func TestDbTSetLastInsertId(t *testing.T) {
	var m = Model[genericUser]()
	var u = &genericUser{}
	m.setLastInsertId(u, 101)
	if u.Id != 101 {
		t.Fatalf("setLastInsertId 失败: %d", u.Id)
	}
}
