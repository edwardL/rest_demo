package gdb

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReadOnlyMapMethods(t *testing.T) {
	var rm = NewReadOnlyMap(map[string]any{"a": 1, "b": "x"})
	if rm.Len() != 2 {
		t.Fatalf("Len 错误")
	}
	var v any
	var ok bool
	v, ok = rm.Get("a")
	if !ok || v.(int) != 1 {
		t.Fatalf("Get 错误")
	}
	var keys = rm.Keys()
	if len(keys) != 2 {
		t.Fatalf("Keys 错误: %v", keys)
	}
	var count int
	rm.Range(func(key string, val any) bool {
		count++
		return true
	})
	if count != 2 {
		t.Fatalf("Range 遍历错误")
	}
}

func TestOrderedMapMethods(t *testing.T) {
	var om = NewOrderedMap[string, int]()
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("a", 3)
	var v int
	var ok bool
	v, ok = om.Get("a")
	if !ok || v != 3 {
		t.Fatalf("Get 错误")
	}
	var keys = om.Keys()
	if !reflect.DeepEqual(keys, []string{"a", "b"}) {
		t.Fatalf("Keys 顺序错误: %v", keys)
	}
	om.Delete("a")
	keys = om.Keys()
	if !reflect.DeepEqual(keys, []string{"b"}) {
		t.Fatalf("Delete 错误: %v", keys)
	}
}

func TestOrderedMapJSON(t *testing.T) {
	var om = NewOrderedMap[string, int]()
	om.Set("a", 1)
	om.Set("b", 2)
	var data []byte
	var err error
	data, err = om.MarshalJSON()
	if err != nil || len(data) == 0 {
		t.Fatalf("MarshalJSON 错误: %v", err)
	}

	var target OrderedMap[string, int]
	err = target.UnmarshalJSON(data)
	if err != nil {
		t.Fatalf("UnmarshalJSON 错误: %v", err)
	}
	if target.Len() == 0 {
		t.Fatalf("UnmarshalJSON 结果为空")
	}
}

func (om *OrderedMap[K, T]) Len() int {
	return len(om.keys)
}

func TestJSONType(t *testing.T) {
	type body struct {
		Name string `json:"name"`
	}
	var j JSONType[body]
	j.Data = body{Name: "alice"}
	var v any
	var err error
	v, err = j.Value()
	if err != nil || len(v.([]byte)) == 0 {
		t.Fatalf("JSONType Value 错误: %v", err)
	}

	var j2 JSONType[body]
	err = j2.Scan([]byte(`{"name":"bob"}`))
	if err != nil || j2.Data.Name != "bob" {
		t.Fatalf("JSONType Scan 错误: %v %#v", err, j2.Data)
	}

	var b []byte
	b, err = json.Marshal(j2)
	if err != nil || len(b) == 0 {
		t.Fatalf("JSONType MarshalJSON 错误: %v", err)
	}
}

func TestENUMType(t *testing.T) {
	var e ENUMType[int] = []int{1, 2}
	var v any
	var err error
	v, err = e.Value()
	if err != nil || v.(string) != "1,2" {
		t.Fatalf("ENUMType Value 错误: %v %v", v, err)
	}

	var e2 ENUMType[int]
	err = e2.Scan("3,4")
	if err != nil || len(e2) != 2 || e2[0] != 3 {
		t.Fatalf("ENUMType Scan(int) 错误: %v %v", e2, err)
	}

	var e3 ENUMType[string]
	err = e3.Scan([]byte("a,b"))
	if err != nil || len(e3) != 2 || e3[1] != "b" {
		t.Fatalf("ENUMType Scan(string) 错误: %v %v", e3, err)
	}
}

func TestPageAndWrapper(t *testing.T) {
	var p = NewPage[int](2, 20)
	if p.CurrPage != 2 || p.PageNums != 20 || p.TotalNums != 0 {
		t.Fatalf("NewPage 初始化错误: %+v", p)
	}

	var w = QueryWrapper()
	if w == nil {
		t.Fatalf("QueryWrapper 为空")
	}
	_ = w.GetSql()
	_ = w.GetArgs()
	_ = w.GetWhereSql()
	_ = w.GetWhereArgs()
}

func TestStreamCallback(t *testing.T) {
	type row struct {
		Id int `json:"id" gdb:"id"`
	}
	var called bool
	var sc = StreamCallback[row](func(r row) (bool, error) {
		called = true
		if r.Id != 10 {
			return false, nil
		}
		return true, nil
	})
	var next bool
	var err error
	next, err = sc.call(map[string]any{"id": 10}, nil, true)
	if err != nil || !next || !called {
		t.Fatalf("StreamCallback 结构体转换失败: next=%v err=%v called=%v", next, err, called)
	}
}
