package gdbtmp

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func initLog() {
}

var traceFunc = func(ctx context.Context) string {
	return ""
}

type TData struct {
	Id          int64     `json:"id"`                    // ID
	Ts          int64     `json:"ts"`                    // TS
	Uid         string    `json:"uid"`                   // uid
	Ip          string    `json:"ip"`                    // IP
	HostIp      string    `json:"host_ip"`               // 主机IP
	Ipv6        string    `json:"ipv6"`                  // IPv6
	HostIpv6    string    `json:"host_ipv6" gdbtmp:"-"`  // 主机IPv6
	LocalIp     string    `json:"local_ip" gdbtmp:"aaa"` // 源地址
	Int8        int8      `json:"int8"`
	Int16       int16     `json:"int16"`
	Int32       int32     `json:"int32"`
	Int64       int64     `json:"int64"`
	Uint8       uint8     `json:"uint8"`
	Uint16      uint16    `json:"uint16"`
	Uint32      uint32    `json:"uint32"`
	Uint64      uint64    `json:"uint64"`
	Float32     float32   `json:"float32"`
	Float64     float64   `json:"float64"`
	Bool        bool      `json:"bool"`
	CreateTime1 time.Time `json:"create_time1" gdbtmp:"type:Unix"`
	CreateTime2 time.Time `json:"create_time2" gdbtmp:"create_time;type:UnixMilli"`
	CreateTime3 time.Time `json:"create_time3" gdbtmp:"type:2006-01-02 15:04:05"`
}

func TestDb_MapToStruct(t *testing.T) {
	initLog()
	var err error
	var md = map[string]any{
		"aaa":          "DDDDDDDDD",
		"host_ip":      "172.16.12.12",
		"id":           1,
		"ip":           "192.168.101.10",
		"ipv6":         "::0",
		"host_ipv6":    "sss",
		"ts":           1,
		"uid":          "11111",
		"int8":         "1",
		"int16":        "2",
		"int32":        "3",
		"int64":        "4",
		"uint8":        "5",
		"uint16":       "6",
		"uint32":       "7",
		"uint64":       "8",
		"float32":      "9",
		"float64":      "10",
		"bool":         "11",
		"create_time1": time.Now().Unix(),
		"create_time2": time.Now().UnixMilli(),
		"create_time3": time.Now().Format(time.DateTime),
	}
	var td *TData
	err = MapToStruct(md, &td)
	fmt.Println(err)
	Dump(td)
}

func TestDb_MapsToStruct(t *testing.T) {
	initLog()
	var err error
	var md = []map[string]any{
		{
			"aaa":          "DDDDDDDDD",
			"host_ip":      "172.16.12.12",
			"id":           1,
			"ip":           "192.168.101.10",
			"ipv6":         "::0",
			"host_ipv6":    "sss",
			"ts":           1,
			"uid":          "11111",
			"int8":         "1",
			"int16":        "2",
			"int32":        "3",
			"int64":        "4",
			"uint8":        "5",
			"uint16":       "6",
			"uint32":       "7",
			"uint64":       "8",
			"float32":      "9",
			"float64":      "10",
			"bool":         "11",
			"create_time1": time.Now().Unix(),
			"create_time2": time.Now().UnixMilli(),
			"create_time3": time.Now().Format(time.DateTime),
		},
		{
			"aaa":          "2222DDDDDDDDD",
			"host_ip":      "222172.16.12.12",
			"id":           221,
			"ip":           "22192.168.101.10",
			"ipv6":         "22::0",
			"host_ipv6":    "22sss",
			"ts":           221,
			"uid":          "2211111",
			"int8":         "21",
			"int16":        "22",
			"int32":        "23",
			"int64":        "24",
			"uint8":        "25",
			"uint16":       "26",
			"uint32":       "27",
			"uint64":       "28",
			"float32":      "29",
			"float64":      "210",
			"bool":         "211",
			"create_time1": time.Now().Unix(),
			"create_time2": time.Now().UnixMilli(),
			"create_time3": time.Now().Format(time.DateTime),
		},
	}
	var td []*TData
	err = MapsToStruct(md, &td)
	fmt.Println(err)
	Dump(td)
}

func TestDb_StructToMap(t *testing.T) {
	initLog()
	var err error
	var md = TData{
		Id:          1,
		Ts:          2,
		Uid:         "3",
		Ip:          "4",
		HostIp:      "5",
		Ipv6:        "6",
		HostIpv6:    "7",
		LocalIp:     "8",
		Int8:        9,
		Int16:       10,
		Int32:       11,
		Int64:       12,
		Uint8:       13,
		Uint16:      14,
		Uint32:      15,
		Uint64:      16,
		Float32:     17,
		Float64:     18,
		Bool:        true,
		CreateTime1: time.Now(),
		CreateTime2: time.Now(),
		CreateTime3: time.Now(),
	}
	m, err := StructToMap(md)
	fmt.Println(err)
	Dump(m)
}

func TestDb_StructToMaps(t *testing.T) {
	initLog()
	var err error
	var md = []*TData{
		{
			Id:          1,
			Ts:          2,
			Uid:         "3",
			Ip:          "4",
			HostIp:      "5",
			Ipv6:        "6",
			HostIpv6:    "7",
			LocalIp:     "8",
			Int8:        9,
			Int16:       10,
			Int32:       11,
			Int64:       12,
			Uint8:       13,
			Uint16:      14,
			Uint32:      15,
			Uint64:      16,
			Float32:     17,
			Float64:     18,
			Bool:        true,
			CreateTime1: time.Now(),
			CreateTime2: time.Now(),
			CreateTime3: time.Now(),
		},
		{
			Id:          21,
			Ts:          22,
			Uid:         "23",
			Ip:          "24",
			HostIp:      "25",
			Ipv6:        "26",
			HostIpv6:    "27",
			LocalIp:     "28",
			Int8:        29,
			Int16:       210,
			Int32:       211,
			Int64:       212,
			Uint8:       213,
			Uint16:      214,
			Uint32:      215,
			Uint64:      216,
			Float32:     217,
			Float64:     218,
			Bool:        true,
			CreateTime1: time.Now(),
			CreateTime2: time.Now(),
			CreateTime3: time.Now(),
		},
	}
	m, err := StructToMaps(md)
	fmt.Println(err)
	Dump(m)
}

// 测试基础结构体
type User struct {
	ID     int     `gdbtmp:"id"`
	Name   string  `gdbtmp:"name"`
	Age    int     `gdbtmp:"age"`
	Score  float64 `gdbtmp:"score"`
	Active bool    `gdbtmp:"active"`
}

// 测试嵌套结构体
type Profile struct {
	Address string `gdbtmp:"address"`
	Phone   string `gdbtmp:"phone"`
}

type UserWithProfile struct {
	User
	Profile
	Level int `gdbtmp:"level"`
}

// 测试正常转换场景
func TestSlicesToStruct_Basic(t *testing.T) {
	// 输入数据：首行为字段名，后续为值
	data := [][]string{
		{"id", "name", "age", "score", "active"},
		{"1", "Alice", "20", "95.5", "true"},
		{"2", "Bob", "30", "88.0", "false"},
	}

	var users []*User
	err := SlicesToStruct(data, &users)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	// 验证结果长度
	if len(users) != 2 {
		t.Errorf("期望2条数据，实际%d条", len(users))
	}

	// 验证第一条数据
	expected1 := &User{ID: 1, Name: "Alice", Age: 20, Score: 95.5, Active: true}
	if !reflect.DeepEqual(users[0], expected1) {
		t.Errorf("第一条数据不匹配，期望%+v，实际%+v", expected1, users[0])
	}

	// 验证第二条数据
	expected2 := &User{ID: 2, Name: "Bob", Age: 30, Score: 88.0, Active: false}
	if !reflect.DeepEqual(users[1], expected2) {
		t.Errorf("第二条数据不匹配，期望%+v，实际%+v", expected2, users[1])
	}
}

// 测试嵌套结构体转换
func TestSlicesToStruct_Nested(t *testing.T) {
	data := [][]string{
		{"id", "name", "address", "phone", "level"},
		{"100", "Charlie", "Beijing", "13800138000", "3"},
	}

	var users []*UserWithProfile
	err := SlicesToStruct(data, &users)
	if err != nil {
		t.Fatalf("嵌套结构体转换失败: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("期望1条数据，实际%d条", len(users))
	}

	expected := &UserWithProfile{
		User:    User{ID: 100, Name: "Charlie"}, // Age/Score/Active未赋值（输入无对应字段）
		Profile: Profile{Address: "Beijing", Phone: "13800138000"},
		Level:   3,
	}
	if !reflect.DeepEqual(users[0], expected) {
		t.Errorf("嵌套数据不匹配，期望%+v，实际%+v", expected, users[0])
	}
}

// 测试输入数据长度不匹配（字段名数量与值数量不一致）
func TestSlicesToStruct_MismatchedLength(t *testing.T) {
	data := [][]string{
		{"id", "name"},       // 2个字段
		{"1", "Alice", "20"}, // 3个值（不匹配）
	}

	var users []*User
	err := SlicesToStruct(data, &users)
	if err == nil {
		t.Error("期望出现长度不匹配错误，但未报错")
	} else if err.Error() != "值数量不匹配，期望 2 个，实际 3 个" {
		t.Errorf("错误信息不正确，实际: %v", err)
	}
}

// 测试目标类型错误（非[]*结构体指针的指针）
func TestSlicesToStruct_InvalidDestType(t *testing.T) {
	data := [][]string{{"id"}, {"1"}}

	// 错误类型1：非指针
	var users []*User
	err := SlicesToStruct(data, users)
	if err == nil || err.Error() != "接收结果类型错误,应该为[]*结构体指针" {
		t.Errorf("错误类型1验证失败，实际错误: %v", err)
	}

	// 错误类型2：指针但非切片
	var user *User
	err = SlicesToStruct(data, &user)
	if err == nil || err.Error() != "接收结果类型错误,应该为[]*结构体指针" {
		t.Errorf("错误类型2验证失败，实际错误: %v", err)
	}

	// 错误类型3：切片但非指针元素
	var userSlice []User // 不是[]*User
	err = SlicesToStruct(data, &userSlice)
	if err == nil || err.Error() != "接收结果类型错误,应该为[]*结构体指针" {
		t.Errorf("错误类型3验证失败，实际错误: %v", err)
	}
}

// 测试忽略结构体中不存在的字段
func TestSlicesToStruct_IgnoreUnknownField(t *testing.T) {
	data := [][]string{
		{"id", "name", "unknown_field"}, // 包含结构体中不存在的字段
		{"3", "Dave", "ignore_me"},
	}

	var users []*User
	err := SlicesToStruct(data, &users)
	if err != nil {
		t.Fatalf("不应因未知字段报错: %v", err)
	}

	// 验证已知字段正常赋值，未知字段被忽略
	expected := &User{ID: 3, Name: "Dave"}
	if !reflect.DeepEqual(users[0], expected) {
		t.Errorf("未知字段处理错误，期望%+v，实际%+v", expected, users[0])
	}
}

// 测试类型转换失败（如字符串无法转为int）
func TestSlicesToStruct_InvalidTypeConversion(t *testing.T) {
	data := [][]string{
		{"id", "name"},
		{"not_a_number", "Eve"}, // id字段值无法转为int
	}

	var users []*User
	err := SlicesToStruct(data, &users)
	if err == nil {
		t.Error("期望类型转换错误，但未报错")
	}
	fmt.Println(err)
}

// 测试空输入数据
func TestSlicesToStruct_EmptyInput(t *testing.T) {
	var data [][]string // 空切片
	var users []*User
	err := SlicesToStruct(data, &users)
	if err == nil || err.Error() != "输入切片不能为空" {
		t.Errorf("空输入验证失败，实际错误: %v", err)
	}
}
