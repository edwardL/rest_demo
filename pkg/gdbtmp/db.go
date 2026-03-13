package gdbtmp

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

const (
	txKey        = "Tx " // 日志打印前缀
	defBatchSize = 100   // 默认分批大小
)

// ExecResult 更新操作结果
type ExecResult struct {
	AffectRecs   int64
	LastAffectId int64
}

// ConvVal 类型转换函数类型
// 传入 key val obj 返回当前新的val
// obj使用只读封装避免使用不当出现数据竞争
type ConvVal func(key string, val any, m *ReadOnlyMap) any

// Db 操作对象
type Db struct {
	db                   *sql.DB
	sqlTx                *Tx
	SqlResult            *Result
	log                  GdbLog
	gs                   *Sql
	ctx                  context.Context
	writeLog             bool // 是否打印日志
	writeHhDbLog         bool // 是否打印框架日志
	writeErrSql          bool // 是否打印错误sql
	writeCompSql         bool // 是否打印完整sql
	err                  error
	fieldTypeReplaceFunc ConvVal // 更新和新增字段类型替换
	//listResultsByList    hhdb.ReResultsByList
	//mapResultsByList     hhdb.ReResultsByMap01
	execResult ExecResult
}

// syncRes 同步结果
func (d *Db) syncRes(r *OrmResult) *OrmResult {
	if r == nil {
		r = new(OrmResult)
	}
	if r.Error == nil {
		r.Error = d.err
	}
	//r.ListResultsByList = d.listResultsByList
	r.ExecResult = d.execResult
	return r
}

// OrmResult 统一返回结果
type OrmResult struct {
	Error error
	//ListResultsByList hhdb.ReResultsByList // 不保留副本
	ExecResult ExecResult
}

// New 创建对象
func New(table ...any) *Db {
	return NewCtx(context.Background(), table...)
}

// NewCtx 创建带上下文对象
func NewCtx(ctx context.Context, table ...any) *Db {
	if ctx == nil {
		ctx = context.Background()
	}
	var db = &Db{
		gs:           NewSql(table...),
		log:          defLog.Clone(),
		writeHhDbLog: conf.WriteHhDbLog,
		writeLog:     conf.WriteLog,
		writeErrSql:  conf.WriteErrSql,
		writeCompSql: conf.WriteCompSql,
	}
	if db.db == nil {
		db.err = errors.New("未初始化数据库")
	}
	db.Ctx(ctx)
	return db
}

// Scan 解析到结构体 如果是数组 数组里必须是指针 []*Struct 兼容单结构体和数组
func (d *Db) Scan(dest any) (r *OrmResult) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	if IsArrayOrSlice(dest) {
		return d.Select(dest)
	} else {
		return d.One(dest)
	}
}

// ScanAndCount 解析到结构体  数组里必须是指针 []*Struct
func (d *Db) ScanAndCount(dest any, count *int64) (r *OrmResult) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	return d.SelectAndCount(dest, count)
}

// Select 解析到数组结构体 数组里必须是指针 []*Struct
func (d *Db) Select(dest any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var ms []map[string]any
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	if destPoint, ok := dest.(*[]map[string]any); ok {
		ms, d.err = d.Maps()
		if d.err != nil {
			return d.syncRes(r)
		}
		*destPoint = ms
		return d.syncRes(r)
	}
	var sls [][]string
	sls, d.err = d.queryToSlices()
	if d.err != nil {
		return d.syncRes(r)
	}
	d.err = SlicesToStruct(sls, dest)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// SelectAndCount 解析到数组结构体并统计数量 数组里必须是指针 []*Struct
func (d *Db) SelectAndCount(dest any, count *int64) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var ms []map[string]any
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var ctx = setLogCallDepthCtx(d.ctx, getLogCallDepthCtx(d.ctx))
	var totalRes *OrmResult
	var totalQuery = d.Clone().PageReset()
	if totalQuery.gs.group.Query != "" {
		totalQuery = New().Table("(?) t", totalQuery)
	}
	totalRes = totalQuery.Ctx(ctx).Count(count)
	if totalRes.Error != nil {
		return totalRes
	}
	if *count == 0 { // 无数据不再查询
		return d.syncRes(r)
	}
	if destPoint, ok := dest.(*[]map[string]any); ok {
		ms, d.err = d.Maps()
		if d.err != nil {
			return d.syncRes(r)
		}
		*destPoint = ms
		return d.syncRes(r)
	}
	var sls [][]string
	sls, d.err = d.queryToSlices()
	if d.err != nil {
		return d.syncRes(r)
	}
	d.err = SlicesToStruct(sls, dest)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// One 解析到结构体 *Struct
func (d *Db) One(dest any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var m map[string]any
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Limits(1)
	if destPoint, ok := dest.(*map[string]any); ok {
		m, d.err = d.Map()
		if d.err != nil {
			return d.syncRes(r)
		}
		*destPoint = m
		return d.syncRes(r)
	}
	var sl [][]string
	sl, d.err = d.queryToSlices()
	if d.err != nil {
		return d.syncRes(r)
	}
	d.err = SliceToStruct(sl, dest)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// Count 获取统计语句 f 为COUNT(f[0])
func (d *Db) Count(i *int64, f ...string) (r *OrmResult) {
	r = d.syncRes(r)
	var intVal int64
	var res *Result
	res, d.err = d.gs.Count(f...)
	if d.err != nil {
		return d.syncRes(r)
	}
	d.SqlResult = res
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
	} else {
	}
	if d.err != nil {
		return d.syncRes(r)
	}
	if d.err != nil {
		return d.syncRes(r)
	}
	*i = intVal
	return d.syncRes(r)
}

// Update 更新某个字段
func (d *Db) Update(column string, val any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.gs.Update(column, val)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Updates 更新多个字段 data = map[string]any 或者 Struct
func (d *Db) Updates(data any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.gs.Updates(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
func (d *Db) UpdateIgnore(data any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.gs.UpdateIgnore(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Save 存在更新，否则新增 data = map[string]any 或者 Struct
func (d *Db) Save(data any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	if data == nil {
		return d.syncRes(r)
	}
	var m map[string]any
	m, d.err = toMap(data)
	if d.err != nil {
		return d.syncRes(r)
	}
	batchArray := []map[string]any{m}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	return d.save(batchArray)
}

// SaveInBatches 批量新增和更新
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *Db) SaveInBatches(data any, batchSize ...int) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	if data == nil {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var bs = defBatchSize
	if len(batchSize) > 0 {
		bs = batchSize[0]
	}
	var batchArray []map[string]any
	batchArray, d.err = toMaps(data)
	if d.err != nil {
		return d.syncRes(r)
	}
	if len(batchArray) != 0 {
		for i := 0; i < len(batchArray); i += bs {
			end := i + bs
			if end > len(batchArray) {
				end = len(batchArray)
			}
			r = d.save(batchArray[i:end])
			if r.Error != nil {
				return r
			}
		}
	}
	return r
}

// save 存在更新，否则新增
func (d *Db) save(batchArray []map[string]any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.gs.SaveInBatches(batchArray)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Delete 删除 DELETE {t} From table
func (d *Db) Delete(t ...string) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.gs.Delete(t...)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.TxExecDeleteWithCtxAndShowFlag(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.ExecDeleteWithCtxAndShowFlag(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Create 创建 data = map[string]any 或者 Struct
func (d *Db) Create(data any) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	switch vt := data.(type) {
	case *Db:
		data = vt.gs
	case *DbTx:
		data = vt.gs
	}
	var res, e = d.gs.Create(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// CreateInBatches 批量创建
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *Db) CreateInBatches(data any, batchSize ...int) (r *OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var bs = defBatchSize
	if len(batchSize) > 0 {
		bs = batchSize[0]
	}
	var batchArray []map[string]any
	batchArray, d.err = toMaps(data)
	if d.err != nil {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	if len(batchArray) != 0 {
		for i := 0; i < len(batchArray); i += bs {
			end := i + bs
			if end > len(batchArray) {
				end = len(batchArray)
			}
			r = d.createInBatches(batchArray[i:end])
			if r.Error != nil {
				return r
			}
		}
	}
	return r
}

// createInBatches 批量创建
func (d *Db) createInBatches(data []map[string]any) (r *OrmResult) {
	var res, e = d.gs.CreateInBatches(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Joins .
// [query] ON xx = xxx
func (d *Db) Joins(query string, args ...any) *Db {
	if d.err != nil {
		return d
	}
	d.gs.Joins(query, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// Join .
// JOIN table ON xx = xxx
func (d *Db) Join(table, on string, args ...any) *Db {
	if d.err != nil {
		return d
	}
	d.gs.Join(table, on, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// InnerJoin .
// INNER JOIN table ON xx = xxx
func (d *Db) InnerJoin(table, on string, args ...any) *Db {
	if d.err != nil {
		return d
	}
	d.gs.InnerJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// LeftJoin .
// LEFT JOIN table ON xx = xxx
func (d *Db) LeftJoin(table, on string, args ...any) *Db {
	if d.err != nil {
		return d
	}
	d.gs.LeftJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.err
	return d

}

// RightJoin .
// RIGHT JOIN table ON xx = xxx
func (d *Db) RightJoin(table, on string, args ...any) *Db {
	if d.err != nil {
		return d
	}
	d.gs.RightJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// Where 基础查询条件
// db.Where("name = ?", "xxx")
// db.Where("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *Db) Where(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Where(query, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// WhereOr 或查询条件
// db.WhereOr("name = ?", "xxx")
// db.WhereOr("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *Db) WhereOr(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereOr(query, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// WhereReset 重置查询条件
func (d *Db) WhereReset() (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereReset()
	d.err = d.gs.err
	return d
}

// WhereGroup 查询条件组会加括号 AND (id = ? AND name = ?)
func (d *Db) WhereGroup(wf WhereFunc) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereGroup(wf)
	d.err = d.gs.err
	return d
}

// WhereGroupOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *Db) WhereGroupOr(wf WhereFunc) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereGroupOr(wf)
	d.err = d.gs.err
	return d
}

// Table 表名 或者 子查询 每次table都会创建一个新的以支持事务
// db.Table("table_name")
// db.Table("(?) tb",db.Table("tb"))
// db.Table(*Struct{}) 模型结构体
func (d *Db) Table(table any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	newD := d.Clone()
	newD.sqlTx = d.sqlTx
	newD.gs = NewSql()
	newD.gs.Table(table, d.handleArgs(args)...)
	return newD
}

// Clone 复制一个实例
func (d *Db) Clone() (db *Db) {
	return &Db{
		db:                 hhdb.GetDBOP(),
		sqlTx:                nil,
		log:                  d.log,
		gs:                   d.gs.Clone(),
		ctx:                  setLogCallDepthCtx(d.ctx, 3),
		err:                  d.err,
		writeLog:             d.writeLog,
		writeHhDbLog:         d.writeHhDbLog,
		writeCompSql:         d.writeCompSql,
		writeErrSql:          d.writeErrSql,
		fieldTypeReplaceFunc: d.fieldTypeReplaceFunc,
	}
}

// Field 字段列表
// db.Field("id,ts","name"...)
func (d *Db) Field(f ...string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Fields(f)
	d.err = d.gs.err
	return d
}

// Fields 字段
// 字符串 db.Fields("id,ts,? AS name","姓名")
// 切片 db.Fields([]string{"id","ts","? AS name"},"姓名")
func (d *Db) Fields(f any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Fields(f, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// FieldsByData 查询结构体字段 兼容gdb标签
// db.FieldsByData(Struct{},map[string]string{})
// 标签 Field string 使用 Field 字段构建sql
// 标签 Field string `json:"field"` 使用 field 字段构建sql
// 标签 Field string `json:"field" gdbtmp:"db_field"` 使用 db_field 字段构建sql
// 标签 Field string `json:"-"` 忽略这个字段
// 标签 Field string `json:"field" gdbtmp:"-"` 忽略这个字段
// 替换 Field string `json:"field"` replaceField{"field":"db_field"} 使用 db_field 字段构建sql
// 替换 Field string `json:"field"` replaceField{"field":""} 忽略这个字段
func (d *Db) FieldsByData(f any, replaceField map[string]string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.FieldsByData(f, replaceField, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// FieldsByDataAlias 查询结构体字段 兼容gdb标签
// FieldsByData 方法 加上别名 tbAlias.field AS field
func (d *Db) FieldsByDataAlias(f any, replaceField map[string]string, tbAlias string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.FieldsByDataAlias(f, replaceField, tbAlias, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// OmitFields 新增和更新多字段排除的字段
// f[string | []string]
func (d *Db) OmitFields(f any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.OmitFields(f)
	d.err = d.gs.err
	return d
}

// ReplaceFields 新增和更新多字段替换字段
// fm {"field":"db_field"} 字段field更换为db_field
func (d *Db) ReplaceFields(fm map[string]string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.ReplaceFields(fm)
	d.err = d.gs.err
	return d
}

// Limit limit操作
// Limit("?,?", 1,2)
// Limit("?", 2)
// Limit("CASE WHEN type = ? THEN ? WHEN type = ? THEN ? ELSE ?", 1,2,3,4,5)
func (d *Db) Limit(limit string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Limit(limit, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// Limits limit操作
// Limits(1,2)
// Limits(2)
func (d *Db) Limits(l int, l2 ...int) (db *Db) {
	if d.err != nil {
		return d
	}
	if len(l2) > 0 {
		d.gs.Limit("?,?", l, l2[0])
	} else {
		d.gs.Limit("?", l)
	}
	d.err = d.gs.err
	return d
}

// Offset .
func (d *Db) Offset(offset int) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Offset(offset)
	d.err = d.gs.err
	return d
}

// Page .
func (d *Db) Page(page, pageSize int) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Page(page, pageSize)
	d.err = d.gs.err
	return d
}

// PageReset 重置分页
func (d *Db) PageReset() (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.PageReset()
	d.err = d.gs.err
	return d
}

// Group .
// f[string | []string]
func (d *Db) Group(f any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Group(f, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// Order .
// f[string | []string]
func (d *Db) Order(f any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Order(f, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// OrderByFilter 带过滤排序 处理前端排序字段
// OrderByFilter("id desc,aa asc",{"id":"new_field","aa",""}) = new_field desc,aa asc
// ignoreErr 是否忽略错误 false 会返回 xxx 字段不支持排序
func (d *Db) OrderByFilter(orderStr string, filter map[string]string, ignoreErr ...bool) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.OrderByFilter(orderStr, filter, ignoreErr...)
	d.err = d.gs.err
	return d
}

// Union 合并查询
func (d *Db) Union(dbList ...*Db) (db *Db) {
	if d.err != nil {
		return d
	}
	var gsList = make([]*Sql, len(dbList))
	for k, v := range dbList {
		gsList[k] = v.gs
	}
	d.gs.Union(gsList...)
	d.err = d.gs.err
	return d
}

// UnionAll 合并查询
func (d *Db) UnionAll(dbList ...*Db) (db *Db) {
	if d.err != nil {
		return d
	}
	var gsList = make([]*Sql, len(dbList))
	for k, v := range dbList {
		gsList[k] = v.gs
	}
	d.gs.UnionAll(gsList...)
	d.err = d.gs.err
	return d
}

// Raw 原生sql查询
func (d *Db) Raw(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Raw(query, d.handleArgs(args)...)
	d.err = d.gs.err
	return d
}

// GetWhereLen 获取where 条件数量
func (d *Db) GetWhereLen() int {
	return len(d.gs.whereCtrl)
}

// ConvFieldsType 字段转换类型用于map查询
// .Map 或者Maps查询时 返回的map[string]any
// any 类型转换成指定类型
// 请不要修改第三个参数 避免数据竞争
func (d *Db) ConvFieldsType(f ConvVal) (db *Db) {
	if d.err != nil {
		return d
	}
	d.fieldTypeReplaceFunc = f
	return d
}

// GetJoinLen 获取join 条件数量
func (d *Db) GetJoinLen() int {
	return len(d.gs.joinCtrl)
}

// Ctx 设置上下文
func (d *Db) Ctx(ctx context.Context) *Db {
	d.ctx = ctx
	return d
}

// WriteLog 是否打印日志
func (d *Db) WriteLog(flag bool, hhDbFlag ...bool) *Db {
	d.writeLog = flag
	if len(hhDbFlag) > 0 {
		d.writeHhDbLog = hhDbFlag[0]
	}
	return d
}

// WriteHhDbLog 是否打印框架层日志
func (d *Db) WriteHhDbLog(flag bool) *Db {
	d.writeHhDbLog = flag
	return d
}

// WriteErrSql 是否打印错误sql
func (d *Db) WriteErrSql(flag bool) *Db {
	d.writeErrSql = flag
	return d
}

// WriteCompSql 是否完整sql
func (d *Db) WriteCompSql(flag bool) *Db {
	d.writeCompSql = flag
	return d
}

// GetGSql 获取sql构建器
func (d *Db) GetGSql() *Sql {
	return d.gs
}

// Maps 查询map列表
func (d *Db) Maps() (result []map[string]any, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var res *Result
	res, d.err = d.gs.Select()
	if d.err != nil {
		return nil, d.err
	}
	d.SqlResult = res
	var reResultsByList hhdb.ReResultsByMap01
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reResultsByList = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		reResultsByList = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	//d.listResultsByList = reResultsByList
	d.err = reResultsByList.
	d.printLog(reResultsByList..String(), tx, res)
	if reResultsByList. != nil && d.fieldTypeReplaceFunc != nil {
		// 新建map防止数据竞争
		var newMaps = make([]map[string]any, len(.))
		for k, v := range {
			var newMap = make(map[string]any, len(v))
			for mk, mv := range v {
				newMap[mk] = d.fieldTypeReplaceFunc(mk, mv, NewReadOnlyMap(v))
			}
			newMaps[k] = newMap
		}
		reResultsByList.ReData = newMaps
	}
	return , d.err
}

// Map 查询map
func (d *Db) Map() (res map[string]any, err error) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Limits(1)
	var maps []map[string]any
	maps, d.err = d.Maps()
	if d.err != nil {
		return nil, d.err
	}
	if len(maps) > 0 {
		return maps[0], nil
	}
	return map[string]any{}, nil
}

// Transaction 闭包事务
func (d *Db) Transaction(f func(db *DbTx) error) (err error) {
	var sqlDb *sql.DB
	err, sqlDb = d.db.()
	if err != nil {
		d.err = err
		return err
	}
	var tTx *sql.Tx
	tTx, err = sqlDb.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tTx.Commit()
			d.err = err
		} else {
			_ = tTx.Rollback()
		}
	}()
	dbTx := &DbTx{
		New(),
	}
	dbTx.Db.sqlTx = &Tx{tTx}
	err = f(dbTx)
	return err
}

// Tx 设置事务
func (d *Db) Tx(tx *Tx) *Db {
	d.sqlTx = tx
	return d
}

// GetTx 获取事务对象
func (d *Db) GetTx() *Tx {
	return d.sqlTx
}

// LogLevel 设置当前操纵日志等级
func (d *Db) LogLevel(lv LogLevel) (db *Db) {
	d.log.SetLevel(lv)
	return d
}

// Stream 流式查询
// sc = gdbtmp.StreamCallback(func(d map[string]any|你的结构体) (err error, next bool) {})
func (d *Db) Stream(sc streamCallbackFace) (r *OrmResult) {
	var res *Result
	res, d.err = d.gs.Select()
	if d.err != nil {
		return d.syncRes(r)
	}
	var taskId string
	var cols []string
	d.err, taskId, cols = d.db.(d.ctx, d.writeHhDbLog, res.Sql, res.Args...)
	if d.err != nil {
		return d.syncRes(r)
	}
	defer d.db.(taskId)
	var resetOneByOneTimer = 1 * time.Second
	var resetOneByOne = time.NewTicker(resetOneByOneTimer)
	defer resetOneByOne.Stop()
	go func() { // 刷新时间
		for {
			select {
			case <-d.ctx.Done():
				return
			case <-resetOneByOne.C:
				_ = d.db.(taskId)
			}
		}
	}()
	var stop bool
	var rows []string
	var err error
	var next bool
	for {
		select {
		case <-d.ctx.Done():
			defLog.CtxDebug(d.ctx, "context Done")
			return d.syncRes(r)
		default:
			err, stop, rows = d.db.(taskId)
			if err != nil {
				d.err = errors.Join(d.err, err)
				return d.syncRes(r)
			}
			if stop {
				return d.syncRes(r)
			}
			var m = make(map[string]any, len(cols))
			for k, v := range rows {
				m[cols[k]] = v
			}
			err, next = sc.call(rows, cols)
			if err != nil {
				d.err = errors.Join(d.err, err)
			}
			if !next {
				return d.syncRes(r)
			}
			resetOneByOne.Reset(resetOneByOneTimer)
		}
	}
}

// queryToSlices 执行查询返回 [][]string
func (d *Db) queryToSlices() (result [][]string, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var res *Result
	res, d.err = d.gs.Select()
	if d.err != nil {
		return nil, d.err
	}
	d.SqlResult = res
	var reResultsByList hhdb.ReResultsByList
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reResultsByList = d.db.(d.ctx, d.sqlTx.Tx, res.Sql, d.writeHhDbLog, res.Args...)
	} else {
		reResultsByList = d.db.(d.ctx, res.Sql, d.writeHhDbLog, res.Args...)
	}
	//d.listResultsByList = reResultsByList
	d.err = reResultsByList.
	d.printLog(reResultsByList..String(), tx, res)
	return reResultsByList., d.err
}

// handleArgs 处理参数，兼容子查询
func (d *Db) handleArgs(args []any) []any {
	newArgs := make([]any, len(args))
	for k, v := range args {
		switch vt := v.(type) {
		case *Db:
			newArgs[k] = vt.gs
		case *DbTx:
			newArgs[k] = vt.gs
		default:
			newArgs[k] = v
		}
	}
	return newArgs
}

// getPrintSql 获取打印的sql
func (d *Db) getPrintSql(r *Result, lv LogLevel) string {
	if lv == ErrorLogLevel && !d.writeErrSql {
		return "none"
	}
	if d.writeCompSql {
		return r.CompSql()
	}
	return r.Sql
}

// printLog 打印日志
func (d *Db) printLog(st string, txKey string, res *Result) {
	if d.writeLog {
		d.ctx = appendLogCallDepthCtx(d.ctx, 1)
		if d.err == nil {
			d.log.CtxDebugf(d.ctx, "[%s] [Affect:%d] %s%s", st, d.execResult.AffectRecs, txKey, d.getPrintSql(res, DebugLogLevel))
		} else {
			d.log.CtxErrorf(d.ctx, "Sql Err %s%s %v", txKey, d.getPrintSql(res, ErrorLogLevel), d.err)
		}
	}
}
