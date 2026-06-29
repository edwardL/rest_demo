//go:build ignore

package example

import (
	"context"
	"errors"
	"fmt"
	"rest_demo/pkg/gdbtmp"
)

func dbInit() {
	var dbConf = map[string]string{
		"db_host":    "",
		"db_port":    "",
		"db_user":    "",
		"db_pass":    "",
		"db_name":    "",
		"db_charset": "utf8mb4",
		"db_type":    "mysql",
	}
	var err error
	_ = dbConf
	_ = err
	db := gdbtmp.New()
	_ = db
	if err != nil {
		fmt.Println(err)
	}
}

// 工具函数
func util() {
	dbInit()
	var db = gdbtmp.New()
	// 工具函数 GenWhere() 构建where条件
	var q, a = db.
		Where("id IN (?) AND name = ? AND area IN (?)", []int{1, 2, 3}, "make", []string{"bj", "sh"}).
		WhereOr("age > ?", 20).
		GenWhere()
	fmt.Println(q) // id IN (?, ?, ?) AND name = ? AND area IN (?, ?) OR age > ?
	fmt.Println(a) // [1 2 3 make bj sh 20]

	db = gdbtmp.New()
	// 工具函数 GetWhereLen() 当前的where条件长度
	var l = db.
		Where("id IN (?) AND name = ? AND area IN (?)", []int{1, 2, 3}, "make", []string{"bj", "sh"}).
		WhereOr("age > ?", 20).
		GetWhereLen()
	fmt.Println(l) // 2

	db = gdbtmp.New()
	// 工具函数 GetJoinLen() 当前的join条件长度
	l = db.
		LeftJoin("t", "t1.id = t2.id").
		GetJoinLen()
	fmt.Println(l) // 1

	// 工具函数 GetGSql() 返回当前查询的sql构建器
	var gs = gdbtmp.New().Table("tb t2").
		LeftJoin("t", "t1.id = t2.id").
		GetGSql()
	fmt.Println(gs.Select()) // &{SELECT * FROM tb t2 LEFT JOIN t ON t1.id = t2.id []} <nil>

	// 工具函数 Raw() 原生拼接
	gdbtmp.New("table").
		Where("id < ?", 0).
		Updates(map[string]any{
			"age":  gdbtmp.Raw("age+1"), // 适配自增等特殊场景
			"age2": "age2+1",            // 适配自增场景
		})
	// UPDATE table SET age2 = ?, age = age+1 WHERE id < 0
	// ['age2+1']
}

// 连贯操作函数
func dbCoherent() {
	// 初始化 "table"选填 支持模型结构体
	gdbtmp.New("table")
	// 带上下文初始化 "table"选填 支持模型结构体
	gdbtmp.NewCtx(context.Background(), "table")
	// 连贯操作函数
	gdbtmp.New().
		// 设置操作表
		Table("table").
		// 使用子查询
		Table("(?)", gdbtmp.New("t")).
		// 自定义JOIN类型连表
		Joins("LEFT JOIN t2", "t.id = t2.id").
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
		WhereGroup(func(gs *gdbtmp.Sql) (err error) {
			gs.Where("id = ?", 1).WhereOr("id = ?", 1)
			return nil
		}).
		// 条件组 OR (id = ? OR id = ?) 带括号包裹
		WhereGroupOr(func(gs *gdbtmp.Sql) (err error) {
			gs.Where("id = ?", 1).WhereOr("id = ?", 1)
			return nil
		}).
		// 查询字段
		Field("id,ts", "name", "uid").
		// 查询字段 带参数
		Fields([]string{"id,ts", "name", "uid", "? AS age"}, 1).
		// 查询字段 带参数
		Fields("id,ts,name,uid,? AS age", 2).
		// 解析结构体字段 通过 tag `gdbtmp:"-"` 忽略字段
		FieldsByData(gdbtmp.BaseMode{}, nil).
		// 解析结构体字段 带别名
		FieldsByDataAlias(gdbtmp.BaseMode{}, map[string]string{}, "t").
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
		Union(gdbtmp.New()).
		// 联合查询  合并两个查询不去重
		UnionAll(gdbtmp.New()).
		// 使用原生SQL查询 调用这个 其他的都不生效
		Raw("SELECT * FROM table").
		// 自定义字段类型转换 在查询map的时候转换类型返回给前端
		ConvFieldsType(func(key string, val any, m *gdbtmp.ReadOnlyMap) any {
			// m 为当前结果值
			switch key {
			case "id", "ts":
				intVal, _ := gdbtmp.ToInt(val) // id 和 ts 返回数组类型给前端 避免出现 id:"1" 的情况
				return intVal
			}
			// 保持原样
			return val
		}).
		// 数组上下文
		Ctx(context.Background()).
		// 当前操作是否打印日志
		WriteLog(true).
		// 当前操作是否打印框架HhDb日志
		WriteHhDbLog(true).
		// 设置当前操作的日志级别
		LogLevel(gdbtmp.DebugLogLevel).
		Select("这里不能是字符串")
}

// 结束操作函数
func dbEdn() {
	var err error
	_ = err
	var total int64
	_ = total
	var res map[string]any
	var resList []map[string]any
	// 初始化 "table"选填 支持模型结构体
	gdbtmp.New("table")
	// 带上下文初始化 "table"选填 支持模型结构体
	gdbtmp.NewCtx(context.Background(), "table")
	// 查询单条数据到结构体或map自动判断
	err = gdbtmp.New().Scan(&res).Error
	// 查询多条数据到结构体或map自动判断
	err = gdbtmp.New().Scan(&resList).Error
	// 查询单条数据到结构体或map
	err = gdbtmp.New().One(&res).Error
	// 查询多条数据到结构体或map
	err = gdbtmp.New().Select(&resList).Error
	// 查询总数
	err = gdbtmp.New().Count(&total).Error
	// 更新某个字段
	err = gdbtmp.New().Update("name", "new name").Error
	// 更新多个字段 支持结构体
	err = gdbtmp.New().Updates(map[string]any{
		"name": "new name",
		"age":  18,
	}).Error
	// 连表更新
	err = gdbtmp.New("table t1").
		LeftJoin("t t2", "t1.id = t2.id").
		Updates(map[string]gdbtmp.RawBody{
			"t1.name": "t2.name",
			"t1.age":  "t2.age",
		}).Error
	// 更新多个字段 忽略冲突 支持结构体
	err = gdbtmp.New().UpdateIgnore(map[string]any{
		"name": "new name",
		"age":  18,
	}).Error
	// 存在更新，否则新增 data = map[string]any 或者 Struct
	err = gdbtmp.New().Save(map[string]any{
		"name": "new name",
		"age":  18,
	}).Error
	// 存在更新，否则新增 批量  data = []map[string]any 或者 []Struct blockSize 批量大小 默认100
	err = gdbtmp.New().SaveInBatches([]map[string]any{
		{
			"name": "new name",
			"age":  18,
		},
	}, 50).Error
	// 删除数据
	err = gdbtmp.New().Delete().Error
	// 连表删除
	err = gdbtmp.New().Table("table t1").
		LeftJoin("tb t2", "t1.id = t2.id").
		Where("t2.id IS NULL").
		Delete("t1").Error
	// 新增数据
	err = gdbtmp.New().Create(map[string]any{
		"name": "new name",
		"age":  18,
	}).Error
	// 从查询新增数据 INSERT INTO table (name,age) SELECT name,age FROM table
	err = gdbtmp.New().Field("name,age").Create(gdbtmp.New("table").Field("name,age")).Error
	// 批量新增数据 data = []map[string]any 或者 []Struct blockSize 批量大小 默认100
	err = gdbtmp.New().CreateInBatches([]map[string]any{
		{
			"name": "new name",
			"age":  18,
		},
	}, 50).Error
	// 查询直接返回Map
	res, err = gdbtmp.New().Map()
	// 查询直接返回Map列表
	resList, err = gdbtmp.New().Maps()
	// 事务查询
	err = gdbtmp.New().Transaction(func(tx *gdbtmp.DbTx) error {
		err = gdbtmp.New("table").Select("").Error
		err = gdbtmp.New("table").Delete().Error
		return errors.New("回滚")
	})
}

// 流式查询
func stream() {
	var err = gdbtmp.New("tb").Where("id > 0").Stream(gdbtmp.StreamCallback(func(t map[string]any) (err error, next bool) {
		fmt.Println(t)
		return nil, true
	})).Error
	fmt.Println(err)
}
