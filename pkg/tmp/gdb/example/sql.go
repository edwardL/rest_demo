package example

import (
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/gdb"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
)

// 连贯操作函数
func sqlCoherent() {
	// 初始化 "table"选填 支持模型结构体
	gdb.NewSql("table")
	var res *gdb.Result
	_ = res
	var err error
	_ = err
	// 连贯操作函数
	res, err = gdb.NewSql().
		// 设置操作表
		Table("table").
		// 使用子查询
		Table("(?)", gdb.NewSql("t")).
		// 自定义JOIN类型连表
		Join("LEFT JOIN t2", "t.id = t2.id").
		// 无修饰 JOIN
		Join("table t", "t.id = t2.id").
		// INNER JOIN
		InnerJoin("table t", "t.id = t2.id").
		// LEFT JOIN
		LeftJoin("table t", "t.id = t2.id").
		// RIGHT JOIN
		RightJoin("table t", "t.id = t2.id").
		// AND 条件
		Where("id = ?", 1).
		// OR 条件
		WhereOr("id = ?", 1).
		// 重置查询条件
		WhereReset().
		// 条件组 AND (id = ? OR id = ?) 带括号包裹
		WhereGroup(func(gs *gdb.Sql) (err error) {
			gs.Where("id = ?", 1).WhereOr("id = ?", 1)
			return nil
		}).
		// 条件组 OR (id = ? OR id = ?) 带括号包裹
		WhereGroupOr(func(gs *gdb.Sql) (err error) {
			gs.Where("id = ?", 1).WhereOr("id = ?", 1)
			return nil
		}).
		// 查询字段 带参数
		Fields([]string{"id,ts", "name", "uid", "? AS age"}, 1).
		// 查询字段 带参数
		Fields("id,ts,name,uid,? AS age", 2).
		// 解析结构体字段 通过 tag `gdb:"-"` 忽略字段
		FieldsByData(gdb.BaseMode{}, nil).
		// 解析结构体字段 带别名
		FieldsByDataAlias(gdb.BaseMode{}, map[string]string{}, "t").
		// f[string | []string] 更新 新增忽略的字段
		OmitFields("id,ts").
		// 更新 新增字段名替换
		ReplaceFields(map[string]string{"id": "id2", "ts": "ts2"}).
		// 限制条数
		Limit("?,?", 1, 2).
		// 限制条数
		Limit("?", 2).
		// 偏移量
		Offset(1).
		// 分页
		Page(1, 2).
		// 重置分页 实际是重置Limit 和 Offset
		PageReset().
		// 分组 GROUP BY id,ts
		Group("id,ts").
		// 分组 GROUP BY id,ts
		Group([]string{"id", "ts"}).
		// 排序 ORDER BY id DESC
		Order("id DESC").
		// 排序 ORDER BY id DESC,ts ASC
		Order([]string{"id DESC", "ts ASC"}).
		// 排序 设置只能使用id排序
		OrderByFilter("id DESC,ts ASC", map[string]string{"id": "新名字{a.id} 可不设置"}).
		// 如果存在不允许排序字段时 报错
		OrderByFilter("id DESC,ts ASC", nil, false).
		// 使用查询视图构建搜索条件 设置允许的查询字段
		WhereSearchView(types.DataViewSearchInfo{}, make([]types.SearchItem, 0), map[string]string{"id": "新名字{a.id} 可不设置"}).
		// 根据前端参数构建搜索条件 设置允许的查询字段
		WhereSearch(make([]types.SearchItem, 0), map[string]string{"id": "新名字{a.id} 可不设置"}).
		// 联合查询 合并两个查询并去重
		Union(gdb.NewSql()).
		// 联合查询  合并两个查询不去重
		UnionAll(gdb.NewSql()).
		// 使用原生SQL查询 调用这个 其他的都不生效
		//Raw("SELECT * FROM table").
		Select()
	fmt.Println(err)
	fmt.Println(res.CompSql())
}

// 结束操作函数
func sqlEdn() {
	var err error
	_ = err
	var total int64
	_ = total
	var res *gdb.Result
	// 初始化 "table"选填 支持模型结构体
	gdb.NewSql("table")
	// 查询多条数据到结构体或map
	res, err = gdb.NewSql().Select()
	_ = res.Sql       // sql 带 ?
	_ = res.Args      // 对应的参数
	_ = res.CompSql() // 完整的sql 并非执行的sql 只是方便调试使用
	// 查询总数
	res, err = gdb.NewSql().Count()
	// 更新某个字段
	res, err = gdb.NewSql().Update("name", "new name")
	// 更新多个字段 支持结构体
	res, err = gdb.NewSql().Updates(map[string]any{
		"name": "new name",
		"age":  18,
	})
	// 更新多个字段 支持结构体
	res, err = gdb.NewSql().Updates(map[string]any{
		"name": "new name",
		"age":  18,
	})
	// 更新多个字段 忽略冲突 支持结构体
	res, err = gdb.NewSql().UpdateIgnore(map[string]any{
		"name": "new name",
		"age":  18,
	})
	// 存在更新，否则新增 data = map[string]any 或者 Struct
	res, err = gdb.NewSql().Save(map[string]any{
		"name": "new name",
		"age":  18,
	})
	// 删除数据
	res, err = gdb.NewSql().Delete()
	// 连表删除
	res, err = gdb.NewSql().Table("table t1").
		LeftJoin("tb t2", "t1.id = t2.id").
		Where("t2.id IS NULL").
		Delete("t1")
	// 新增数据
	res, err = gdb.NewSql().Create(map[string]any{
		"name": "new name",
		"age":  18,
	})
	// 从查询新增数据 INSERT INTO table (name,age) SELECT name,age FROM table
	res, err = gdb.NewSql().Fields("name,age").Create(gdb.New("table").Field("name,age"))
}
