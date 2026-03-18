package gdb

import (
	"reflect"
	"sync/atomic"
	"testing"
)

type toolModel struct {
	Name string `json:"name" gdb:"name"`
}

func (toolModel) TableName() string {
	return "tb_tool"
}

func (toolModel) TableAlias() string {
	return "tt"
}

func TestMapToMapAnySeries(t *testing.T) {
	var mAny = map[string]any{"a": 1}
	var got map[string]any
	var err error

	got, err = MapToMapAny(mAny)
	if err != nil || got["a"].(int) != 1 {
		t.Fatalf("map[string]any 转换失败: %v %v", got, err)
	}

	got, err = MapToMapAny(&mAny)
	if err != nil || got["a"].(int) != 1 {
		t.Fatalf("*map[string]any 转换失败: %v %v", got, err)
	}

	var mRaw = map[string]RawBody{"a": RawBody("x")}
	got, err = MapToMapAny(mRaw)
	if err != nil || string(got["a"].(RawBody)) != "x" {
		t.Fatalf("map[string]RawBody 转换失败: %v %v", got, err)
	}

	var mStr = map[string]string{"a": "v"}
	got, err = MapToMapAny(mStr)
	if err != nil || got["a"].(string) != "v" {
		t.Fatalf("map[string]string 转换失败: %v %v", got, err)
	}

	var mInt = map[string]int{"a": 2}
	got, err = MapToMapAny(mInt)
	if err != nil || got["a"].(int) != 2 {
		t.Fatalf("map[string]int 反射转换失败: %v %v", got, err)
	}

	_, err = MapToMapAny(map[int]string{1: "x"})
	if err == nil {
		t.Fatalf("非法 map 类型应返回错误")
	}
}

func TestMapsToMapsAnySeries(t *testing.T) {
	var got []map[string]any
	var err error

	got, err = MapsToMapsAny(nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("nil 输入转换失败: len=%d err=%v", len(got), err)
	}

	var data = []map[string]string{{"a": "1"}, {"b": "2"}}
	got, err = MapsToMapsAny(data)
	if err != nil || len(got) != 2 || got[0]["a"].(string) != "1" {
		t.Fatalf("[]map[string]string 转换失败: %v %v", got, err)
	}

	var listAny = []any{map[string]int{"a": 1}, map[string]any{"b": 2}}
	got, err = MapsToMapsAny(listAny)
	if err != nil || len(got) != 2 {
		t.Fatalf("[]any map 转换失败: %v %v", got, err)
	}

	_, err = MapsToMapsAny([]any{1, 2})
	if err == nil {
		t.Fatalf("非法 []any 输入应返回错误")
	}
}

func TestStructDbFieldAndHelpers(t *testing.T) {
	var fields = StructDbField(toolModel{})
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("StructDbField 基础结果错误: %v", fields)
	}

	fields = StructDbField(toolModel{}, "")
	if len(fields) != 1 || fields[0] != "`tt`.`name`" {
		t.Fatalf("StructDbField 自动别名错误: %v", fields)
	}

	fields = StructDbField(toolModel{}, "x")
	if len(fields) != 1 || fields[0] != "`x`.`name`" {
		t.Fatalf("StructDbField 指定别名错误: %v", fields)
	}

	if !InArr(2, []int{1, 2, 3}) {
		t.Fatalf("InArr 判断错误")
	}

	if As("a", "AS", "b") != "a AS b" {
		t.Fatalf("As 拼接错误")
	}

	var like = EscapeLike("a%b_c")
	if like != "%a\\%b\\_c%" {
		t.Fatalf("EscapeLike 转义错误: %s", like)
	}

	if GetTableName("tb") != "tb" {
		t.Fatalf("GetTableName string 错误")
	}
	if GetTableName(toolModel{}) != "tb_tool" {
		t.Fatalf("GetTableName model 错误")
	}
}

func TestQueryPlaceholderHelpers(t *testing.T) {
	var idx int
	idx = strIndex("a?b?c", "?", 2)
	if idx != 3 {
		t.Fatalf("strIndex 结果错误: %d", idx)
	}

	var replaced string
	var err error
	replaced, err = replaceIndex("a?b", 1, "x")
	if err != nil || replaced != "axb" {
		t.Fatalf("replaceIndex 失败: %s %v", replaced, err)
	}

	replaced, err = genPrePil("id in (?)", 1, 3)
	if err != nil || replaced != "id in (?,?,?)" && replaced != "id in (?, ?, ?)" {
		t.Fatalf("genPrePil 失败: %s %v", replaced, err)
	}

	replaced, err = genPrePilGroup("(a,b) in (?)", 1, 2, 2)
	if err != nil || replaced != "(a,b) in ((?, ?), (?, ?))" {
		t.Fatalf("genPrePilGroup 失败: %s %v", replaced, err)
	}

	var phIndex atomic.Int32
	phIndex.Store(1)
	var query string
	var args []any
	query, args, err = genQueryList("id in (?)", []int{1, 2}, &phIndex)
	if err != nil || (query != "id in (?, ?)" && query != "id in (?,?)") || len(args) != 2 {
		t.Fatalf("genQueryList 失败: %s %v %v", query, args, err)
	}

	phIndex.Store(1)
	query, args, err = genQueryGroupAnyList("(a,b) in (?)", [][]any{{1, 2}, {3, 4}}, &phIndex)
	if err != nil || len(args) != 4 {
		t.Fatalf("genQueryGroupAnyList 失败: %s %v %v", query, args, err)
	}
}

func TestArrayConvertAndOrderValidate(t *testing.T) {
	var in []any = []any{[]int{1, 2}, []string{"a", "b"}, []bool{true, false}, []any{3, "x"}}
	var out [][]any
	var err error
	out, err = arrToArrList(in)
	if err != nil || len(out) != 4 || out[2][0].(int) != 1 || out[2][1].(int) != 0 {
		t.Fatalf("arrToArrList 结果异常: %v %v", out, err)
	}

	_, err = arrToArrList([]any{[]struct{}{{}}})
	if err == nil {
		t.Fatalf("arrToArrList 非法类型应返回错误")
	}

	var order string
	var illegal []string
	order, illegal, err = ValidateOrderParam("name desc, age asc", nil)
	if err != nil || order != "name desc, age asc" || len(illegal) != 0 {
		t.Fatalf("ValidateOrderParam 基础校验失败: %s %v %v", order, illegal, err)
	}

	var fm = map[string]string{"name": "u.name", "age": "u.age"}
	order, illegal, err = ValidateOrderParam("name desc, id asc", fm)
	if err != nil || len(illegal) != 1 || illegal[0] != "id" || order != "u.name desc" {
		t.Fatalf("ValidateOrderParam 字段替换失败: %s %v %v", order, illegal, err)
	}

	_, _, err = ValidateOrderParam("name delete", nil)
	if err == nil {
		t.Fatalf("非法排序规则应返回错误")
	}
}

func TestGetTableNameRecursiveCases(t *testing.T) {
	var m = toolModel{}
	var pm = &m
	if GetTableName(pm) != "tb_tool" {
		t.Fatalf("GetTableName 指针模型失败")
	}

	var list = []toolModel{{Name: "a"}}
	if GetTableName(list) != "tb_tool" {
		t.Fatalf("GetTableName 切片模型失败")
	}

	var empty string
	empty = GetTableName(struct{ A int }{A: 1})
	if empty != "" {
		t.Fatalf("不实现 ModelFace 应返回空字符串")
	}
}

func TestStrToArr(t *testing.T) {
	var arr = strToArr("a,,b", ",", true)
	if !reflect.DeepEqual(arr, []string{"a", "b"}) {
		t.Fatalf("strToArr delEmpty=true 失败: %v", arr)
	}

	arr = strToArr("a,,b", ",", false)
	if !reflect.DeepEqual(arr, []string{"a", "", "b"}) {
		t.Fatalf("strToArr delEmpty=false 失败: %v", arr)
	}
}
