package gdb

import "testing"

type quickUser struct {
	Id   int    `json:"id" gdb:"id"`
	Name string `json:"name" gdb:"name"`
}

func (quickUser) TableName() string {
	return "quick_users"
}

func TestDbQuickFuncWhereCount(t *testing.T) {
	var run = func(name string, fn func(d *Db)) {
		var d = &Db{gs: NewSql("quick_users")}
		fn(d)
		if d.err != nil {
			t.Fatalf("%s 执行失败: %v", name, d.err)
		}
		if d.GetWhereLen() != 1 {
			t.Fatalf("%s where 数量错误: %d", name, d.GetWhereLen())
		}
	}

	run("Eq", func(d *Db) { d.Eq("id", 1) })
	run("NotEq", func(d *Db) { d.NotEq("status", 0) })
	run("Like", func(d *Db) { d.Like("name", "tom") })
	run("LikeRaw", func(d *Db) { d.LikeRaw("email", "%@example.com") })
	run("In", func(d *Db) { d.In("id", []int{1, 2}) })
	run("Lt", func(d *Db) { d.Lt("age", 30) })
	run("Le", func(d *Db) { d.Le("age", 40) })
	run("Gt", func(d *Db) { d.Gt("score", 80) })
	run("Ge", func(d *Db) { d.Ge("score", 90) })
	run("Between", func(d *Db) { d.Between("ts", 1, 2) })
	run("IsNotNull", func(d *Db) { d.IsNotNull("deleted_at") })
}

func TestDbTQuickFuncWhereCount(t *testing.T) {
	var run = func(name string, fn func(d *DbT[quickUser])) {
		var d = &DbT[quickUser]{Db: &Db{gs: NewSql("quick_users")}}
		fn(d)
		if d.Db.err != nil {
			t.Fatalf("%s 执行失败: %v", name, d.Db.err)
		}
		if d.Db.GetWhereLen() != 1 {
			t.Fatalf("%s where 数量错误: %d", name, d.Db.GetWhereLen())
		}
	}

	run("Eq", func(d *DbT[quickUser]) { d.Eq("id", 1) })
	run("NotEq", func(d *DbT[quickUser]) { d.NotEq("status", 0) })
	run("Like", func(d *DbT[quickUser]) { d.Like("name", "tom") })
	run("LikeRaw", func(d *DbT[quickUser]) { d.LikeRaw("email", "%@example.com") })
	run("In", func(d *DbT[quickUser]) { d.In("id", []int{1, 2}) })
	run("Lt", func(d *DbT[quickUser]) { d.Lt("age", 30) })
	run("Le", func(d *DbT[quickUser]) { d.Le("age", 40) })
	run("Gt", func(d *DbT[quickUser]) { d.Gt("score", 80) })
	run("Ge", func(d *DbT[quickUser]) { d.Ge("score", 90) })
	run("Between", func(d *DbT[quickUser]) { d.Between("ts", 1, 2) })
	run("IsNotNull", func(d *DbT[quickUser]) { d.IsNotNull("deleted_at") })
}
