package gdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	hhdb "nwgit.gzhhit.com/BD/hhitdb.git"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"reflect"
	"strings"
	"time"
)

type ExecType string

const (
	txKey        = "Tx " // 日志打印前缀
	defBatchSize = 100   // 默认分批大小

	ExecTypeInsert  = "INSERT"
	ExecTypeUpdate  = "UPDATE"
	ExecTypeDelete  = "DELETE"
	ExecTypeReplace = "REPLACE"

	WhereOpFindInSet = "FIND_IN_SET"
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

// WhereFunc 闭包where 给where 加 ()
type WhereFunc func(db *Db) (err error)

// Db 操作对象
type Db struct {
	HhDb                 hhdb.SitkDB
	sqlTx                *sql.Tx
	SqlResult            *Result
	log                  GdbLog
	gs                   *Sql
	ctx                  context.Context
	sqlDrive             string
	sqlWrapper           map[string]NewWrapperFunc
	writeLog             bool    // 是否打印日志
	writeHhDbLog         bool    // 是否打印框架日志
	writeErrSql          bool    // 是否打印错误sql
	writeCompSql         bool    // 是否打印完整sql
	err                  error   // 最终错误
	emptyError           error   // 空错误
	fieldTypeReplaceFunc ConvVal // 更新和新增字段类型替换

	// 处理参数
	dbConvInitPtr bool // 初始化指针
	timeLocation  *time.Location

	//listResultsByList    hhdb.ReResultsByList
	//mapResultsByList     hhdb.ReResultsByMap01
	isQuery    bool
	execResult ExecResult
}

// syncRes 同步结果
func (d *Db) syncRes(r OrmResult) OrmResult {
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
	//ListResultsByList hhdb.ReResultsByList // 不保留副本Ò
	ExecResult ExecResult
}

// GetLastInsertId 获取最后插入的ID
func (or OrmResult) GetLastInsertId() int64 {
	return or.ExecResult.LastAffectId
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
		HhDb:          hhdb.GetDBOP(),
		gs:            NewSql(table...),
		log:           defLog.Clone(),
		writeHhDbLog:  conf.WriteHhDbLog,
		sqlDrive:      conf.DriveType,
		sqlWrapper:    conf.DriveMap,
		writeLog:      conf.WriteLog,
		writeErrSql:   conf.WriteErrSql,
		writeCompSql:  conf.WriteCompSql,
		dbConvInitPtr: conf.DbConvInitPtr,
		timeLocation:  conf.TimeLocation,
	}
	if db.HhDb == nil {
		db.err = errors.New("未初始化数据库")
	}

	if conf.EmptyError != nil {
		db.emptyError = conf.EmptyError
	}

	db.Ctx(ctx)
	return db
}

// Scan 解析到结构体 如果是数组 []any(Struct) 兼容单结构体和数组
func (d *Db) Scan(dest any) (r OrmResult) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	if conv.IsArrayOrSlice(dest) {
		return d.Select(dest)
	} else {
		return d.One(dest)
	}
}

// ScanAndCount 解析到结构体 []any(Struct)
func (d *Db) ScanAndCount(dest any, count *int64) (r OrmResult) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	return d.SelectAndCount(dest, count)
}

// Select 解析到数组结构体 []any(Struct)
func (d *Db) Select(dest any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var ms []map[string]any
	ms, d.err = d.Maps()
	if d.err != nil {
		return d.syncRes(r)
	}
	if destPoint, ok := dest.(*[]map[string]any); ok {
		*destPoint = ms
		return d.syncRes(r)
	}
	d.err = conv.MapsToStruct(ms, dest, d.dbConvInitPtr)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// SelectAndCount 解析到数组结构体并统计数量 []any(Struct)
func (d *Db) SelectAndCount(dest any, count *int64) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var ms []map[string]any
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var ctx = setLogCallDepthCtx(d.ctx, getLogCallDepthCtx(d.ctx))
	var totalRes OrmResult
	var totalQuery = d.Clone().PageReset()
	if totalQuery.gs.GetGroupLen() != 0 {
		totalQuery = New().Table("(?) t", totalQuery)
	}
	totalRes = totalQuery.Ctx(ctx).Count(count)
	if totalRes.Error != nil {
		return totalRes
	}
	if *count == 0 { // 无数据不再查询
		if d.emptyError != nil {
			d.err = d.emptyError
		}
		return d.syncRes(r)
	}
	ms, d.err = d.Maps()
	if d.err != nil {
		return d.syncRes(r)
	}
	if destPoint, ok := dest.(*[]map[string]any); ok {
		*destPoint = ms
		return d.syncRes(r)
	}
	d.err = conv.MapsToStruct(ms, dest, d.dbConvInitPtr)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// One 解析到结构体 *Struct
func (d *Db) One(dest any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var m map[string]any
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Limits(1)
	m, d.err = d.Map()
	if d.err != nil {
		return d.syncRes(r)
	}
	if destPoint, ok := dest.(*map[string]any); ok {
		*destPoint = m
		return d.syncRes(r)
	}
	d.err = conv.MapToStruct(m, dest, d.dbConvInitPtr)
	if d.err != nil {
		return d.syncRes(r)
	}
	return d.syncRes(r)
}

// Count 获取统计语句 f 为COUNT(f[0])
func (d *Db) Count(i *int64, f ...string) (r OrmResult) {
	r = d.syncRes(r)
	var intVal int64
	var res *Result
	res, d.err = d.getSqlWrapper().Count(f...)
	if d.err != nil {
		return d.syncRes(r)
	}
	d.SqlResult = res
	var reSingleResults hhdb.ReSingleResults
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reSingleResults = d.HhDb.TxGetSingleResultWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		reSingleResults = d.HhDb.GetSingleResultWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.isQuery = true
	d.err = reSingleResults.ExecState
	d.printLog(reSingleResults.ExecCost.String(), tx, res)
	if d.err != nil {
		return d.syncRes(r)
	}
	if reSingleResults.ReData == "" {
		return d.syncRes(r)
	}
	intVal, d.err = conv.ToInt64(reSingleResults.ReData)
	if d.err != nil {
		return d.syncRes(r)
	}
	*i = intVal
	return d.syncRes(r)
}

// GetCount 获取统计语句 f 为COUNT(f[0])
func (d *Db) GetCount(f ...string) (int64, error) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var count int64
	var res = d.Count(&count, f...)
	return count, res.Error
}

// Exists 判断是否存在
func (d *Db) Exists() (bool, error) {
	var intVal int64
	var res *Result
	res, d.err = d.getSqlWrapper().Exists()
	if d.err != nil {
		return false, d.err
	}
	d.SqlResult = res
	var reSingleResults hhdb.ReSingleResults
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reSingleResults = d.HhDb.TxGetSingleResultWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		reSingleResults = d.HhDb.GetSingleResultWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.isQuery = true
	d.err = reSingleResults.ExecState
	d.printLog(reSingleResults.ExecCost.String(), tx, res)
	if d.err != nil {
		return false, d.err
	}
	if reSingleResults.ReData == "" {
		return false, nil
	}
	intVal, d.err = conv.ToInt64(reSingleResults.ReData)
	if d.err != nil {
		return false, d.err
	}
	return intVal == 1, nil
}

// Update 更新某个字段
func (d *Db) Update(column string, val any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.getSqlWrapper().Update(column, val)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Updates 更新多个字段 data = map[string]any 或者 Struct
func (d *Db) Updates(data any, whereField ...[]string) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().Updates(data, whereField...)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
func (d *Db) UpdateIgnore(data any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().UpdateIgnore(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Save 存在更新，否则新增 data = map[string]any 或者 Struct
func (d *Db) Save(data any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	if data == nil {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)

	if d.isMap(reflect.ValueOf(data)) {
		var e error
		var dataMap map[string]any
		dataMap, e = MapToMapAny(data)
		if e == nil { // 是map
			return d.save([]map[string]any{dataMap})
		}
	}

	return d.save([]any{data})
}

// SaveInBatches 批量新增和更新
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *Db) SaveInBatches(data any, batchSize ...int) (r OrmResult) {
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
	var sliceRef = reflect.ValueOf(data)
	if sliceRef.Kind() != reflect.Slice {
		d.err = errors.New("data必须是切片")
		return d.syncRes(r)
	}
	if sliceRef.Len() == 0 {
		return d.syncRes(r)
	}
	if d.isMap(sliceRef.Index(0)) {
		// 断言map 减少反射
		var mapList, e = MapsToMapsAny(data)
		if e == nil { // 是map 直接处理避免反射
			for i := 0; i < len(mapList); i += bs {
				end := i + bs
				if end > len(mapList) {
					end = len(mapList)
				}
				r = d.save(mapList[i:end])
				if r.Error != nil {
					return d.syncRes(r)
				}
			}
			return d.syncRes(r)
		}
	}

	// 不是map 按any切片处理
	var batchArray = make([]any, 0, sliceRef.Len())
	for i := 0; i < sliceRef.Len(); i++ {
		batchArray = append(batchArray, sliceRef.Index(i).Interface())
	}
	for i := 0; i < len(batchArray); i += bs {
		end := i + bs
		if end > len(batchArray) {
			end = len(batchArray)
		}
		r = d.save(batchArray[i:end])
		if r.Error != nil {
			return d.syncRes(r)
		}
	}
	return d.syncRes(r)
}

// save 存在更新，否则新增
func (d *Db) save(batchArray any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().SaveInBatches(batchArray)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecInsertWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecInsertWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Delete 删除 DELETE {t} From table
func (d *Db) Delete(t ...string) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.getSqlWrapper().Delete(t...)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecDeleteWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecDeleteWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// Create 新增数据
// data = map[string]any 或者 Struct
// data = *Db insert init ... select  从查询新增
func (d *Db) Create(data any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	switch vt := data.(type) {
	case *Db:
		data = vt.gs
	case *DbTx:
		data = vt.gs
	}
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().Create(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecInsertWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecInsertWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// CreateInBatches 批量创建
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *Db) CreateInBatches(data any, batchSize ...int) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var bs = defBatchSize
	if len(batchSize) > 0 {
		bs = batchSize[0]
	}
	var sliceRef = reflect.ValueOf(data)
	if sliceRef.Kind() != reflect.Slice {
		d.err = errors.New("data必须是切片")
		return d.syncRes(r)
	}
	if sliceRef.Len() == 0 {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)

	// 断言map 减少反射
	if d.isMap(sliceRef.Index(0)) {
		var mapList, e = MapsToMapsAny(data)
		if e == nil { // 是map 直接处理避免反射
			for i := 0; i < len(mapList); i += bs {
				end := i + bs
				if end > len(mapList) {
					end = len(mapList)
				}
				r = d.createInBatches(mapList[i:end])
				if r.Error != nil {
					return d.syncRes(r)
				}
			}
			return d.syncRes(r)
		}
	}

	// 不是map 按any切片处理
	var batchArray = make([]any, 0, sliceRef.Len())
	for i := 0; i < sliceRef.Len(); i++ {
		batchArray = append(batchArray, sliceRef.Index(i).Interface())
	}
	for i := 0; i < len(batchArray); i += bs {
		end := i + bs
		if end > len(batchArray) {
			end = len(batchArray)
		}
		r = d.createInBatches(batchArray[i:end])
		if r.Error != nil {
			return d.syncRes(r)
		}
	}
	return d.syncRes(r)
}

// Replace 新增或替换数据
// data = map[string]any 或者 Struct
// data = *Db insert init ... select  从查询新增
func (d *Db) Replace(data any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	switch vt := data.(type) {
	case *Db:
		data = vt.gs
	case *DbTx:
		data = vt.gs
	}
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().Replace(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// ReplaceInBatches 批量创建或替换
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *Db) ReplaceInBatches(data any, batchSize ...int) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var bs = defBatchSize
	if len(batchSize) > 0 {
		bs = batchSize[0]
	}
	var sliceRef = reflect.ValueOf(data)
	if sliceRef.Kind() != reflect.Slice {
		d.err = errors.New("data必须是切片")
		return d.syncRes(r)
	}
	if sliceRef.Len() == 0 {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)

	if d.isMap(sliceRef.Index(0)) {
		// 断言map 减少反射
		var mapList, e = MapsToMapsAny(data)
		if e == nil { // 是map 直接处理避免反射
			for i := 0; i < len(mapList); i += bs {
				end := i + bs
				if end > len(mapList) {
					end = len(mapList)
				}
				r = d.replaceInBatches(mapList[i:end])
				if r.Error != nil {
					return d.syncRes(r)
				}
			}
			return d.syncRes(r)
		}
	}

	// 不是map 按any切片处理
	var batchArray = make([]any, 0, sliceRef.Len())
	for i := 0; i < sliceRef.Len(); i++ {
		batchArray = append(batchArray, sliceRef.Index(i).Interface())
	}

	for i := 0; i < len(batchArray); i += bs {
		end := i + bs
		if end > len(batchArray) {
			end = len(batchArray)
		}
		r = d.replaceInBatches(batchArray[i:end])
		if r.Error != nil {
			return d.syncRes(r)
		}
	}
	return d.syncRes(r)
}

// createInBatches 批量创建
func (d *Db) createInBatches(data any) (r OrmResult) {
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().CreateInBatches(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecInsertWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecInsertWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// replaceInBatches 批量创建或替换
func (d *Db) replaceInBatches(data any) (r OrmResult) {
	d.gs.Ctx(d.ctx)
	var res, e = d.getSqlWrapper().ReplaceInBatches(data)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// IgnoreDuplicate 更新条件
// 设置 field = true
// Duplicate 会忽略该字段
func (d *Db) IgnoreDuplicate(fieldIgnore map[string]bool) (db *Db) {
	d.gs.IgnoreDuplicate(fieldIgnore)
	return d
}

// SetDuplicate 更新条件
// name = IfNull(VALUES(name),”)
func (d *Db) SetDuplicate(keyUpdate map[string]string) (db *Db) {
	d.gs.SetDuplicate(keyUpdate)
	return d
}

// Joins .
// [query] ON xx = xxx
func (d *Db) Joins(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Joins(query, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// Join .
// JOIN table ON xx = xxx
func (d *Db) Join(table, on string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Join(table, on, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// InnerJoin .
// INNER JOIN table ON xx = xxx
func (d *Db) InnerJoin(table, on string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.InnerJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// LeftJoin .
// LEFT JOIN table ON xx = xxx
func (d *Db) LeftJoin(table, on string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.LeftJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d

}

// RightJoin .
// RIGHT JOIN table ON xx = xxx
func (d *Db) RightJoin(table, on string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.RightJoin(table, on, d.handleArgs(args)...)
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
	return d
}

// WhereCond 动态处理 OR AND
func (d *Db) WhereCond(cond Cond, query string, args ...any) (db *Db) {
	if cond == CondAnd {
		return d.Where(query, args...)
	}
	return d.WhereOr(query, args...)
}

// WhereReset 重置查询条件
func (d *Db) WhereReset() (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereReset()
	d.err = d.gs.Err()
	return d
}

// WhereGroup 查询条件组会加括号 AND (id = ? AND name = ?)
func (d *Db) WhereGroup(wf WhereFunc) (db *Db) {
	if d.err != nil {
		return d
	}
	var gdb = New()
	d.err = wf(gdb)
	if d.err != nil {
		return d
	}
	d.gs.appendWhereGroup(gdb.GetGSql(), CondAnd)
	d.err = d.gs.err
	return d
}

// WhereGroupOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *Db) WhereGroupOr(wf WhereFunc) (db *Db) {
	if d.err != nil {
		return d
	}
	var gdb = New()
	d.err = wf(gdb)
	if d.err != nil {
		return d
	}
	d.gs.appendWhereGroup(gdb.GetGSql(), CondOr)
	d.err = d.gs.err
	return d
}

// WhereBlock 查询条件组会加括号 AND (id = ? AND name = ?)
func (d *Db) WhereBlock(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	return d.Where("("+query+")", args...)
}

// WhereBlockOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *Db) WhereBlockOr(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	return d.WhereOr("("+query+")", args...)
}

// WhereGroupCond 动态处理 AND OR
func (d *Db) WhereGroupCond(cond Cond, wf WhereFunc) (db *Db) {
	if cond == CondAnd {
		return d.WhereGroup(wf)
	}
	return d.WhereGroupOr(wf)
}

// Wrapper 查询构建器
// 会覆盖，请在最前面使用
func (d *Db) Wrapper(wrap *Wrapper) (db *Db) {
	if wrap == nil {
		return d
	}
	if wrap.Db.gs.GetTable() == "" {
		wrap.Db.gs.Table(d.gs.GetTable())
	}
	d.gs = wrap.Db.gs
	return d
}

// Table 表名 或者 子查询
// db.Table("table_name")
// db.Table("(?) tb",db.Table("tb"))
// db.Table(*Struct{}) 模型结构体
func (d *Db) Table(table any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Table(table, d.handleArgs(args)...)
	return d
}

// As 表别名
func (d *Db) As(alias string) (db *Db) {
	d.gs.As(alias)
	return d
}

// tableClone 表名 或者 子查询 每次table都会创建一个新的以支持事务
func (d *Db) tableClone(table any, args ...any) (db *Db) {
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
		HhDb:                 hhdb.GetDBOP(),
		sqlTx:                nil,
		log:                  d.log,
		gs:                   d.gs.Clone(),
		ctx:                  setLogCallDepthCtx(d.ctx, 3),
		err:                  d.err,
		emptyError:           d.emptyError,
		writeLog:             d.writeLog,
		writeHhDbLog:         d.writeHhDbLog,
		writeCompSql:         d.writeCompSql,
		writeErrSql:          d.writeErrSql,
		fieldTypeReplaceFunc: d.fieldTypeReplaceFunc,
		dbConvInitPtr:        d.dbConvInitPtr,
		timeLocation:         d.timeLocation,
	}
}

// Field 字段列表
// db.Field("id,ts","name"...)
func (d *Db) Field(f ...string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Fields(f)
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
	return d
}

// FieldsByData 查询结构体字段 兼容gdb标签
// db.FieldsByData(Struct{},map[string]string{})
// 标签 Field string 使用 Field 字段构建sql
// 标签 Field string `json:"field"` 使用 field 字段构建sql
// 标签 Field string `json:"field" gdb:"db_field"` 使用 db_field 字段构建sql
// 标签 Field string `json:"-"` 忽略这个字段
// 标签 Field string `json:"field" gdb:"-"` 忽略这个字段
// 替换 Field string `json:"field"` replaceField{"field":"db_field"} 使用 db_field 字段构建sql
// 替换 Field string `json:"field"` replaceField{"field":""} 忽略这个字段
func (d *Db) FieldsByData(f any, replaceField map[string]string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.FieldsByData(f, replaceField, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// FieldsByDataAlias 查询结构体字段 兼容gdb标签
// FieldsByData 方法 加上别名 tbAlias.field AS field
func (d *Db) FieldsByDataAlias(f any, replaceField map[string]string, tbAlias string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.FieldsByDataAlias(f, replaceField, tbAlias, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// OmitFields 新增和更新多字段排除的字段
// f[string | []string]
func (d *Db) OmitFields(f any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.OmitFields(f)
	d.err = d.gs.Err()
	return d
}

// ReplaceFields 新增和更新多字段替换字段
// fm {"field":"db_field"} 字段field更换为db_field
func (d *Db) ReplaceFields(fm map[string]string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.ReplaceFields(fm)
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
	return d
}

// Offset .
func (d *Db) Offset(offset int) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Offset(offset)
	d.err = d.gs.Err()
	return d
}

// Page .
func (d *Db) Page(page, pageSize int) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Page(page, pageSize)
	d.err = d.gs.Err()
	return d
}

// PageReset 重置分页
func (d *Db) PageReset() (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.PageReset()
	d.err = d.gs.Err()
	return d
}

// Group .
// f[string | []string]
func (d *Db) Group(f any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Group(f, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// Groups .
// f string
func (d *Db) Groups(f ...string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Group(f)
	d.err = d.gs.Err()
	return d
}

// Order .
// f[string | []string]
func (d *Db) Order(f any, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Order(f, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// Orders .
// f string
func (d *Db) Orders(f ...string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Order(f)
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
	return d
}

// WhereSearchView 处理表单查询条件 使用视图字段过滤
// 视图推荐构建 视图字段配置 处理前端查询条件
// filterField 数组允许的查询字段
// filterField {"id":"new_field"} 不支持空val
func (d *Db) WhereSearchView(sD types.DataViewSearchInfo, sSItem []types.SearchItem, filterField map[string]string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereSearchView(sD, sSItem, filterField)
	d.err = d.gs.Err()
	return d
}

// WhereSearch 处理表单查询条件 忽略视图配置
// 处理前端查询条件
// filterField 数组允许的查询字段
// ignoreErr 是否忽略错误 false 会返回 搜索条件错误,搜索字段不存在
func (d *Db) WhereSearch(sSItem []types.SearchItem, filterField map[string]string, ignoreErr ...bool) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.WhereSearch(sSItem, filterField, ignoreErr...)
	d.err = d.gs.Err()
	return d
}

// WhereSearchParams 处理前端搜索查询条件
// 处理前端查询条件
func (d *Db) WhereSearchParams(sp *HhSearchParam) (db *Db) {
	if d.err != nil {
		return d
	}
	if sp == nil {
		return d
	}

	if len(sp.SearchParams) > 0 {
		d.WhereSearch(sp.SearchParams, nil)
	}

	if sp.SortParams != "" {
		d.Order(sp.SortParams)
	}

	if sp.CurrPage > 0 && sp.PageNums > 0 {
		d.Page(sp.CurrPage, sp.PageNums)
	}
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
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
	d.err = d.gs.Err()
	return d
}

// Raw 原生sql查询
func (d *Db) Raw(query string, args ...any) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.Raw(query, d.handleArgs(args)...)
	d.err = d.gs.Err()
	return d
}

// RawExec 原生sql更新
func (d *Db) RawExec(execSql string, args ...any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	return d.RawExecType(execSql, "", args...)
}

// RawExecType 原生sql更新
func (d *Db) RawExecType(execSql string, execType ExecType, args ...any) (r OrmResult) {
	if d.err != nil {
		return d.syncRes(r)
	}
	var res, e = d.getSqlWrapper().RawExec(execSql, d.handleArgs(args)...)
	if e != nil {
		d.err = e
		return d.syncRes(r)
	}

	if execType == "" {
		var prefix = ""
		var sqlStr = res.Sql.String()
		for i := 0; i < len(sqlStr); i++ {
			if sqlStr[i] == ' ' {
				prefix = strings.ToUpper(sqlStr[:i])
				break
			}
		}
		switch prefix {
		case ExecTypeUpdate, ExecTypeReplace:
			execType = ExecTypeUpdate
		case ExecTypeInsert:
			execType = ExecTypeInsert
		case ExecTypeDelete:
			execType = ExecTypeDelete
		default:
			d.err = errors.New("recognition sql ExecType Error")
		}
	}
	if d.err != nil {
		return d.syncRes(r)
	}
	d.SqlResult = res
	var st = time.Now()
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		switch execType {
		case ExecTypeUpdate, ExecTypeReplace:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecUpdateWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		case ExecTypeInsert:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecInsertWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		case ExecTypeDelete:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.TxExecDeleteWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		default:
			d.err = errors.New("ExecType is invalid")
		}
	} else {
		switch execType {
		case ExecTypeUpdate, ExecTypeReplace:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecUpdateWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		case ExecTypeInsert:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecInsertWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		case ExecTypeDelete:
			d.err, d.execResult.AffectRecs, d.execResult.LastAffectId = d.HhDb.ExecDeleteWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
		default:
			d.err = errors.New("ExecType is invalid")
		}
	}
	d.printLog(time.Since(st).String(), tx, res)
	return d.syncRes(r)
}

// GetWhereLen 获取where 条件数量
func (d *Db) GetWhereLen() int {
	return d.gs.GetWhereLen()
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
	return d.gs.GetJoinLen()
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
	res, d.err = d.getSqlWrapper().Select()
	if d.err != nil {
		return nil, d.err
	}
	d.SqlResult = res
	var reResultsByList hhdb.ReResultsByMap01
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reResultsByList = d.HhDb.TxGetRowByMapWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		reResultsByList = d.HhDb.GetRowByMapWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	//d.listResultsByList = reResultsByList
	d.isQuery = true
	d.err = reResultsByList.ExecState
	d.printLog(reResultsByList.ExecCost.String(), tx, res)
	if d.err != nil {
		return nil, d.err
	}
	if len(reResultsByList.ReData) == 0 {
		reResultsByList.ReData = []map[string]any{}
		if d.emptyError != nil {
			d.err = d.emptyError
			return reResultsByList.ReData, d.err
		}
	}
	if len(reResultsByList.ReData) > 0 && d.fieldTypeReplaceFunc != nil {
		// 新建map防止数据竞争
		var newMaps = make([]map[string]any, len(reResultsByList.ReData))
		for k, v := range reResultsByList.ReData {
			var newMap = make(map[string]any, len(v))
			for mk, mv := range v {
				newMap[mk] = d.fieldTypeReplaceFunc(mk, mv, NewReadOnlyMap(v))
			}
			newMaps[k] = newMap
		}
		reResultsByList.ReData = newMaps
	}
	return reResultsByList.ReData, d.err
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

// TypeMaps 查询带类型的map列表
func (d *Db) TypeMaps(t map[string]conv.ValType) (result []map[string]any, err error) {
	result = make([]map[string]any, 0)
	if d.err != nil {
		return result, d.err
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	result, d.err = d.Maps()
	if d.err != nil {
		return result, d.err
	}
	if len(result) <= 0 {
		return result, nil
	}
	var vt conv.ValType
	var ok bool
	for k, m := range result {
		for col, mv := range m {
			if vt, ok = t[col]; ok {
				result[k][col] = conv.ToValType(mv, vt)
			}
		}
	}
	return result, nil
}

// TypeMap 查询带类型的map
func (d *Db) TypeMap(t map[string]conv.ValType) (result map[string]any, err error) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Limits(1)
	var maps []map[string]any
	maps, d.err = d.TypeMaps(t)
	if d.err != nil {
		return nil, d.err
	}
	if len(maps) > 0 {
		return maps[0], nil
	}
	return map[string]any{}, nil
}

// OrderMaps 查询有序map列表
func (d *Db) OrderMaps() (result []*OrderedMap[string, any], err error) {
	result = make([]*OrderedMap[string, any], 0)
	if d.err != nil {
		return result, d.err
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	var sls [][]string
	sls, d.err = d.queryToSlices()
	if d.err != nil {
		return result, d.err
	}
	if len(sls) <= 1 {
		return result, nil
	}
	var columns = sls[0]
	for i := 1; i < len(sls); i++ {
		var mp = NewOrderedMap[string, any]()
		for j := 0; j < len(sls[i]); j++ {
			mp.Set(sls[0][j], sls[i][j])
		}
		mp.SetKey(columns)
		result = append(result, mp)
	}
	return result, err
}

// OrderMap 查询OrderMap
func (d *Db) OrderMap() (res *OrderedMap[string, any], err error) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Limits(1)
	var maps []*OrderedMap[string, any]
	maps, d.err = d.OrderMaps()
	if d.err != nil {
		return nil, d.err
	}
	if len(maps) > 0 {
		return maps[0], nil
	}
	return NewOrderedMap[string, any](), nil
}

// Pluck 查询一列
// slice 切片 &[][string,int...]
func (d *Db) Pluck(field string, slice any) (r OrmResult) {
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	d.Field(field)
	var sliceList [][]string
	sliceList, d.err = d.queryToSlices()
	if d.err != nil {
		return d.syncRes(r)
	}
	if len(sliceList) <= 1 {
		return d.syncRes(r)
	}
	if len(sliceList[0]) > 1 {
		d.err = errors.New("pluck 查询返回字段必须唯一")
		return d.syncRes(r)
	}
	d.err = conv.SliceToSlice(sliceList, slice)
	return d.syncRes(r)
}

// Slices 查询切片列表
func (d *Db) Slices() (result [][]string, err error) {
	if d.err != nil {
		return nil, d.err
	}
	d.ctx = appendLogCallDepthCtx(d.ctx, 1)
	return d.queryToSlices()
}

// Transaction 闭包事务
func (d *Db) Transaction(f func(db *DbTx) error) (err error) {
	var sqlDb *sql.DB
	err, sqlDb = d.HhDb.GetSitkDBObj()
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
	dbTx.Db.sqlTx = tTx
	err = f(dbTx)
	return err
}

// Tx 设置事务
func (d *Db) Tx(tx *sql.Tx) *Db {
	d.sqlTx = tx
	return d
}

// GetTx 获取事务对象
func (d *Db) GetTx() *sql.Tx {
	return d.sqlTx
}

// LogLevel 设置当前操纵日志等级
func (d *Db) LogLevel(lv LogLevel) (db *Db) {
	d.log.SetLevel(lv)
	return d
}

// LogCallDepth 修改当前操纵日志深度
func (d *Db) LogCallDepth(callDepth int) (db *Db) {
	d.ctx = appendLogCallDepthCtx(d.ctx, callDepth)
	return d
}

// EmptyError 未找到数据返回错误 默认返回空
func (d *Db) EmptyError(err ...error) (db *Db) {
	if len(err) > 0 && err[0] != nil {
		var dbErr DbError
		if errors.As(err[0], &dbErr) {
			dbErr.isNotFound = true
			d.emptyError = dbErr
		} else {
			d.emptyError = DbError{
				isNotFound: true,
				Code:       "",
				Msg:        err[0].Error(),
			}
		}
	} else {
		d.emptyError = ErrRecordNotFound
	}
	return d
}

// Stream 流式查询
// sc = gdb.StreamCallback(func(d map[string]any|你的结构体) (err error, next bool) {})
func (d *Db) Stream(sc streamCallbackFace) (r OrmResult) {
	var res *Result
	res, d.err = d.getSqlWrapper().Select()
	if d.err != nil {
		return d.syncRes(r)
	}
	var taskId string
	var cols []string
	d.err, taskId, cols = d.HhDb.GetRowViaOneByOneStart(d.ctx, d.writeHhDbLog, res.Sql.String(), res.Args...)
	if d.err != nil {
		return d.syncRes(r)
	}
	d.isQuery = true
	var ctx, cancel = context.WithCancel(d.ctx)
	defer cancel()
	defer d.HhDb.GetRowOneByOneStop(taskId)
	var resetOneByOneTimer = 1 * time.Second
	var resetOneByOne = time.NewTicker(resetOneByOneTimer)
	defer resetOneByOne.Stop()
	go func() { // 刷新时间
		for {
			select {
			case <-ctx.Done():
				return
			case <-resetOneByOne.C:
				_ = d.HhDb.UpdateRowOneByOneItem(taskId)
			}
		}
	}()
	var stop bool
	var rows []string
	var err error
	var next bool
	for {
		select {
		case <-ctx.Done():
			defLog.CtxDebug(d.ctx, "context Done")
			return d.syncRes(r)
		default:
			err, stop, rows = d.HhDb.GetRowDataOneByOneByList(taskId)
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
			next, err = sc.call(m, cols, d.dbConvInitPtr)
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

// CteQuery 虚拟表查询
// WITH
//
//	临时表1名称 [(列名1, 列名2, ...)] AS (SELECT/INSERT/UPDATE/DELETE 语句),
//	{tableAlias} AS ({query}),
//
// cte 默认 WITH
func (d *Db) CteQuery(query *CteQuery, cte ...string) (db *Db) {
	if d.err != nil {
		return d
	}
	d.gs.CteQuery(query, cte...)
	d.err = d.gs.Err()
	return d
}

// queryToSlices 执行查询返回 [][]string
func (d *Db) queryToSlices() (result [][]string, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var res *Result
	res, d.err = d.getSqlWrapper().Select()
	if d.err != nil {
		return nil, d.err
	}
	d.SqlResult = res
	var reResultsByList hhdb.ReResultsByList
	var tx = ""
	if d.sqlTx != nil {
		tx = txKey
		reResultsByList = d.HhDb.TxGetRowByListWithCtxAndShowFlag(d.ctx, d.sqlTx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	} else {
		reResultsByList = d.HhDb.GetRowByListWithCtxAndShowFlag(d.ctx, res.Sql.String(), d.writeHhDbLog, res.Args...)
	}
	//d.listResultsByList = reResultsByList
	d.isQuery = true
	d.err = reResultsByList.ExecState
	d.printLog(reResultsByList.ExecCost.String(), tx, res)
	if d.err != nil {
		return nil, d.err
	}
	if len(reResultsByList.ReData) <= 1 && d.emptyError != nil {
		d.err = d.emptyError
		return reResultsByList.ReData, d.err
	}
	return reResultsByList.ReData, d.err
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

// getSqlWrapper 获取sql构建器
func (d *Db) getSqlWrapper() SqlWrapperFace {
	if sqlBuildFun, ok := conf.DriveMap[d.sqlDrive]; ok {
		return sqlBuildFun(d.gs)
	}
	return d.gs
}

// getPrintSql 获取打印的sql
func (d *Db) getPrintSql(r *Result, lv LogLevel) string {
	if lv == ErrorLogLevel && !d.writeErrSql {
		return "none"
	}
	if d.writeCompSql {
		return r.CompSql()
	}
	return r.Sql.String()
}

// isMap 判断是否map
func (d *Db) isMap(ref reflect.Value) bool {
	if !ref.IsValid() {
		return false
	}
	for ref.Kind() == reflect.Ptr || ref.Kind() == reflect.Interface {
		if ref.IsNil() {
			return false
		}
		ref = ref.Elem()
	}

	return ref.Kind() == reflect.Map
}

// printLog 打印日志
func (d *Db) printLog(st string, txKey string, res *Result) {
	if d.writeLog {
		d.ctx = appendLogCallDepthCtx(d.ctx, 1)
		if d.err == nil {
			var affect = ""
			if !d.isQuery {
				affect = fmt.Sprintf("[Affect:%d] ", d.execResult.AffectRecs)
			}
			d.log.CtxDebugf(d.ctx, "[%s] %s%s%s", st, affect, txKey, d.getPrintSql(res, DebugLogLevel))
		} else {
			d.log.CtxErrorf(d.ctx, "Sql Err %s%s %v", txKey, d.getPrintSql(res, ErrorLogLevel), d.err)
		}
	}
}
