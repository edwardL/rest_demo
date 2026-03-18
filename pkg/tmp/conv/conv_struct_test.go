package conv

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

type testBase struct {
	Id        int       `json:"id" gdb:"id"`
	CreatedAt time.Time `json:"created_at" gdb:"created_at;type:Unix"`
}

type testUser struct {
	testBase
	Name  string `json:"name" gdb:"name"`
	Score *int   `json:"score" gdb:"score"`
	Skip  string `json:"skip" gdb:"-"`
}

type testNoId struct {
	Name string
}

type testRawMessage struct {
	Id      int              `json:"id" gdb:"id"`
	Name    string           `json:"name" gdb:"name"`
	Data    json.RawMessage  `json:"data" gdb:"data"`
	PtrData *json.RawMessage `json:"ptr_data" gdb:"ptr_data"`
}

func TestStructToMapAndMapToStruct(t *testing.T) {
	var score = 9
	var user = testUser{
		testBase: testBase{
			Id:        1,
			CreatedAt: time.Unix(1700000000, 0),
		},
		Name:  "alice",
		Score: &score,
		Skip:  "ignored",
	}

	var m map[string]any
	var err error
	m, err = StructToMap(user, false)
	if err != nil {
		t.Fatalf("StructToMap unexpected error: %v", err)
	}
	if m["id"].(int) != 1 {
		t.Fatalf("StructToMap id mismatch, got=%v", m["id"])
	}
	if m["name"].(string) != "alice" {
		t.Fatalf("StructToMap name mismatch, got=%v", m["name"])
	}
	if m["created_at"].(int64) != 1700000000 {
		t.Fatalf("StructToMap created_at mismatch, got=%v", m["created_at"])
	}
	var ok bool
	_, ok = m["skip"]
	if ok {
		t.Fatalf("StructToMap should ignore skip field")
	}

	var target testUser
	err = MapToStruct(map[string]any{
		"id":         "2",
		"name":       "bob",
		"score":      "7",
		"created_at": "1700000000",
	}, &target, true)
	if err != nil {
		t.Fatalf("MapToStruct unexpected error: %v", err)
	}
	if target.Id != 2 {
		t.Fatalf("MapToStruct id mismatch, got=%d", target.Id)
	}
	if target.Name != "bob" {
		t.Fatalf("MapToStruct name mismatch, got=%s", target.Name)
	}
	if target.Score == nil || *target.Score != 7 {
		t.Fatalf("MapToStruct score mismatch")
	}
	if target.CreatedAt.Unix() != 1700000000 {
		t.Fatalf("MapToStruct created_at mismatch, got=%d", target.CreatedAt.Unix())
	}
}

func TestStructToMapsAndMapsToStruct(t *testing.T) {
	var score1 = 10
	var score2 = 20
	var users = []testUser{
		{
			testBase: testBase{Id: 1, CreatedAt: time.Unix(1700000000, 0)},
			Name:     "u1",
			Score:    &score1,
		},
		{
			testBase: testBase{Id: 2, CreatedAt: time.Unix(1700000100, 0)},
			Name:     "u2",
			Score:    &score2,
		},
	}

	var maps []map[string]any
	var err error
	maps, err = StructToMaps(users, false)
	if err != nil {
		t.Fatalf("StructToMaps unexpected error: %v", err)
	}
	if len(maps) != 2 {
		t.Fatalf("StructToMaps length mismatch, got=%d", len(maps))
	}

	var target []testUser
	err = MapsToStruct(maps, &target, true)
	if err != nil {
		t.Fatalf("MapsToStruct unexpected error: %v", err)
	}
	if len(target) != 2 {
		t.Fatalf("MapsToStruct length mismatch, got=%d", len(target))
	}
	if target[0].Name != "u1" || target[1].Name != "u2" {
		t.Fatalf("MapsToStruct values mismatch, got=%v", target)
	}
}

func TestGetStructFieldList(t *testing.T) {
	var fields = GetStructFieldList(reflect.TypeOf(testUser{}))
	var got = make(map[string]struct{})
	var field string
	for _, field = range fields {
		got[field] = struct{}{}
	}
	var existsId bool
	_, existsId = got["id"]
	if !existsId {
		t.Fatalf("GetStructFieldList missing id")
	}
	var existsName bool
	_, existsName = got["name"]
	if !existsName {
		t.Fatalf("GetStructFieldList missing name")
	}
}

func TestAToB(t *testing.T) {
	var src = map[string]any{
		"name": "tom",
		"age":  18,
	}
	var dst map[string]any
	var err = AToB(src, &dst)
	if err != nil {
		t.Fatalf("AToB unexpected error: %v", err)
	}
	if dst["name"].(string) != "tom" {
		t.Fatalf("AToB mismatch, got=%v", dst)
	}
}

func TestAssignId(t *testing.T) {
	var v testUser
	var err = AssignId(&v, int64(100))
	if err != nil {
		t.Fatalf("AssignId unexpected error: %v", err)
	}
	if v.Id != 100 {
		t.Fatalf("AssignId mismatch, got=%d", v.Id)
	}

	var noId testNoId
	err = AssignId(&noId, 1)
	if err == nil {
		t.Fatalf("AssignId should fail when id field missing")
	}
}

func TestStructToMapWithRawMessage(t *testing.T) {
	var jsonData = []byte(`{"key":"value","num":123}`)
	var rawMsg = json.RawMessage(jsonData)
	var ptrRawMsg = json.RawMessage([]byte(`{"nested":true}`))

	var user = testRawMessage{
		Id:      1,
		Name:    "test",
		Data:    rawMsg,
		PtrData: &ptrRawMsg,
	}

	// 测试结构体转 map
	m, err := StructToMap(user, false)
	if err != nil {
		t.Fatalf("StructToMap with RawMessage unexpected error: %v", err)
	}

	// 验证 ID
	if m["id"].(int) != 1 {
		t.Fatalf("StructToMap id mismatch, got=%v", m["id"])
	}

	// 验证 Name
	if m["name"].(string) != "test" {
		t.Fatalf("StructToMap name mismatch, got=%v", m["name"])
	}

	// 验证 RawMessage 字段
	dataVal, ok := m["data"]
	if !ok {
		t.Fatalf("StructToMap data field missing")
	}
	dataBytes, ok := dataVal.(json.RawMessage)
	if !ok {
		t.Fatalf("StructToMap data type mismatch, got=%T", dataVal)
	}
	if string(dataBytes) != `{"key":"value","num":123}` {
		t.Fatalf("StructToMap data content mismatch, got=%s", string(dataBytes))
	}

	// 验证指针类型的 RawMessage
	ptrDataVal, ok := m["ptr_data"]
	if !ok {
		t.Fatalf("StructToMap ptr_data field missing")
	}
	ptrDataBytes, ok := ptrDataVal.(json.RawMessage)
	if !ok {
		t.Fatalf("StructToMap ptr_data type mismatch, got=%T", ptrDataVal)
	}
	if ptrDataBytes == nil {
		t.Fatalf("StructToMap ptr_data should not be nil")
	}
	if string(ptrDataBytes) != `{"nested":true}` {
		t.Fatalf("StructToMap ptr_data content mismatch, got=%s", string(ptrDataBytes))
	}
}

func TestMapToStructWithRawMessage(t *testing.T) {
	// 测试 map 转结构体，包含 RawMessage
	var target testRawMessage
	err := MapToStruct(map[string]any{
		"id":   1,
		"name": "alice",
		"data": json.RawMessage(`{"key":"value"}`),
	}, &target, true)
	if err != nil {
		t.Fatalf("MapToStruct with RawMessage unexpected error: %v", err)
	}

	if target.Id != 1 {
		t.Fatalf("MapToStruct id mismatch, got=%d", target.Id)
	}
	if target.Name != "alice" {
		t.Fatalf("MapToStruct name mismatch, got=%s", target.Name)
	}
	if len(target.Data) == 0 {
		t.Fatalf("MapToStruct data should not be empty")
	}
	if string(target.Data) != `{"key":"value"}` {
		t.Fatalf("MapToStruct data content mismatch, got=%s", string(target.Data))
	}

	// 测试 byte 切片形式的 RawMessage
	var target2 testRawMessage
	err = MapToStruct(map[string]any{
		"id":   2,
		"name": "bob",
		"data": []byte(`{"num":123}`),
	}, &target2, true)
	if err != nil {
		t.Fatalf("MapToStruct with byte slice unexpected error: %v", err)
	}
	if string(target2.Data) != `{"num":123}` {
		t.Fatalf("MapToStruct byte slice data mismatch, got=%s", string(target2.Data))
	}

	// 测试字符串形式的 RawMessage
	var target3 testRawMessage
	err = MapToStruct(map[string]any{
		"id":   3,
		"name": "charlie",
		"data": `{"str":"test"}`,
	}, &target3, true)
	if err != nil {
		t.Fatalf("MapToStruct with string unexpected error: %v", err)
	}
	if string(target3.Data) != `{"str":"test"}` {
		t.Fatalf("MapToStruct string data mismatch, got=%s", string(target3.Data))
	}
}

func TestStructToMapsAndMapsToStructWithRawMessage(t *testing.T) {
	var rawMsg1 = json.RawMessage(`{"a":1}`)
	var rawMsg2 = json.RawMessage(`{"b":2}`)

	var users = []testRawMessage{
		{
			Id:   1,
			Name: "user1",
			Data: rawMsg1,
		},
		{
			Id:   2,
			Name: "user2",
			Data: rawMsg2,
		},
	}

	// 测试结构体切片转 map 切片
	maps, err := StructToMaps(users, false)
	if err != nil {
		t.Fatalf("StructToMaps with RawMessage unexpected error: %v", err)
	}
	if len(maps) != 2 {
		t.Fatalf("StructToMaps length mismatch, got=%d", len(maps))
	}

	// 验证第一个元素的 RawMessage
	data1, ok := maps[0]["data"].(json.RawMessage)
	if !ok {
		t.Fatalf("StructToMaps data type mismatch, got=%T", maps[0]["data"])
	}
	if string(data1) != `{"a":1}` {
		t.Fatalf("StructToMaps data content mismatch, got=%s", string(data1))
	}

	// 测试 map 切片转结构体切片
	var target []testRawMessage
	err = MapsToStruct([]map[string]any{
		{
			"id":   10,
			"name": "test1",
			"data": json.RawMessage(`{"x":100}`),
		},
		{
			"id":   20,
			"name": "test2",
			"data": json.RawMessage(`{"y":200}`),
		},
	}, &target, true)
	if err != nil {
		t.Fatalf("MapsToStruct with RawMessage unexpected error: %v", err)
	}
	if len(target) != 2 {
		t.Fatalf("MapsToStruct length mismatch, got=%d", len(target))
	}
	if target[0].Id != 10 || target[1].Id != 20 {
		t.Fatalf("MapsToStruct id mismatch, got=%v", []int{target[0].Id, target[1].Id})
	}
	if string(target[0].Data) != `{"x":100}` {
		t.Fatalf("MapsToStruct data[0] mismatch, got=%s", string(target[0].Data))
	}
	if string(target[1].Data) != `{"y":200}` {
		t.Fatalf("MapsToStruct data[1] mismatch, got=%s", string(target[1].Data))
	}
}
