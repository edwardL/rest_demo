package gdb

import (
	"context"
	"database/sql"
	"errors"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"reflect"
)

// TOrmResult 统一返回结果
type TOrmResult[T any] struct {
	OrmResult
	Model *T
}

// DbT 泛型操作对象
type DbT[T any] struct {
	*Db
	tVal     any
	pageInfo *Page[T]
}

// Model 泛型实例
// T 必须为结构体
func Model[T any](ctx ...context.Context) *DbT[T] {
	var t = reflect.TypeOf((*T)(nil)).Elem()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var dbObj = New()
	if t.Kind() == reflect.Struct {
		var tb, _ = getTableNameRecursive(reflect.New(t))
		if tb != "" {
			dbObj = New(tb)
		}
	}
	var db = &DbT[T]{
		Db: dbObj,
	}
	var dbCtx = context.Background()
	if len(ctx) > 0 {
		dbCtx = ctx[0]
	}
	db.Ctx(dbCtx)
	return db
}

// Scan 解析到结构体 如果是数组 数组里必须是指针 []*Struct 兼容单结构体和数组
func (d *DbT[T]) Scan() error {
	return errors.New("不支持该方法")
}

// ScanAndCount 解析到结构体  数组里必须是指针 []*Struct
func (d *DbT[T]) ScanAndCount() error {
	return errors.New("不支持该方法")
}

// Select 解析到数组结构体 数组里必须是指针 []*Struct
func (d *DbT[T]) Select() (result []*T, err error) {
	err = d.Db.Select(&result).Error
	return result, err
}

// SelectPage 分页查询
func (d *DbT[T]) SelectPage() (result *Page[T], err error) {
	var total int64
	var pageData []*T
	err = d.Db.SelectAndCount(&pageData, &total).Error
	if err != nil {
		return nil, err
	}
	if d.pageInfo == nil {
		d.pageInfo = &Page[T]{}
	}
	d.pageInfo.PageData = pageData
	d.pageInfo.TotalNums = int(total)
	return d.pageInfo, err
}

// SelectAndCount 解析到数组结构体并统计数量 数组里必须是指针 []*Struct
func (d *DbT[T]) SelectAndCount() (result []*T, count int64, err error) {
	err = d.Db.LogCallDepth(1).SelectAndCount(&result, &count).Error
	return result, count, err
}

// One 解析到结构体 *Struct
func (d *DbT[T]) One() (result *T, err error) {
	err = d.Db.One(&result).Error
	return result, err
}

// Count 获取统计语句 f 为COUNT(f[0])
func (d *DbT[T]) Count(f ...string) (count int64, err error) {
	err = d.Db.LogCallDepth(1).Count(&count, f...).Error
	return count, err
}

// Exists 判断是否存在
func (d *DbT[T]) Exists() (bool, error) {
	return d.Db.LogCallDepth(1).Exists()
}

// Update 更新某个字段
func (d *DbT[T]) Update(column string, val any) (r TOrmResult[T]) {
	r.OrmResult = d.Db.LogCallDepth(1).Update(column, val)
	return r
}

// Updates 更新多个字段 data = map[string]any 或者 Struct
func (d *DbT[T]) Updates(data any, whereField ...[]string) (r TOrmResult[T]) {
	if dbInfo, ok := data.(*T); ok {
		r.Model = dbInfo
	}
	if dbInfo, ok := data.(T); ok {
		r.Model = &dbInfo
	}
	r.OrmResult = d.Db.Updates(data, whereField...)
	_ = conv.AssignId(&r.Model, r.OrmResult.GetLastInsertId())
	return r
}

// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
func (d *DbT[T]) UpdateIgnore(data any) (r TOrmResult[T]) {
	if dbInfo, ok := data.(*T); ok {
		r.Model = dbInfo
	}
	if dbInfo, ok := data.(T); ok {
		r.Model = &dbInfo
	}
	r.OrmResult = d.Db.LogCallDepth(1).UpdateIgnore(data)
	_ = conv.AssignId(&r.Model, r.OrmResult.GetLastInsertId())
	return r
}

// Save 存在更新，否则新增 data = map[string]any 或者 Struct
func (d *DbT[T]) Save(data any) (r TOrmResult[T]) {
	if dbInfo, ok := data.(*T); ok {
		r.Model = dbInfo
	}
	if dbInfo, ok := data.(T); ok {
		r.Model = &dbInfo
	}
	r.OrmResult = d.Db.Save(data)
	_ = conv.AssignId(&r.Model, r.OrmResult.GetLastInsertId())
	return r
}

// SaveInBatches 批量新增和更新
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *DbT[T]) SaveInBatches(data any, batchSize ...int) (r TOrmResult[T]) {
	r.OrmResult = d.Db.SaveInBatches(data, batchSize...)
	return r
}

// Delete 删除 DELETE {t} From table
func (d *DbT[T]) Delete(t ...string) (r TOrmResult[T]) {
	r.OrmResult = d.Db.LogCallDepth(1).Delete(t...)
	return r
}

// Create 新增数据
// data = map[string]any 或者 Struct
// data = *Db insert init ... select  从查询新增
func (d *DbT[T]) Create(data any) (r TOrmResult[T]) {
	if dbInfo, ok := data.(*T); ok {
		r.Model = dbInfo
	}
	if dbInfo, ok := data.(T); ok {
		r.Model = &dbInfo
	}
	r.OrmResult = d.Db.Create(data)
	_ = conv.AssignId(&r.Model, r.OrmResult.GetLastInsertId())
	return r
}

// CreateInBatches 批量创建
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *DbT[T]) CreateInBatches(data any, batchSize ...int) (r TOrmResult[T]) {
	r.OrmResult = d.Db.CreateInBatches(data, batchSize...)
	return r
}

// Replace 新增或替换数据
// data = map[string]any 或者 Struct
// data = *Db insert init ... select  从查询新增
func (d *DbT[T]) Replace(data any) (r TOrmResult[T]) {
	if dbInfo, ok := data.(*T); ok {
		r.Model = dbInfo
	}
	if dbInfo, ok := data.(T); ok {
		r.Model = &dbInfo
	}
	r.OrmResult = d.Db.Replace(data)
	_ = conv.AssignId(&r.Model, r.OrmResult.GetLastInsertId())
	return r
}

// ReplaceInBatches 批量创建或替换
// data = []map[string]any 或者 []Struct
// batchSize 批量大小 默认100
func (d *DbT[T]) ReplaceInBatches(data any, batchSize ...int) (r TOrmResult[T]) {
	r.OrmResult = d.Db.ReplaceInBatches(data, batchSize...)
	return r
}

// IgnoreDuplicate 更新条件
// 设置 field = true
// Duplicate 会忽略该字段
func (d *DbT[T]) IgnoreDuplicate(fieldIgnore map[string]bool) *DbT[T] {
	d.Db.IgnoreDuplicate(fieldIgnore)
	return d
}

// SetDuplicate 更新条件
// name = IfNull(VALUES(name),”)
func (d *DbT[T]) SetDuplicate(keyUpdate map[string]string) *DbT[T] {
	d.Db.SetDuplicate(keyUpdate)
	return d
}

// Joins .
// [query] ON xx = xxx
func (d *DbT[T]) Joins(query string, args ...any) *DbT[T] {
	d.Db.Joins(query, args...)
	return d
}

// Join .
// JOIN table ON xx = xxx
func (d *DbT[T]) Join(table, on string, args ...any) *DbT[T] {
	d.Db.Join(table, on, args...)
	return d
}

// InnerJoin .
// INNER JOIN table ON xx = xxx
func (d *DbT[T]) InnerJoin(table, on string, args ...any) *DbT[T] {
	d.Db.InnerJoin(table, on, args...)
	return d
}

// LeftJoin .
// LEFT JOIN table ON xx = xxx
func (d *DbT[T]) LeftJoin(table, on string, args ...any) *DbT[T] {
	d.Db.LeftJoin(table, on, args...)
	return d

}

// RightJoin .
// RIGHT JOIN table ON xx = xxx
func (d *DbT[T]) RightJoin(table, on string, args ...any) *DbT[T] {
	d.Db.RightJoin(table, on, args...)
	return d
}

// Where 基础查询条件
// db.Where("name = ?", "xxx")
// db.Where("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *DbT[T]) Where(query string, args ...any) *DbT[T] {
	d.Db.Where(query, args...)
	return d
}

// WhereOr 或查询条件
// db.WhereOr("name = ?", "xxx")
// db.WhereOr("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *DbT[T]) WhereOr(query string, args ...any) *DbT[T] {
	d.Db.WhereOr(query, args...)
	return d
}

// WhereCond 动态处理 OR AND
func (d *DbT[T]) WhereCond(cond Cond, query string, args ...any) *DbT[T] {
	d.Db.WhereCond(cond, query, args...)
	return d
}

// WhereReset 重置查询条件
func (d *DbT[T]) WhereReset() *DbT[T] {
	d.Db.WhereReset()
	return d
}

// WhereGroup 查询条件组会加括号 AND (id = ? AND name = ?)
func (d *DbT[T]) WhereGroup(wf WhereFunc) *DbT[T] {
	d.Db.WhereGroup(wf)
	return d
}

// WhereGroupOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *DbT[T]) WhereGroupOr(wf WhereFunc) *DbT[T] {
	d.Db.WhereGroupOr(wf)
	return d
}

// WhereBlock 查询条件组会加括号 AND (id = ? AND name = ?)
func (d *DbT[T]) WhereBlock(query string, args ...any) *DbT[T] {
	if d.err != nil {
		return d
	}
	d.Db.WhereBlock(query, args...)
	return d
}

// WhereBlockOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *DbT[T]) WhereBlockOr(query string, args ...any) *DbT[T] {
	if d.err != nil {
		return d
	}
	d.Db.WhereBlockOr(query, args...)
	return d
}

// WhereGroupCond 动态处理 AND OR
func (d *DbT[T]) WhereGroupCond(cond Cond, wf WhereFunc) *DbT[T] {
	d.Db.WhereGroupCond(cond, wf)
	return d
}

// Wrapper 查询构建器
// 会覆盖，请在最前面使用
func (d *DbT[T]) Wrapper(wrap *Wrapper) *DbT[T] {
	if wrap == nil {
		return d
	}
	d.Db.Wrapper(wrap)
	return d
}

// Table 表名 或者 子查询 每次table都会创建一个新的以支持事务
// db.Table("table_name")
// db.Table("(?) tb",db.Table("tb"))
// db.Table(*Struct{}) 模型结构体
func (d *DbT[T]) Table(table any, args ...any) *DbT[T] {
	d.Db.Table(table, args...)
	return d
}

// As 表别名
func (d *DbT[T]) As(alias string) *DbT[T] {
	d.Db.As(alias)
	return d
}

// Clone 复制一个实例
func (d *DbT[T]) Clone() *DbT[T] {
	var newDb = &DbT[T]{
		d.Db.Clone(),
		d.tVal,
		d.pageInfo,
	}
	return newDb
}

// Field 字段列表
// db.Field("id,ts","name"...)
func (d *DbT[T]) Field(f ...string) *DbT[T] {
	d.Db.Field(f...)
	return d
}

// Fields 字段
// 字符串 db.Fields("id,ts,? AS name","姓名")
// 切片 db.Fields([]string{"id","ts","? AS name"},"姓名")
func (d *DbT[T]) Fields(f any, args ...any) *DbT[T] {
	d.Db.Fields(f, args...)
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
func (d *DbT[T]) FieldsByData(f any, replaceField map[string]string, args ...any) *DbT[T] {
	d.Db.FieldsByData(f, replaceField, args...)
	return d
}

// FieldsByDataAlias 查询结构体字段 兼容gdb标签
// FieldsByData 方法 加上别名 tbAlias.field AS field
func (d *DbT[T]) FieldsByDataAlias(f any, replaceField map[string]string, tbAlias string, args ...any) *DbT[T] {
	d.Db.FieldsByDataAlias(f, replaceField, tbAlias, args...)
	return d
}

// OmitFields 新增和更新多字段排除的字段
// f[string | []string]
func (d *DbT[T]) OmitFields(f any) *DbT[T] {
	d.Db.OmitFields(f)
	return d
}

// ReplaceFields 新增和更新多字段替换字段
// fm {"field":"db_field"} 字段field更换为db_field
func (d *DbT[T]) ReplaceFields(fm map[string]string) *DbT[T] {
	d.Db.ReplaceFields(fm)
	return d
}

// Limit limit操作
// Limit("?,?", 1,2)
// Limit("?", 2)
// Limit("CASE WHEN type = ? THEN ? WHEN type = ? THEN ? ELSE ?", 1,2,3,4,5)
func (d *DbT[T]) Limit(limit string, args ...any) *DbT[T] {
	d.Db.Limit(limit, args...)
	return d
}

// Limits limit操作
// Limits(1,2)
// Limits(2)
func (d *DbT[T]) Limits(l int, l2 ...int) *DbT[T] {
	d.Db.Limits(l, l2...)
	return d
}

// Offset .
func (d *DbT[T]) Offset(offset int) *DbT[T] {
	d.Db.Offset(offset)
	return d
}

// Page .
func (d *DbT[T]) Page(page, pageSize int) *DbT[T] {
	d.Db.Page(page, pageSize)
	if d.pageInfo == nil {
		d.pageInfo = &Page[T]{
			CurrPage: page,
			PageNums: pageSize,
		}
	} else {
		d.pageInfo.PageNums = pageSize
		d.pageInfo.CurrPage = page
	}
	return d
}

// PageReset 重置分页
func (d *DbT[T]) PageReset() *DbT[T] {
	d.Db.PageReset()
	d.pageInfo = nil
	return d
}

// Group .
// f[string | []string]
func (d *DbT[T]) Group(f any, args ...any) *DbT[T] {
	d.Db.Group(f, args...)
	return d
}

// Groups .
// f string
func (d *DbT[T]) Groups(f ...string) *DbT[T] {
	d.Db.Groups(f...)
	return d
}

// Order .
// f[string | []string]
func (d *DbT[T]) Order(f any, args ...any) *DbT[T] {
	d.Db.Order(f, args...)
	return d
}

// Orders .
// f string
func (d *DbT[T]) Orders(f ...string) *DbT[T] {
	d.Db.Orders(f...)
	return d
}

// OrderByFilter 带过滤排序 处理前端排序字段
// OrderByFilter("id desc,aa asc",{"id":"new_field","aa",""}) = new_field desc,aa asc
// ignoreErr 是否忽略错误 false 会返回 xxx 字段不支持排序
func (d *DbT[T]) OrderByFilter(orderStr string, filter map[string]string, ignoreErr ...bool) *DbT[T] {
	d.Db.OrderByFilter(orderStr, filter, ignoreErr...)
	return d
}

// WhereSearchView 处理表单查询条件 使用视图字段过滤
// 视图推荐构建 视图字段配置 处理前端查询条件
// filterField 数组允许的查询字段
// filterField {"id":"new_field"} 不支持空val
func (d *DbT[T]) WhereSearchView(sD types.DataViewSearchInfo, sSItem []types.SearchItem, filterField map[string]string) *DbT[T] {
	d.Db.WhereSearchView(sD, sSItem, filterField)
	return d
}

// WhereSearch 处理表单查询条件 忽略视图配置
// 处理前端查询条件
// filterField 数组允许的查询字段
// ignoreErr 是否忽略错误 false 会返回 搜索条件错误,搜索字段不存在
func (d *DbT[T]) WhereSearch(sSItem []types.SearchItem, filterField map[string]string, ignoreErr ...bool) *DbT[T] {
	d.Db.WhereSearch(sSItem, filterField, ignoreErr...)
	return d
}

// WhereSearchParams 处理前端搜索查询条件
// 处理前端查询条件
func (d *DbT[T]) WhereSearchParams(sp *HhSearchParam) *DbT[T] {
	d.Db.WhereSearchParams(sp)
	if sp.CurrPage > 0 && sp.PageNums > 0 {
		d.Page(sp.CurrPage, sp.PageNums)
	}
	return d
}

// Union 合并查询
func (d *DbT[T]) Union(dbList ...*Db) *DbT[T] {
	d.Db.Union(dbList...)
	return d
}

// UnionAll 合并查询
func (d *DbT[T]) UnionAll(dbList ...*Db) *DbT[T] {
	d.Db.UnionAll(dbList...)
	return d
}

// Raw 原生sql查询
func (d *DbT[T]) Raw(query string, args ...any) *DbT[T] {
	d.Db.Raw(query, args...)
	return d
}

// ConvFieldsType 字段转换类型用于map查询
// .Map 或者Maps查询时 返回的map[string]any
// any 类型转换成指定类型
// 请不要修改第三个参数 避免数据竞争
func (d *DbT[T]) ConvFieldsType(f ConvVal) *DbT[T] {
	d.Db.ConvFieldsType(f)
	return d
}

// Ctx 设置上下文
func (d *DbT[T]) Ctx(ctx context.Context) *DbT[T] {
	d.Db.Ctx(setLogCallDepthCtx(ctx, getLogCallDepthCtx(nil)+1))
	return d
}

// WriteLog 是否打印日志
func (d *DbT[T]) WriteLog(flag bool, hhDbFlag ...bool) *DbT[T] {
	d.Db.WriteLog(flag, hhDbFlag...)
	return d
}

// WriteHhDbLog 是否打印框架层日志
func (d *DbT[T]) WriteHhDbLog(flag bool) *DbT[T] {
	d.Db.WriteHhDbLog(flag)
	return d
}

// WriteErrSql 是否打印错误sql
func (d *DbT[T]) WriteErrSql(flag bool) *DbT[T] {
	d.Db.WriteErrSql(flag)
	return d
}

// WriteCompSql 是否完整sql
func (d *DbT[T]) WriteCompSql(flag bool) *DbT[T] {
	d.Db.WriteCompSql(flag)
	return d
}

// GetGSql 获取sql构建器
func (d *DbT[T]) GetGSql() *Sql {
	return d.Db.GetGSql()
}

// Maps 查询map列表
func (d *DbT[T]) Maps() (result []map[string]any, err error) {
	return d.Db.LogCallDepth(1).Maps()
}

// Map 查询map
func (d *DbT[T]) Map() (res map[string]any, err error) {
	return d.Db.LogCallDepth(1).Map()
}

// TypeMaps 查询带类型的map列表
func (d *DbT[T]) TypeMaps(t map[string]conv.ValType) (result []map[string]any, err error) {
	return d.Db.LogCallDepth(1).TypeMaps(t)
}

// TypeMap 查询带类型的map
func (d *DbT[T]) TypeMap(t map[string]conv.ValType) (result map[string]any, err error) {
	return d.Db.LogCallDepth(1).TypeMap(t)
}

// OrderMaps 查询有序map列表
func (d *DbT[T]) OrderMaps() (result []*OrderedMap[string, any], err error) {
	return d.Db.LogCallDepth(1).OrderMaps()
}

// OrderMap 查询OrderMap
func (d *DbT[T]) OrderMap() (res *OrderedMap[string, any], err error) {
	return d.Db.LogCallDepth(1).OrderMap()
}

// Pluck 查询一列
// slice 切片 &[][string,int...]
func (d *DbT[T]) Pluck(field string) (slice []T, err error) {
	err = d.Db.LogCallDepth(1).Pluck(field, &slice).Error
	return slice, err
}

// Slices 查询切片列表
func (d *DbT[T]) Slices() (result [][]string, err error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.queryToSlices()
}

// Transaction 闭包事务
func (d *DbT[T]) Transaction(f func(db *DbTx) error) (err error) {
	return d.Db.LogCallDepth(1).Transaction(f)
}

// Tx 设置事务
func (d *DbT[T]) Tx(tx *sql.Tx) *DbT[T] {
	d.Db.Tx(tx)
	return d
}

// GetTx 获取事务对象
func (d *DbT[T]) GetTx() *sql.Tx {
	return d.Db.GetTx()
}

// LogLevel 设置当前操纵日志等级
func (d *DbT[T]) LogLevel(lv LogLevel) *DbT[T] {
	d.Db.LogLevel(lv)
	return d
}

// LogCallDepth 修改当前操纵日志深度
func (d *DbT[T]) LogCallDepth(callDepth int) *DbT[T] {
	d.Db.LogCallDepth(callDepth)
	return d
}

// EmptyError 未找到数据返回错误 默认返回空
func (d *DbT[T]) EmptyError(err ...error) *DbT[T] {
	d.Db.EmptyError(err...)
	return d
}

// CteQuery 虚拟表查询
// WITH
//
//	临时表1名称 [(列名1, 列名2, ...)] AS (SELECT/INSERT/UPDATE/DELETE 语句),
//	{tableAlias} AS ({query}),
//
// cte 默认 WITH
func (d *DbT[T]) CteQuery(query *CteQuery, cte ...string) *DbT[T] {
	if d.err != nil {
		return d
	}
	d.Db.CteQuery(query, cte...)
	return d
}

// Stream 流式查询
// sc = gdb.StreamCallback(func(d map[string]any|你的结构体) (err error, next bool) {})
func (d *DbT[T]) Stream(sc func(t T) (next bool, err error)) (r TOrmResult[T]) {
	r.OrmResult = d.Db.LogCallDepth(1).Stream(StreamCallback[T](sc))
	return r
}

// setLastInsertId .
func (d *DbT[T]) setLastInsertId(t *T, lastInsertId int64) {
	var ptrValue = reflect.ValueOf(t)
	if ptrValue.IsNil() {
		return
	}
	var structValue = ptrValue.Elem()
	var idField = structValue.FieldByName("Id")
	if !idField.IsValid() {
		return
	}
	if !idField.CanSet() {
		return
	}
	idField.SetInt(lastInsertId)
}
