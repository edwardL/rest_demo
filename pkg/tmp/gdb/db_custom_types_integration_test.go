package gdb

import (
	"fmt"
	hhdb "nwgit.gzhhit.com/BD/hhitdb.git"
	"testing"
)

type customProfile struct {
	Age  int    `json:"age"`
	City string `json:"city"`
}

type customTypeRow struct {
	Id      int64                   `json:"id" gdb:"id"`
	Name    string                  `json:"name" gdb:"name"`
	Profile JSONType[customProfile] `json:"profile" gdb:"profile"`
	Levels  ENUMType[int]           `json:"levels" gdb:"levels"`
	Labels  ENUMType[string]        `json:"labels" gdb:"labels"`
}

func (customTypeRow) TableName() string {
	return "custom_types_test"
}

func ensureCustomTypesTable(t *testing.T) {
	t.Helper()
	var sqlDb = hhdb.GetDBOP().GetDBObj()
	if sqlDb == nil {
		t.Skipf("数据库对象为空，跳过自定义类型集成测试")
	}

	var createSql = `CREATE TABLE IF NOT EXISTS custom_types_test (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(64) NOT NULL,
	profile JSON NULL,
	levels VARCHAR(128) NOT NULL DEFAULT '',
	labels VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

	var _, err = sqlDb.Exec(createSql)
	if err != nil {
		t.Skipf("创建测试表失败，跳过测试: %v", err)
	}

	_, err = sqlDb.Exec("TRUNCATE TABLE custom_types_test")
	if err != nil {
		t.Skipf("清空测试表失败，跳过测试: %v", err)
	}
}

func TestDb_CustomJSONAndENUM_CreateUpdateQuery(t *testing.T) {
	initDbByRootConfig(t)
	ensureCustomTypesTable(t)

	var row = customTypeRow{
		Name:    "alpha",
		Profile: JSONType[customProfile]{Data: customProfile{Age: 18, City: "gz"}},
		Levels:  ENUMType[int]{1, 2},
		Labels:  ENUMType[string]{"a", "b"},
	}

	var createRes = New(&customTypeRow{}).Create(&row)
	if createRes.Error != nil {
		t.Fatalf("Create 自定义类型失败: %v", createRes.Error)
	}
	if createRes.GetLastInsertId() <= 0 {
		t.Fatalf("Create 返回ID异常: %d", createRes.GetLastInsertId())
	}

	var one customTypeRow
	var oneRes = New(&customTypeRow{}).Where("id = ?", createRes.GetLastInsertId()).One(&one)
	if oneRes.Error != nil {
		t.Fatalf("One 查询失败: %v", oneRes.Error)
	}
	if one.Profile.Data.Age != 18 || one.Profile.Data.City != "gz" {
		t.Fatalf("JSONType 查询值异常: %#v", one.Profile.Data)
	}
	if len(one.Levels) != 2 || one.Levels[0] != 1 || one.Levels[1] != 2 {
		t.Fatalf("ENUMType[int] 查询值异常: %#v", one.Levels)
	}
	if len(one.Labels) != 2 || one.Labels[0] != "a" || one.Labels[1] != "b" {
		t.Fatalf("ENUMType[string] 查询值异常: %#v", one.Labels)
	}

	var updateRes = New(&customTypeRow{}).Where("id = ?", createRes.GetLastInsertId()).Updates(map[string]any{
		"profile": JSONType[customProfile]{Data: customProfile{Age: 20, City: "sz"}},
		"levels":  ENUMType[int]{3},
		"labels":  ENUMType[string]{"x", "y", "z"},
	})
	if updateRes.Error != nil {
		t.Fatalf("Updates 自定义类型失败: %v", updateRes.Error)
	}

	var check customTypeRow
	var checkRes = New(&customTypeRow{}).Where("id = ?", createRes.GetLastInsertId()).One(&check)
	if checkRes.Error != nil {
		t.Fatalf("更新后查询失败: %v", checkRes.Error)
	}
	if check.Profile.Data.Age != 20 || check.Profile.Data.City != "sz" {
		t.Fatalf("JSONType 更新后值异常: %#v", check.Profile.Data)
	}
	if len(check.Levels) != 1 || check.Levels[0] != 3 {
		t.Fatalf("ENUMType[int] 更新后值异常: %#v", check.Levels)
	}
	if len(check.Labels) != 3 || check.Labels[0] != "x" || check.Labels[2] != "z" {
		t.Fatalf("ENUMType[string] 更新后值异常: %#v", check.Labels)
	}

	var list []customTypeRow
	var selectRes = New(&customTypeRow{}).Order("id ASC").Select(&list)
	if selectRes.Error != nil {
		t.Fatalf("Select 查询失败: %v", selectRes.Error)
	}
	if len(list) != 1 {
		t.Fatalf("Select 数量异常: %d", len(list))
	}
	if fmt.Sprint(list[0].Levels) != fmt.Sprint(check.Levels) {
		t.Fatalf("Select ENUMType 值异常: %#v %#v", list[0].Levels, check.Levels)
	}
}
