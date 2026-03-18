package gdb

import (
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"testing"
)

func newSearchParamSample() *HhSearchParam {
	var sp = &types.SearchParams{
		SearchParams: []types.SearchItem{
			{FieldName: "name", FieldCmpOp: "=", FieldValue: "tom"},
			{FieldName: "age", FieldCmpOp: ">", FieldValue: "18"},
			{FieldName: "status", FieldCmpOp: "=", FieldValue: 1},
		},
		SortParams: "name desc,age asc",
	}
	return NewSearchParamsHandle(sp)
}

func TestSearchParamBasicMethods(t *testing.T) {
	var p = newSearchParamSample()
	if p.GetSearchParams() == nil || len(p.GetSearchItem()) != 3 {
		t.Fatalf("GetSearchParams/GetSearchItem 错误")
	}

	var v any
	var err error
	v, err = p.GetSearchValue("name")
	if err != nil || v.(string) != "tom" {
		t.Fatalf("GetSearchValue 错误: %v %v", v, err)
	}

	var i int
	i, err = p.GetSearchIntValue("age")
	if err != nil || i != 18 {
		t.Fatalf("GetSearchIntValue 错误: %d %v", i, err)
	}

	var s string
	s, err = p.GetSearchStringValue("status")
	if err != nil || s != "1" {
		t.Fatalf("GetSearchStringValue 错误: %s %v", s, err)
	}

	var op string
	op, err = p.GetSearchOp("age")
	if err != nil || op != ">" {
		t.Fatalf("GetSearchOp 错误: %s %v", op, err)
	}
}

func TestSearchParamReplaceAndKeep(t *testing.T) {
	var p = newSearchParamSample()
	p.ReplaceFieldName("name", "user_name")
	var _, err = p.GetSearchValue("name")
	if err == nil {
		t.Fatalf("ReplaceFieldName 未生效")
	}

	p.ReplaceFieldValue("user_name", func(v any) any {
		return v.(string) + "_x"
	})
	var nv any
	nv, err = p.GetSearchValue("user_name")
	if err != nil || nv.(string) != "tom_x" {
		t.Fatalf("ReplaceFieldValue 错误: %v %v", nv, err)
	}

	p.RemoveSearch("status")
	if len(p.GetSearchItem()) != 2 {
		t.Fatalf("RemoveSearch 错误: %v", p.GetSearchItem())
	}

	p.KeepSearch("user_name")
	if len(p.GetSearchItem()) != 1 || p.GetSearchItem()[0].FieldName != "user_name" {
		t.Fatalf("KeepSearch 错误: %v", p.GetSearchItem())
	}
}

func TestSearchParamSortAndRequired(t *testing.T) {
	var p = newSearchParamSample()
	p.ReplaceSort(map[string]string{"name": "u.name", "age": "u.age"})
	if p.SortParams != "u.name desc,u.age asc" {
		t.Fatalf("ReplaceSort 错误: %s", p.SortParams)
	}

	p.KeepSort("u.name")
	if p.SortParams != "u.name desc" {
		t.Fatalf("KeepSort 错误: %s", p.SortParams)
	}

	var err error
	err = p.RequiredSearch("user_not_exist", "必须参数")
	if err == nil {
		t.Fatalf("RequiredSearch 缺参应报错")
	}

	p.AddSearch("x", "=", 1)
	err = p.RequiredSearch("x", "")
	if err != nil {
		t.Fatalf("AddSearch 或 RequiredSearch 错误: %v", err)
	}
}
