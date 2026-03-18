package gdb

import (
	"encoding/json"
	"fmt"
	hhdb "nwgit.gzhhit.com/BD/hhitdb.git"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type configTestUser struct {
	Id     int64   `json:"id" gdb:"id"`
	Name   string  `json:"name" gdb:"name"`
	Age    int     `json:"age" gdb:"age"`
	Email  string  `json:"email" gdb:"email"`
	Score  float64 `json:"score" gdb:"score"`
	Active bool    `json:"active" gdb:"active"`
}

func (configTestUser) TableName() string {
	return "users_test"
}

var dbTestInitOnce sync.Once
var dbTestInitErr error

func loadDbConfigForTest() (map[string]string, error) {
	var dbConf = map[string]string{
		"db_host":            "",
		"db_port":            "",
		"db_user":            "",
		"db_pass":            "",
		"db_name":            "",
		"db_charset":         "utf8mb4",
		"db_type":            "mysql",
		"need_mode_onebyone": "1",
	}

	var confPaths = []string{
		filepath.Join("..", "..", "config.json"),
		filepath.Join("config.json"),
		filepath.Join("..", "test_conf.json"),
	}

	var confPath string
	for _, confPath = range confPaths {
		var content []byte
		var readErr error
		content, readErr = os.ReadFile(confPath)
		if readErr != nil {
			continue
		}
		var unmarshalErr = json.Unmarshal(content, &dbConf)
		if unmarshalErr != nil {
			return nil, fmt.Errorf("解析配置文件失败 %s: %w", confPath, unmarshalErr)
		}
		return dbConf, nil
	}

	return nil, fmt.Errorf("未找到数据库配置文件，已尝试路径: %v", confPaths)
}

func initDbByRootConfig(t *testing.T) {
	t.Helper()
	dbTestInitOnce.Do(func() {
		var dbConf map[string]string
		dbConf, dbTestInitErr = loadDbConfigForTest()
		if dbTestInitErr != nil {
			return
		}

		dbTestInitErr = hhdb.InitDB(dbConf["db_type"], &dbConf)
		if dbTestInitErr != nil {
			return
		}

		Init(WithWriteLog(true), WithWriteHhDbLog(false))

		var sqlDb = hhdb.GetDBOP().GetDBObj()
		if sqlDb == nil {
			dbTestInitErr = fmt.Errorf("数据库对象为空")
			return
		}

		var createSql = `CREATE TABLE IF NOT EXISTS users_test (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(64) NOT NULL,
	age INT NOT NULL,
	email VARCHAR(128) NOT NULL,
	score DECIMAL(10,2) NOT NULL DEFAULT 0,
	active TINYINT(1) NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

		var _, execErr = sqlDb.Exec(createSql)
		if execErr != nil {
			dbTestInitErr = fmt.Errorf("创建测试表失败: %w", execErr)
			return
		}

		var _, checkErr = New().Raw("SELECT id,name FROM users_test LIMIT 1").Map()
		if checkErr != nil {
			dbTestInitErr = fmt.Errorf("测试表 users_test 不可用: %w", checkErr)
			return
		}
	})

	if dbTestInitErr != nil {
		t.Skipf("数据库初始化失败，跳过集成测试: %v", dbTestInitErr)
	}
}

func resetUsersTestTable(t *testing.T) {
	t.Helper()
	var sqlDb = hhdb.GetDBOP().GetDBObj()
	if sqlDb == nil {
		t.Skipf("数据库对象为空，跳过集成测试")
	}
	var _, err = sqlDb.Exec("TRUNCATE TABLE users_test")
	if err != nil {
		t.Skipf("清理测试表失败，跳过集成测试: %v", err)
	}
}

func TestDb_ConfigConnectionAndCrud(t *testing.T) {
	initDbByRootConfig(t)
	resetUsersTestTable(t)

	var user = configTestUser{
		Name:   "alice",
		Age:    20,
		Email:  "alice@example.com",
		Score:  88.5,
		Active: true,
	}

	var createRes = New(&configTestUser{}).Create(&user)
	if createRes.Error != nil {
		t.Fatalf("Create 失败: %v", createRes.Error)
	}
	if createRes.GetLastInsertId() <= 0 {
		t.Fatalf("Create 返回插入ID异常: %d", createRes.GetLastInsertId())
	}

	var users []configTestUser
	var selectRes = New(&configTestUser{}).Where("age >= ?", 18).Order("id ASC").Select(&users)
	if selectRes.Error != nil {
		t.Fatalf("Select 失败: %v", selectRes.Error)
	}
	if len(users) != 1 {
		t.Fatalf("Select 结果数量错误，期望1，实际%d", len(users))
	}

	var one configTestUser
	var oneRes = New(&configTestUser{}).Where("id = ?", createRes.GetLastInsertId()).One(&one)
	if oneRes.Error != nil {
		t.Fatalf("One 失败: %v", oneRes.Error)
	}
	if one.Name != "alice" {
		t.Fatalf("One 返回数据错误，期望alice，实际%s", one.Name)
	}

	var updateRes = New(&configTestUser{}).Where("id = ?", createRes.GetLastInsertId()).Update("name", "alice_update")
	if updateRes.Error != nil {
		t.Fatalf("Update 失败: %v", updateRes.Error)
	}

	var check configTestUser
	var checkRes = New(&configTestUser{}).Where("id = ?", createRes.GetLastInsertId()).One(&check)
	if checkRes.Error != nil {
		t.Fatalf("更新后查询失败: %v", checkRes.Error)
	}
	if check.Name != "alice_update" {
		t.Fatalf("更新结果错误，期望alice_update，实际%s", check.Name)
	}

	var count int64
	var countRes = New(&configTestUser{}).Count(&count)
	if countRes.Error != nil {
		t.Fatalf("Count 失败: %v", countRes.Error)
	}
	if count != 1 {
		t.Fatalf("Count 结果错误，期望1，实际%d", count)
	}

	var exists bool
	var existsErr error
	exists, existsErr = New(&configTestUser{}).Where("id = ?", createRes.GetLastInsertId()).Exists()
	if existsErr != nil {
		t.Fatalf("Exists 失败: %v", existsErr)
	}
	if !exists {
		t.Fatalf("Exists 结果错误，期望true")
	}

	var deleteRes = New(&configTestUser{}).Where("id = ?", createRes.GetLastInsertId()).Delete()
	if deleteRes.Error != nil {
		t.Fatalf("Delete 失败: %v", deleteRes.Error)
	}
}

func TestDb_ConfigTransactionAndRawQuery(t *testing.T) {
	initDbByRootConfig(t)
	resetUsersTestTable(t)

	var txErr = New().Transaction(func(dbTx *DbTx) error {
		var txUser = configTestUser{
			Name:   "tx_user",
			Age:    30,
			Email:  "tx_user@example.com",
			Score:  92,
			Active: true,
		}
		var createRes = dbTx.Table(&configTestUser{}).Create(&txUser)
		if createRes.Error != nil {
			return createRes.Error
		}
		return nil
	})
	if txErr != nil {
		t.Fatalf("Transaction 失败: %v", txErr)
	}

	var m map[string]any
	var mapErr error
	m, mapErr = New().Raw("SELECT id,name FROM users_test WHERE name = ?", "tx_user").Map()
	if mapErr != nil {
		t.Fatalf("Raw + Map 查询失败: %v", mapErr)
	}
	if m["name"] != "tx_user" {
		t.Fatalf("Raw + Map 结果错误，期望tx_user，实际%v", m["name"])
	}

	var rows []map[string]any
	rows, mapErr = New().Raw("SELECT id,name FROM users_test").Maps()
	if mapErr != nil {
		t.Fatalf("Raw + Maps 查询失败: %v", mapErr)
	}
	if len(rows) != 1 {
		t.Fatalf("Raw + Maps 结果数量错误，期望1，实际%d", len(rows))
	}
}
