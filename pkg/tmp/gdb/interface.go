package gdb

type NewWrapperFunc func(s *Sql) SqlWrapperFace

// SqlWrapperFace 构建sql操作对象
type SqlWrapperFace interface {
	// RawExec 原生更新
	RawExec(execSql string, args ...any) (r *Result, err error)
	// Select 获取查询语句
	Select() (r *Result, err error)
	// Count 获取统计语句 f 为COUNT(f[0])
	Count(f ...string) (r *Result, err error)
	// Exists 判断是否存在
	Exists() (r *Result, err error)
	// Update 更新某个字段
	Update(column string, val any) (r *Result, err error)
	// Updates 更新多个字段 data = map[string]any 或者 Struct
	Updates(data any, whereField ...[]string) (r *Result, err error)
	// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
	UpdateIgnore(data any) (r *Result, err error)
	// Save 插入并更新
	Save(data any) (r *Result, err error)
	// SaveInBatches 生成批量插入数据的sql语句
	// data = map[string]any 或者 Struct
	SaveInBatches(data any) (r *Result, err error)
	// Delete 删除 DELETE {t} From table
	Delete(t ...string) (r *Result, err error)
	// Create 创建 data = map[string]any 或者 Struct
	Create(data any) (r *Result, err error)
	// CreateInBatches 批量创建
	// data = []map[string]any 或者 []Struct
	CreateInBatches(data any) (r *Result, err error)
	// Replace 创建或替换 data = map[string]any 或者 Struct
	Replace(data any) (r *Result, err error)
	// ReplaceInBatches 批量创建或替换
	// data = []map[string]any 或者 []Struct
	ReplaceInBatches(data any) (r *Result, err error)
	// GenWhere 获取where条件
	GenWhere() (whereStr string, args []any)
}
