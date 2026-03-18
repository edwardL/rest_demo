package performance_test

import (
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb"
	"testing"
)

// TestStructSql 用于SQL测试的结构体
type TestStructSql struct {
	ID     int    `json:"id" gdb:"id"`
	Name   string `json:"name" gdb:"name"`
	Age    int    `json:"age" gdb:"age"`
	Email  string `json:"email" gdb:"email"`
	Active bool   `json:"active" gdb:"active"`
}

// TableName 为TestStructSql定义表名
func (t TestStructSql) TableName() string {
	return "users"
}

// BenchmarkSql_Select 测试Select方法性能
func BenchmarkSql_Select(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SelectWithFields 测试Select方法带字段性能
func BenchmarkSql_SelectWithFields(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Fields("id, name, age").
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SelectWithWhere 测试Select方法带Where条件性能
func BenchmarkSql_SelectWithWhere(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("id = ?", i).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SelectWithComplexWhere 测试Select方法带复杂Where条件性能
func BenchmarkSql_SelectWithComplexWhere(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("id = ? AND name = ? AND age > ?", i, "test_user", 18).
			WhereOr("email LIKE ?", "%@example.com").
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SelectWithJoin 测试Select方法带Join性能
func BenchmarkSql_SelectWithJoin(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users u").
			LeftJoin("profiles p", "u.id = p.user_id").
			Where("u.id = ?", i).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SelectWithPagination 测试Select方法带分页性能
func BenchmarkSql_SelectWithPagination(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("age > ?", 18).
			Page(1, 20).
			Order("id DESC").
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Update 测试Update方法性能
func BenchmarkSql_Update(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("id = ?", i).
			Update("name", "updated_name")
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Updates 测试Updates方法性能
func BenchmarkSql_Updates(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := map[string]any{
			"name":   "test_name",
			"age":    25,
			"email":  "test@example.com",
			"active": true,
		}
		sqlObj, err := gdb.NewSql("users").
			Where("id = ?", i).
			Updates(testData)
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Create 测试Create方法性能
func BenchmarkSql_Create(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := map[string]any{
			"name":   "test_name",
			"age":    25,
			"email":  "test@example.com",
			"active": true,
		}
		sqlObj, err := gdb.NewSql("users").
			Create(testData)
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_CreateInBatches 测试CreateInBatches方法性能
func BenchmarkSql_CreateInBatches(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testData := []map[string]any{
			{
				"name":   "test_name_1",
				"age":    25,
				"email":  "test1@example.com",
				"active": true,
			},
			{
				"name":   "test_name_2",
				"age":    30,
				"email":  "test2@example.com",
				"active": false,
			},
		}
		sqlObj, err := gdb.NewSql("users").
			CreateInBatches(testData)
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Delete 测试Delete方法性能
func BenchmarkSql_Delete(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("id = ?", i).
			Delete()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Raw 测试Raw方法性能
func BenchmarkSql_Raw(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj := gdb.NewSql().Raw("SELECT * FROM users WHERE id = ?", i)
		_ = sqlObj
	}
}

// BenchmarkSql_Table 测试Table方法性能
func BenchmarkSql_Table(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj := gdb.NewSql().Table("users")
		_ = sqlObj
	}
}

// BenchmarkSql_StructBasedTable 测试基于结构体的Table方法性能
func BenchmarkSql_StructBasedTable(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql(TestStructSql{}).Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_ComplexQuery 测试复杂查询性能
func BenchmarkSql_ComplexQuery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users u").
			LeftJoin("profiles p", "u.id = p.user_id").
			Fields("u.id, u.name, u.email, p.bio").
			Where("u.active = ?", true).
			Where("u.age BETWEEN ? AND ?", 18, 65).
			WhereOr("u.email LIKE ?", "%@company.com").
			Group("u.id").
			Order("u.id DESC").
			Page(1, 50).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_WhereInWithNestedSlice 测试Where中IN子句带嵌套切片的性能
func BenchmarkSql_WhereInWithNestedSlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("(id,ts) IN (?)", [][]any{{1, 2}, {3, 4}}).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_WhereInWithAnySlice 测试Where中IN子句带any切片的性能
func BenchmarkSql_WhereInWithAnySlice(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqlObj, err := gdb.NewSql("users").
			Where("id IN (?)", []any{1, 2, 3, 4}).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_Subquery 测试子查询性能
func BenchmarkSql_Subquery(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subQuery := gdb.NewSql("orders").Fields("user_id").Where("status = ?", "active")
		sqlObj, err := gdb.NewSql("users").
			Where("id IN (?)", subQuery).
			Select()
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}

// BenchmarkSql_SubqueryUpdate 测试子查询更新性能
func BenchmarkSql_SubqueryUpdate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subQuery := gdb.NewSql("users").Fields("name").Where("id = ?", i)
		sqlObj, err := gdb.NewSql("profiles").
			Where("user_id = ?", i).
			Update("user_name", subQuery)
		if err != nil {
			b.Fatal(err)
		}
		_ = sqlObj
	}
}
