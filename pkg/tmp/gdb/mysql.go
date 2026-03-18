package gdb

import (
	"context"
	"errors"
	"fmt"
	"nwgit.gzhhit.com/BD/hhitcommcode.git/utils/conv"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

// WhereSqlFunc 闭包where 给where 加 ()
type WhereSqlFunc func(gs *Sql) (err error)

// Sql 构建sql操作对象
type Sql struct {
	ctx        context.Context // ctx
	err        error           // 错误信息
	table      string          // 表名
	tableAlias string          // 表别名
	args       []any           // 参数
	fields     *GenCtrl        // 字段参数
	rawSql     *GenCtrl        // 原生sql参数
	cteQuery   *CteQuery       // cte复杂查询
	union      []*GenCtrl      // union参数
	whereCtrl  []*GenCtrl      // where参数
	joinCtrl   []*GenCtrl      // 连表参数
	limitCtrl  *GenCtrl        // 分页参数
	offsetCtrl *GenCtrl        // 分页参数
	group      *GenCtrl        // 分组参数
	order      *GenCtrl        // 排序参数
	// 处理参数
	dbConvInitPtr bool // 初始化指针
	timeLocation  *time.Location

	fieldReplace    map[string]string // 更新和新增字段名替换
	fieldSelect     map[string]string // 更新和新增指定的字段
	fieldOmit       map[string]string // 不更新和新增指定字段
	ignoreDuplicate map[string]bool   // 更新时，如果字段值存在，则更新
	setDuplicate    map[string]string // 更新时，自定义更新规则 name = IfNULL(name,'')
}

// NewSql 创建sql对象
// 不支持变量替换 只有第一个参数生效
func NewSql(table ...any) *Sql {
	gs := &Sql{
		fields:        NewGenCtrl(),        // 字段参数
		rawSql:        NewGenCtrl(),        // 原生sql参数
		cteQuery:      new(CteQuery),       // cte复杂查询
		union:         make([]*GenCtrl, 0), // union参数
		whereCtrl:     make([]*GenCtrl, 0), // where参数
		joinCtrl:      make([]*GenCtrl, 0), // 连表参数
		limitCtrl:     NewGenCtrl(),        // 分页参数
		offsetCtrl:    NewGenCtrl(),        // 分页参数
		group:         NewGenCtrl(),        // 分组参数
		order:         NewGenCtrl(),        // 排序参数
		dbConvInitPtr: conf.DbConvInitPtr,
		timeLocation:  conf.TimeLocation,
	}
	if len(table) > 0 {
		gs.Table(table[0])
	}
	return gs
}

// GenCtrl 构造条件
type GenCtrl struct {
	Query *strings.Builder
	Ct    string // and or
	Args  []any
}

// NewGenCtrl 获取构造条件
func NewGenCtrl() *GenCtrl {
	return &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
		Args:  make([]any, 0),
	}
}

// Err 获取错误
func (d *Sql) Err() error {
	return d.err
}

// Ctx 获取查询语句
func (d *Sql) Ctx(ctx context.Context) {
	d.ctx = ctx
}

// GetTable 获取表名
func (d *Sql) GetTable() string {
	return d.table
}

// Select 获取查询语句
func (d *Sql) Select() (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理字段
	if d.fields.Query.Len() == 0 {
		r.Sql.WriteString("SELECT * FROM")
	} else {
		r.Sql.WriteString("SELECT ")
		r.Sql.WriteString(d.fields.Query.String())
		r.Sql.WriteString(" FROM")
		r.Args = append(r.Args, d.fields.Args...)
	}

	// 处理表
	r.Sql.WriteString(" ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}
	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	}

	// 处理 GROUP BY
	if d.group.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.group.Query.String())
		if d.group.Args != nil {
			r.Args = append(r.Args, d.group.Args...)
		}
	}

	// 处理 ORDER BY
	if d.order.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.order.Query.String())
		if d.order.Args != nil {
			r.Args = append(r.Args, d.order.Args...)
		}
	}

	// 处理 LIMIT OFFSET
	if d.limitCtrl.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.limitCtrl.Query.String())
		if d.limitCtrl.Args != nil {
			r.Args = append(r.Args, d.limitCtrl.Args...)
		}
	}
	if d.offsetCtrl.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.offsetCtrl.Query.String())
		if d.offsetCtrl.Args != nil {
			r.Args = append(r.Args, d.offsetCtrl.Args...)
		}
	}

	// 处理 union
	if len(d.union) > 0 {
		for _, v := range d.union {
			r.Sql.WriteString(" ")
			r.Sql.WriteString(v.Ct)
			r.Sql.WriteString(" ")
			r.Sql.WriteString(v.Query.String())
			r.Args = append(r.Args, v.Args...)
		}
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Count 获取统计语句 f 为COUNT(f[0])
func (d *Sql) Count(f ...string) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var cf = "*"
	if len(f) > 0 {
		cf = f[0]
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("SELECT COUNT(")
	r.Sql.WriteString(cf)
	r.Sql.WriteString(") FROM")
	// 处理表
	r.Sql.WriteString(" " + d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}
	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	}

	// 处理 GROUP BY
	if d.group.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.group.Query.String())
		if d.group.Args != nil {
			r.Args = append(r.Args, d.group.Args...)
		}
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Exists 判断是否存在
func (d *Sql) Exists() (r *Result, err error) {
	var res *Result
	res, d.err = d.Select()
	if d.err != nil {
		return nil, d.err
	}
	var rawSql = res.Sql.String()
	res.Sql.Reset()
	res.Sql.WriteString("SELECT EXISTS (")
	res.Sql.WriteString(rawSql)
	res.Sql.WriteString(") AS is_exists")
	return res, nil
}

// Update 更新某个字段
func (d *Sql) Update(column string, val any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("UPDATE ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	var rfFlag = false
	if d.fieldReplace != nil {
		if rfField, rfOk := d.fieldReplace[column]; rfOk {
			rfFlag = true
			column = rfField
		}
	}
	if !rfFlag {
		if _, kfOk := mysqlKeywords[column]; kfOk {
			column = "`" + column + "`"
		}
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	// 处理更新字段
	if vr, ok := val.(RawBody); ok {
		r.Sql.WriteString(" SET ")
		r.Sql.WriteString(column)
		r.Sql.WriteString(" = ")
		r.Sql.WriteString(string(vr))
	} else {
		r.Sql.WriteString(" SET ")
		r.Sql.WriteString(column)
		r.Sql.WriteString(" = ?")
		r.Args = append(r.Args, val)
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql.Reset()
		r.Args = make([]any, 0)
		d.err = errors.New("缺少更新条件")
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Updates 更新多个字段 data = map[string]any 或者 Struct
func (d *Sql) Updates(data any, whereField ...[]string) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("UPDATE ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	// 处理更新字段
	var upField = ""
	var upFieldMap = map[string]string{}
	var m map[string]any
	var isStruct bool
	var igFn string
	var wv any
	var ok bool
	var fn string
	m, isStruct, d.err = toMap(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)
	for _, wfItem := range whereField {
		for _, wf := range wfItem {
			if wv, ok = m[wf]; ok {
				d.Where(wf+" = ?", wv)
				delete(m, wf)
			} else {
				d.err = errors.New("缺少更新条件:" + wf)
			}
		}
	}
	for _, igFn = range zeroValIgnoreField {
		if _, ok = m[igFn]; ok && conv.IsEmpty(m[igFn]) {
			delete(m, igFn)
		}
	}
	if m != nil {
		_, _, upFieldMap = d.getInsField([]map[string]any{m}, isStruct)
		for k, v := range m {
			if fn, ok = upFieldMap[k]; ok {
				if v != nil {
					if vr, isRaw := v.(RawBody); isRaw {
						upField += ", " + fn + " = " + string(vr)
					} else {
						upField += ", " + fn + " = ?"
						r.Args = append(r.Args, v)
					}
				}
			}
		}
		if upField != "" {
			upField = strings.TrimLeft(upField, ", ")
		}
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	if upField != "" {
		r.Sql.WriteString(" SET ")
		r.Sql.WriteString(upField)
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql.Reset()
		r.Args = make([]any, 0)
		d.err = errors.New("缺少更新条件")
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
func (d *Sql) UpdateIgnore(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("UPDATE IGNORE ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	// 处理更新字段
	var upField = ""
	var upFieldMap = map[string]string{}
	var m map[string]any
	var isStruct bool
	m, isStruct, d.err = toMap(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)
	if m != nil {
		_, _, upFieldMap = d.getInsField([]map[string]any{m}, isStruct)
		for k, v := range m {
			if fn, ok := upFieldMap[k]; ok {
				if v != nil {
					if vr, isRaw := v.(RawBody); isRaw {
						upField += ", " + fn + " = " + string(vr)
					} else {
						upField += ", " + fn + " = ?"
						r.Args = append(r.Args, v)
					}
				}
			}
		}
		if upField != "" {
			upField = strings.TrimLeft(upField, ", ")
		}
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	if upField != "" {
		r.Sql.WriteString(" SET ")
		r.Sql.WriteString(upField)
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Save 插入并更新
func (d *Sql) Save(data any) (r *Result, err error) {
	return d.SaveInBatches([]any{data})
}

// SaveInBatches 生成批量插入数据的sql语句
// data = map[string]any 或者 Struct
func (d *Sql) SaveInBatches(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	if d.ignoreDuplicate == nil {
		d.ignoreDuplicate = map[string]bool{}
	}
	if d.setDuplicate == nil {
		d.setDuplicate = make(map[string]string)
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	var insFileMap = map[string]string{}
	var mArr []map[string]any
	var isStruct bool
	mArr, isStruct, d.err = toMaps(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)
	if len(mArr) > 0 {
		_, _, insFileMap = d.getInsField(mArr, isStruct)
		//字段名
		var setDuplicate string
		var oldFieldName []string
		var fieldName []string
		var fieldIgnoreSql = new(strings.Builder)
		fieldIgnoreSql.WriteString(" ON DUPLICATE KEY UPDATE ")
		for k, v := range mArr {
			placeholder := make([]string, 0)
			if k == 0 {
				for k2, _ := range v {
					if fn, ok := insFileMap[k2]; ok {
						oldFieldName = append(oldFieldName, k2)
						//	处理主键
						if !d.ignoreDuplicate[k2] {
							if setDuplicate, ok = d.setDuplicate[fn]; ok {
								fieldIgnoreSql.WriteString(" ")
								fieldIgnoreSql.WriteString(fn)
								fieldIgnoreSql.WriteString(" = ")
								fieldIgnoreSql.WriteString(setDuplicate)
								fieldIgnoreSql.WriteString(",")
							} else {
								fieldIgnoreSql.WriteString(" ")
								fieldIgnoreSql.WriteString(fn)
								fieldIgnoreSql.WriteString(" = VALUES ( ")
								fieldIgnoreSql.WriteString(fn)
								fieldIgnoreSql.WriteString(" ),")
							}
						}
					}
				}
			}
			for _, v2 := range oldFieldName {
				if k == 0 {
					fieldName = append(fieldName, insFileMap[v2])
				}
				placeholder = append(placeholder, "?")
				r.Args = append(r.Args, v[v2])
			}
			//	处理 ?
			r.Sql.WriteString("(")
			r.Sql.WriteString(strings.Join(placeholder, ","))
			r.Sql.WriteString("),")
		}
		//处理尾部逗号
		var fieldIgnoreSqlStr = fieldIgnoreSql.String()
		fieldIgnoreSqlStr = fieldIgnoreSqlStr[:len(fieldIgnoreSqlStr)-1]
		var rSqlStr = r.Sql.String()
		rSqlStr = rSqlStr[:len(rSqlStr)-1]
		r.Sql.Reset()
		r.Sql.WriteString("INSERT INTO ")
		r.Sql.WriteString(d.table)
		if d.tableAlias != "" {
			r.Sql.WriteString(" AS ")
			r.Sql.WriteString(d.tableAlias)
		}
		r.Sql.WriteString(" (")
		r.Sql.WriteString(strings.Join(fieldName, ","))
		r.Sql.WriteString(") VALUES ")
		r.Sql.WriteString(rSqlStr)
		r.Sql.WriteString(fieldIgnoreSqlStr)
		r.Sql.WriteString(";")
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Delete 删除 DELETE {t} From table
func (d *Sql) Delete(t ...string) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var genSql = new(strings.Builder)
	if len(t) > 0 {
		genSql.WriteString("DELETE ")
		genSql.WriteString(strings.Join(t, ","))
		genSql.WriteString(" FROM ")
		genSql.WriteString(d.table)
		if d.tableAlias != "" {
			genSql.WriteString(" AS ")
			genSql.WriteString(d.tableAlias)
		}
	} else {
		genSql.WriteString("DELETE FROM ")
		genSql.WriteString(d.table)
		if d.tableAlias != "" {
			genSql.WriteString(" AS ")
			genSql.WriteString(d.tableAlias)
		}
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString(genSql.String())

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	var query string
	var args []any

	// 处理连表
	query, args = d.genJoin()
	if query != "" {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(query)
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql.WriteString(" WHERE ")
			r.Sql.WriteString(query)
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql.Reset()
		r.Args = make([]any, 0)
		d.err = errors.New("缺少删除条件")
	}

	// 处理limit
	// 处理 LIMIT OFFSET
	if d.limitCtrl.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.limitCtrl.Query.String())
		if d.limitCtrl.Args != nil {
			r.Args = append(r.Args, d.limitCtrl.Args...)
		}
	}
	if d.offsetCtrl.Query.Len() != 0 {
		r.Sql.WriteString(" ")
		r.Sql.WriteString(d.offsetCtrl.Query.String())
		if d.offsetCtrl.Args != nil {
			r.Args = append(r.Args, d.offsetCtrl.Args...)
		}
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Create 创建 data = map[string]any 或者 Struct
func (d *Sql) Create(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("INSERT INTO ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	// 判断是否子查询 处理 INSERT INTO Table (id, ts) SELECT id, ts FROM Table2;
	if childSql, ok := data.(*Sql); ok {
		var childRes *Result
		childRes, err = childSql.Select()
		if err != nil {
			d.err = err
			return nil, d.err
		}
		// 处理字段
		if d.fields.Query.Len() != 0 {
			r.Sql.WriteString(" (")
			r.Sql.WriteString(d.fields.Query.String())
			r.Sql.WriteString(")")
			r.Args = append(r.Args, d.fields.Args...)
		}
		r.Sql.WriteString(" ")
		r.Sql.WriteString(childRes.Sql.String())
		r.Args = append(r.Args, childRes.Args...)
		return r, d.err
	}

	var insField = ""
	var insFieldMap = map[string]string{}
	var insArgs = make([]any, 0)
	var m map[string]any
	var isStruct bool
	m, isStruct, d.err = toMap(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)

	if m != nil {
		_, _, insFieldMap = d.getInsField([]map[string]any{m}, isStruct)
		for k, v := range m {
			if fn, ok := insFieldMap[k]; ok {
				insField += ", " + fn
				insArgs = append(insArgs, v)
			}
		}
		if insField != "" {
			insField = strings.TrimLeft(insField, ", ")
		}
	}
	if insField != "" {
		r.Sql.WriteString(" ( ")
		r.Sql.WriteString(insField)
		r.Sql.WriteString(" )")
		var val string
		for _, v := range insArgs {
			if strV, aOk := v.(RawBody); aOk {
				val += ", " + string(strV)
			} else {
				val += ", ?"
				r.Args = append(r.Args, v)
			}
		}
		if val != "" {
			val = strings.TrimLeft(val, ", ")
		}
		r.Sql.WriteString(" VALUES ( ")
		r.Sql.WriteString(val)
		r.Sql.WriteString(" )")
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// CreateInBatches 批量创建
// data = []map[string]any 或者 []Struct
func (d *Sql) CreateInBatches(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("INSERT INTO ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	var insField = ""
	var insFileArr = make([]string, 0)
	var insArgs = make([][]any, 0)
	var mArr []map[string]any
	var isStruct bool
	mArr, isStruct, d.err = toMaps(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)
	if len(mArr) > 0 {
		insField, insFileArr, _ = d.getInsField(mArr, isStruct)
		for _, m := range mArr {
			args := make([]any, 0)
			for _, key := range insFileArr {
				if v, ok := m[key]; ok {
					args = append(args, v)
				} else {
					args = append(args, "")
				}
			}
			if insField != "" {
				insField = strings.TrimLeft(insField, ", ")
				insArgs = append(insArgs, args)
			}
		}
	}

	if insField != "" {
		r.Sql.WriteString(" ( ")
		r.Sql.WriteString(insField)
		r.Sql.WriteString(" )")
		var valList string
		for _, v := range insArgs {
			var val string
			for _, vv := range v {
				if strV, aOk := vv.(RawBody); aOk {
					val += ", " + string(strV)
				} else {
					val += ", ?"
					r.Args = append(r.Args, vv)
				}
			}
			if val != "" {
				val = strings.TrimLeft(val, ", ")
			}
			valList += ", (" + val + ")"
		}
		if valList != "" {
			valList = strings.TrimLeft(valList, ", ")
		}
		r.Sql.WriteString(" VALUES ")
		r.Sql.WriteString(valList)
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// Replace 创建或替换 data = map[string]any 或者 Struct
func (d *Sql) Replace(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("REPLACE INTO ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	// 判断是否子查询 处理 INSERT INTO Table (id, ts) SELECT id, ts FROM Table2;
	if childSql, ok := data.(*Sql); ok {
		var childRes *Result
		childRes, err = childSql.Select()
		if err != nil {
			d.err = err
			return nil, d.err
		}
		// 处理字段
		if d.fields.Query.Len() != 0 {
			r.Sql.WriteString(" (")
			r.Sql.WriteString(d.fields.Query.String())
			r.Sql.WriteString(")")
			r.Args = append(r.Args, d.fields.Args...)
		}
		r.Sql.WriteString(" ")
		r.Sql.WriteString(childRes.Sql.String())
		r.Args = append(r.Args, childRes.Args...)
		return r, d.err
	}

	var insField = ""
	var insFieldMap = map[string]string{}
	var insArgs = make([]any, 0)
	var m map[string]any
	var isStruct bool
	m, isStruct, d.err = toMap(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)

	if m != nil {
		_, _, insFieldMap = d.getInsField([]map[string]any{m}, isStruct)
		for k, v := range m {
			if fn, ok := insFieldMap[k]; ok {
				insField += ", " + fn
				insArgs = append(insArgs, v)
			}
		}
		if insField != "" {
			insField = strings.TrimLeft(insField, ", ")
		}
	}
	if insField != "" {
		r.Sql.WriteString(" ( ")
		r.Sql.WriteString(insField)
		r.Sql.WriteString(" )")
		var val string
		for _, v := range insArgs {
			if strV, aOk := v.(RawBody); aOk {
				val += ", " + string(strV)
			} else {
				val += ", ?"
				r.Args = append(r.Args, v)
			}
		}
		if val != "" {
			val = strings.TrimLeft(val, ", ")
		}
		r.Sql.WriteString(" VALUES ( ")
		r.Sql.WriteString(val)
		r.Sql.WriteString(" )")
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// ReplaceInBatches 批量创建或替换
// data = []map[string]any 或者 []Struct
func (d *Sql) ReplaceInBatches(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  new(strings.Builder),
		Args: make([]any, 0),
	}

	if d.rawSql.Query.Len() != 0 {
		r.Sql.Reset()
		r.Sql.WriteString(d.rawSql.Query.String())
		r.Args = d.rawSql.Args
		return r, d.err
	}

	r.Sql.WriteString("REPLACE INTO ")
	r.Sql.WriteString(d.table)
	if d.tableAlias != "" {
		r.Sql.WriteString(" AS ")
		r.Sql.WriteString(d.tableAlias)
	}

	var insField = ""
	var insFileArr = make([]string, 0)
	var insArgs = make([][]any, 0)
	var mArr []map[string]any
	var isStruct bool
	mArr, isStruct, d.err = toMaps(appendLogCallDepthCtx(d.ctx, 2), data, d.dbConvInitPtr)
	if len(mArr) > 0 {
		insField, insFileArr, _ = d.getInsField(mArr, isStruct)
		for _, m := range mArr {
			args := make([]any, 0)
			for _, key := range insFileArr {
				if v, ok := m[key]; ok {
					args = append(args, v)
				} else {
					args = append(args, "")
				}
			}
			if insField != "" {
				insField = strings.TrimLeft(insField, ", ")
				insArgs = append(insArgs, args)
			}
		}
	}

	if insField != "" {
		r.Sql.WriteString(" ( ")
		r.Sql.WriteString(insField)
		r.Sql.WriteString(" )")
		var valList string
		for _, v := range insArgs {
			var val string
			for _, vv := range v {
				if strV, aOk := vv.(RawBody); aOk {
					val += ", " + string(strV)
				} else {
					val += ", ?"
					r.Args = append(r.Args, vv)
				}
			}
			if val != "" {
				val = strings.TrimLeft(val, ", ")
			}
			valList += ", (" + val + ")"
		}
		if valList != "" {
			valList = strings.TrimLeft(valList, ", ")
		}
		r.Sql.WriteString(" VALUES ")
		r.Sql.WriteString(valList)
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// IgnoreDuplicate 更新条件
// 设置 field = true
// Duplicate 会忽略该字段
func (d *Sql) IgnoreDuplicate(fieldIgnore map[string]bool) *Sql {
	d.ignoreDuplicate = fieldIgnore
	return d
}

// SetDuplicate 更新条件
// name = IfNull(VALUES(name),”)
func (d *Sql) SetDuplicate(keyUpdate map[string]string) *Sql {
	d.setDuplicate = keyUpdate
	return d
}

// getInsField 获取允许新增的字段
func (d *Sql) getInsField(mArr []map[string]any, isStruct bool) (string, []string, map[string]string) {
	var insField = ""
	var fsLen = len(d.fieldSelect)
	var ofLen = len(d.fieldOmit)
	var rfLen = len(d.fieldReplace)
	var insFileArr = make([]string, 0)
	var insFileMap = map[string]string{}
	for f, v := range mArr[0] {
		if InArr(f, zeroValIgnoreField) && conv.IsEmpty(v) {
			continue
		}
		var mapKey = f
		var fOk = true
		var ofOk = false
		var rfOk = false
		var rfField = ""
		var kfOk = false
		if fsLen > 0 {
			_, fOk = d.fieldSelect[f]
		}
		if ofLen > 0 {
			_, ofOk = d.fieldOmit[f]
		}
		if rfLen > 0 {
			rfField, rfOk = d.fieldReplace[f]
		}
		_, kfOk = mysqlKeywords[f]
		if fOk && !ofOk {
			if rfOk {
				f = rfField
			} else {
				if kfOk || isStruct {
					f = "`" + f + "`"
				}
			}
			insField += ", " + f
			insFileArr = append(insFileArr, mapKey)
			insFileMap[mapKey] = f
		}
	}
	return insField, insFileArr, insFileMap
}

// Joins .
// [query] ON xx = xxx
func (d *Sql) Joins(query string, args ...any) *Sql {
	if d.err != nil {
		return d
	}
	if d.ppNum(query) != len(args) {
		d.err = errors.New("[ " + query + " ] 参数和占位符不匹配")
		return d
	}
	query, args, d.err = d.genQuery(query, args)
	if d.err != nil {
		return d
	}
	var genCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Args:  args,
	}
	genCtrl.Query.WriteString(query)
	d.joinCtrl = append(d.joinCtrl, genCtrl)
	return d
}

// Join .
// JOIN table ON xx = xxx
func (d *Sql) Join(table, on string, args ...any) *Sql {
	if d.err != nil {
		return d
	}
	var query = "JOIN " + table
	if on != "" {
		query += " ON " + on
	}
	return d.Joins(query, args...)
}

// InnerJoin .
// INNER JOIN table ON xx = xxx
func (d *Sql) InnerJoin(table, on string, args ...any) *Sql {
	if d.err != nil {
		return d
	}
	var query = "INNER JOIN " + table
	if on != "" {
		query += " ON " + on
	}
	return d.Joins(query, args...)
}

// LeftJoin .
// LEFT JOIN table ON xx = xxx
func (d *Sql) LeftJoin(table, on string, args ...any) *Sql {
	if d.err != nil {
		return d
	}
	var query = "LEFT JOIN " + table
	if on != "" {
		query += " ON " + on
	}
	return d.Joins(query, args...)
}

// RightJoin .
// RIGHT JOIN table ON xx = xxx
func (d *Sql) RightJoin(table, on string, args ...any) *Sql {
	if d.err != nil {
		return d
	}
	var query = "RIGHT JOIN " + table
	if on != "" {
		query += " ON " + on
	}
	return d.Joins(query, args...)
}

// Where 基础查询条件
// db.Where("name = ?", "xxx")
// db.Where("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *Sql) Where(query string, args ...any) (gs *Sql) {
	return d.appendWhere(query, CondAnd, args...)
}

// WhereOr 或查询条件
// db.WhereOr("name = ?", "xxx")
// db.WhereOr("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *Sql) WhereOr(query string, args ...any) (gs *Sql) {
	return d.appendWhere(query, CondOr, args...)
}

// WhereReset 重置查询条件
func (d *Sql) WhereReset() (gs *Sql) {
	d.whereCtrl = make([]*GenCtrl, 0)
	return d
}

// appendWhere 构建查询
func (d *Sql) appendWhere(query string, ct string, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	if query == "" {
		return d
	}
	if d.ppNum(query) != len(args) {
		d.err = errors.New("[ " + query + " ] 参数和占位符不匹配")
		return d
	}
	if len(d.whereCtrl) <= 0 {
		ct = ""
		d.whereCtrl = make([]*GenCtrl, 0)
	}

	query, args, d.err = d.genQuery(query, args)
	if d.err != nil {
		return d
	}
	var genCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    ct,
		Args:  args,
	}
	genCtrl.Query.WriteString(query)
	d.whereCtrl = append(d.whereCtrl, genCtrl)
	return d
}

// WhereGroup 查询条件组 AND (id = ? AND name = ?)
func (d *Sql) WhereGroup(wf WhereSqlFunc) (gs *Sql) {
	var db = NewSql()
	d.err = wf(db)
	if d.err != nil {
		return d
	}
	return d.appendWhereGroup(db, CondAnd)
}

// WhereGroupOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *Sql) WhereGroupOr(wf WhereSqlFunc) (gs *Sql) {
	var db = NewSql()
	d.err = wf(db)
	if d.err != nil {
		return d
	}
	return d.appendWhereGroup(db, CondOr)
}

// appendWhereGroup 构建查询条件组
func (d *Sql) appendWhereGroup(db *Sql, ct string) (gs *Sql) {
	if d.err != nil {
		return d
	}
	if len(d.whereCtrl) <= 0 {
		ct = ""
		d.whereCtrl = make([]*GenCtrl, 0)
	}
	var whereStr, args = db.genWhere()
	if whereStr == "" {
		return d
	}
	d.err = db.err
	var genCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    ct,
		Args:  args,
	}
	genCtrl.Query.WriteString("( ")
	genCtrl.Query.WriteString(whereStr)
	genCtrl.Query.WriteString(" )")
	d.whereCtrl = append(d.whereCtrl, genCtrl)
	return d
}

// Table 表名 或者 子查询
// db.Table("table_name")
// db.Table("(?) tb",db.Table("tb"))
// db.Table(*Struct{}) 模型结构体
func (d *Sql) Table(tb any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var table = GetTableName(tb)
	if table == "" {
		d.err = errors.New("未解析到表名")
		return d
	}
	if d.ppNum(table) != len(args) {
		d.err = errors.New("[ " + table + " ] 参数和占位符不匹配")
		return d
	}
	d.table = table
	if args != nil {
		d.table, args, d.err = d.genQuery(d.table, args)
		if d.err != nil {
			return d
		}
		for k, v := range args {
			if db, ok := v.(*Sql); ok {
				r, e := db.Select()
				if e != nil {
					d.err = e
					return d
				}
				d.table, d.err = replaceIndex(d.table, k+1, r.Sql.String())
				d.args = append(d.args, r.Args...)
			} else {
				d.args = append(d.args, v)
			}
		}
	}
	return d
}

// As 表别名
func (d *Sql) As(alias string) (gs *Sql) {
	d.tableAlias = alias
	return d
}

// Clone 复制一个实例
func (d *Sql) Clone() (gs *Sql) {
	var newSql = &Sql{
		err:             d.err,
		table:           d.table,
		args:            d.args,
		fieldSelect:     d.fieldSelect,
		fieldOmit:       d.fieldOmit,
		fieldReplace:    d.fieldReplace,
		ignoreDuplicate: d.ignoreDuplicate,
		setDuplicate:    d.setDuplicate,
	}
	newSql.fields = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.fields.Ct,
	}
	newSql.fields.Query.WriteString(d.fields.Query.String())
	newSql.fields.Args = make([]any, len(d.fields.Args), cap(d.fields.Args))
	copy(newSql.fields.Args, d.fields.Args)

	newSql.rawSql = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.rawSql.Ct,
	}
	newSql.rawSql.Query.WriteString(d.rawSql.Query.String())
	newSql.rawSql.Args = make([]any, len(d.rawSql.Args), cap(d.rawSql.Args))
	copy(newSql.rawSql.Args, d.rawSql.Args)

	newSql.cteQuery = &CteQuery{
		Cte: d.cteQuery.Cte,
		err: d.cteQuery.err,
	}
	newSql.cteQuery.VirtualTable = make([]*VirtualTable, len(d.cteQuery.VirtualTable), cap(d.cteQuery.VirtualTable))
	copy(newSql.cteQuery.VirtualTable, d.cteQuery.VirtualTable)

	newSql.union = make([]*GenCtrl, len(d.union), cap(d.union))
	copy(newSql.union, d.union)

	newSql.whereCtrl = make([]*GenCtrl, len(d.whereCtrl), cap(d.whereCtrl))
	copy(newSql.whereCtrl, d.whereCtrl)

	newSql.joinCtrl = make([]*GenCtrl, len(d.joinCtrl), cap(d.joinCtrl))
	copy(newSql.joinCtrl, d.joinCtrl)

	newSql.limitCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.limitCtrl.Ct,
	}
	newSql.limitCtrl.Query.WriteString(d.limitCtrl.Query.String())
	newSql.limitCtrl.Args = make([]any, len(d.limitCtrl.Args), cap(d.limitCtrl.Args))
	copy(newSql.limitCtrl.Args, d.limitCtrl.Args)

	newSql.offsetCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.offsetCtrl.Ct,
	}
	newSql.offsetCtrl.Query.WriteString(d.offsetCtrl.Query.String())
	newSql.offsetCtrl.Args = make([]any, len(d.offsetCtrl.Args), cap(d.offsetCtrl.Args))
	copy(newSql.offsetCtrl.Args, d.offsetCtrl.Args)

	newSql.group = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.rawSql.Ct,
	}
	newSql.group.Query.WriteString(d.group.Query.String())
	newSql.group.Args = make([]any, len(d.group.Args), cap(d.group.Args))
	copy(newSql.group.Args, d.group.Args)

	newSql.order = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    d.order.Ct,
	}
	newSql.order.Query.WriteString(d.order.Query.String())
	newSql.order.Args = make([]any, len(d.order.Args), cap(d.order.Args))
	copy(newSql.order.Args, d.order.Args)
	return newSql
}

// Fields 字段
// 字符串 db.Fields("id,ts,? AS name","姓名")
// 切片 db.Fields([]string{"id","ts","? AS name"},"姓名")
func (d *Sql) Fields(f any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	if d.fields.Args == nil {
		d.fields.Args = make([]any, 0)
	}
	d.fieldSelect = map[string]string{}
	switch field := f.(type) {
	case string:
		d.fields.Query.Reset()
		d.fields.Query.WriteString(field)
		fieldArr := strings.Split(strings.ReplaceAll(field, " ", ""), ",")
		for _, v := range fieldArr {
			d.fieldSelect[v] = v
		}
	case []string:
		if field != nil {
			newF := ""
			for _, v := range field {
				d.fieldSelect[v] = v
				newF += ", " + v
			}
			if newF != "" {
				d.fields.Query.Reset()
				d.fields.Query.WriteString(strings.TrimLeft(newF, ", "))
			}
		}
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(d.fields.Query.String()) != len(args) {
		d.err = errors.New("[ " + d.fields.Query.String() + " ] 参数和占位符不匹配")
		return d
	}
	var query string
	query, d.fields.Args, d.err = d.genQuery(d.fields.Query.String(), args)
	d.fields.Query.Reset()
	d.fields.Query.WriteString(query)
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
func (d *Sql) FieldsByData(f any, replaceField map[string]string, args ...any) (gs *Sql) {
	if d.err != nil || f == nil {
		return d
	}
	if replaceField == nil {
		replaceField = make(map[string]string)
	}
	if args == nil {
		args = make([]any, 0)
	}
	inpFieldArr := make([]string, 0)
	if fd, ok := f.(map[string]any); ok {
		for k, _ := range fd {
			inpFieldArr = append(inpFieldArr, k)
		}
	} else {
		inpFieldArr = StructDbField(f)
	}
	fieldArr := make([]string, 0)
	for _, fn := range inpFieldArr {
		if nf, ok := replaceField[fn]; ok {
			if nf != "" {
				fieldArr = append(fieldArr, nf)
			}
		} else {
			if _, kfOk := mysqlKeywords[fn]; kfOk {
				fn = "`" + fn + "`"
			}
			fieldArr = append(fieldArr, fn)
		}
	}
	return d.Fields(fieldArr, args...)
}

// FieldsByDataAlias 查询结构体字段 兼容gdb标签
// FieldsByData 方法 加上别名 tbAlias.field AS field
func (d *Sql) FieldsByDataAlias(f any, replaceField map[string]string, tbAlias string, args ...any) (gs *Sql) {
	if d.err != nil || f == nil {
		return d
	}
	if replaceField == nil {
		replaceField = make(map[string]string)
	}
	if args == nil {
		args = make([]any, 0)
	}
	inpFieldArr := make([]string, 0)
	if fd, ok := f.(map[string]any); ok {
		for k, _ := range fd {
			inpFieldArr = append(inpFieldArr, k)
		}
	} else {
		inpFieldArr = StructDbField(f)
	}
	fieldArr := make([]string, 0)
	for _, fn := range inpFieldArr {
		if nf, ok := replaceField[fn]; ok {
			if nf != "" {
				if tbAlias != "" {
					nf = tbAlias + "." + nf + " AS " + nf
				}
				fieldArr = append(fieldArr, nf)
			}
		} else {
			if _, kfOk := mysqlKeywords[fn]; kfOk {
				fn = "`" + fn + "`"
			}
			fn = tbAlias + "." + fn + " AS " + fn
			fieldArr = append(fieldArr, fn)
		}
	}
	return d.Fields(fieldArr, args...)
}

// OmitFields 新增和更新多字段排除的字段
// f[string | []string]
func (d *Sql) OmitFields(f any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.fieldOmit = map[string]string{}
	switch field := f.(type) {
	case string:
		fieldArr := strings.Split(strings.ReplaceAll(field, " ", ""), ",")
		for _, v := range fieldArr {
			d.fieldOmit[v] = v
		}
	case []string:
		if field != nil {
			for _, v := range field {
				d.fieldOmit[v] = v
			}
		}
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	return d
}

// ReplaceFields 替换字段
// fm {"field":"db_field"} 字段field更换为db_field
func (d *Sql) ReplaceFields(fm map[string]string) (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.fieldReplace = fm
	return d
}

// Limit limit操作
// Limit("?,?", 1,2)
// Limit("?", 2)
func (d *Sql) Limit(limit string, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	if d.ppNum(limit) != len(args) {
		d.err = errors.New("[ " + limit + " ] 占位符数量不匹配")
		return d
	}
	d.limitCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
	}
	var query string
	query, d.limitCtrl.Args, d.err = d.genQuery("LIMIT "+limit, args)
	d.limitCtrl.Query.WriteString(query)
	return d
}

// Offset .
func (d *Sql) Offset(offset int) (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.offsetCtrl = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
		Args:  []any{offset},
	}
	d.offsetCtrl.Query.WriteString("OFFSET ?")
	return d
}

// Page .
func (d *Sql) Page(page, pageSize int) (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.Limit("?", pageSize).Offset((page - 1) * pageSize)
	return d
}

// PageReset 重置分页
func (d *Sql) PageReset() (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.limitCtrl = &GenCtrl{
		Query: new(strings.Builder),
	}
	d.offsetCtrl = &GenCtrl{
		Query: new(strings.Builder),
	}
	return d
}

// Group .
// f[string | []string]
func (d *Sql) Group(f any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var groupField = ""
	switch field := f.(type) {
	case string:
		groupField = field
	case []string:
		groupField = strings.Join(field, ", ")
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(groupField) != len(args) {
		d.err = errors.New("[ " + groupField + " ] 参数和占位符不匹配")
		return d
	}
	d.group = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
		Args:  make([]any, 0),
	}
	var query string
	query, d.group.Args, d.err = d.genQuery("GROUP BY "+groupField, args)
	d.group.Query.WriteString(query)
	return d
}

// Order .
// f[string | []string]
func (d *Sql) Order(f any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var orderField = ""
	switch field := f.(type) {
	case string:
		orderField = field
	case []string:
		orderField = strings.Join(field, ", ")
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(orderField) != len(args) {
		d.err = errors.New("[ " + orderField + " ] 参数和占位符不匹配")
		return d
	}
	if !strings.Contains(orderField, "?") {
		orderField, _, d.err = ValidateOrderParam(orderField, nil)
		if d.err != nil {
			return d
		}
	}
	d.order = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
		Args:  make([]any, 0),
	}
	var query string
	query, d.order.Args, d.err = d.genQuery("ORDER BY "+orderField, args)
	d.order.Query.WriteString(query)
	return d
}

// OrderByFilter 带过滤排序 处理前端排序字段
// OrderByFilter("id desc,aa asc",{"id":"new_field","aa",""}) = new_field desc,aa asc
// ignoreErr 是否忽略错误 false 会返回 xxx 字段不支持排序
func (d *Sql) OrderByFilter(orderStr string, filter map[string]string, ignoreErr ...bool) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var igErr = true // 默认忽略错误
	if len(ignoreErr) > 0 {
		igErr = ignoreErr[0]
	}
	var illegalField []string
	orderStr, illegalField, d.err = ValidateOrderParam(orderStr, filter)
	if d.err != nil {
		return d
	}
	if !igErr && len(illegalField) > 0 {
		d.err = errors.New("字段：" + strings.Join(illegalField, ",") + " 不能用于排序")
		return d
	}
	d.order = &GenCtrl{
		Query: new(strings.Builder),
		Ct:    "",
		Args:  make([]any, 0),
	}
	d.order.Query.WriteString("ORDER BY ")
	d.order.Query.WriteString(orderStr)
	return d
}

// WhereSearchView 处理表单查询条件
// 视图推荐构建 视图字段配置 处理前端查询条件
// filterField 数组允许的查询字段
// filterField {"id":"new_field"} 不支持空val
func (d *Sql) WhereSearchView(sD types.DataViewSearchInfo, sSItem []types.SearchItem, filterField map[string]string) (gs *Sql) {
	if d.err != nil {
		return d
	}
	newDb := NewSql()
	var viewFileMap = make(map[string]types.DataViewField)
	for _, v := range sD.FieldList {
		viewFileMap[v.FieldAlias] = v
	}

	var err error
	for _, sItem := range sSItem {
		var query = ""
		var op = ""
		var whereVal []any
		if v1, yes := viewFileMap[sItem.FieldName]; yes {
			query = v1.FieldAlias
			if filterField != nil {
				q, ok := filterField[sItem.FieldName]
				if !ok {
					d.err = errors.New("搜索条件错误,搜索字段不存在")
					return d
				}
				query = q
			}
			op, whereVal, err = d.getSearchWhere(sItem.FieldCmpOp, sItem.FieldValue)
			if err != nil {
				d.err = err
				return d
			}
		}
		if op != "" {
			newDb.Where(query+" "+op, whereVal...)
		}
	}
	whereStr, args := newDb.genWhere()
	d.err = newDb.err
	d.appendWhere(whereStr, "AND", args...)
	return d
}

// WhereSearch 处理表单查询条件 忽略视图配置
func (d *Sql) WhereSearch(sSItem []types.SearchItem, filterField map[string]string, ignoreErr ...bool) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var igErr = false // 默认不忽略错误
	if len(ignoreErr) > 0 {
		igErr = ignoreErr[0]
	}
	var err error
	var op = ""
	newDb := NewSql()
	for _, sItem := range sSItem {
		var query string = sItem.FieldName
		if filterField != nil {
			q, ok := filterField[sItem.FieldName]
			if !ok && !igErr {
				d.err = errors.New("搜索条件错误,搜索字段不存在")
				return d
			} else {
				if q != "" {
					query = q
				}
			}
		}
		if query != "" {
			if strings.ToUpper(sItem.FieldCmpOp) == WhereOpFindInSet {
				newDb.Where("FIND_IN_SET(?, "+sItem.FieldName+")", sItem.FieldValue)
			} else {
				var whereVal []any
				op, whereVal, err = d.getSearchWhere(sItem.FieldCmpOp, sItem.FieldValue)
				if err != nil {
					d.err = err
					return d
				}
				if op != "" && query != FieldRemove {
					newDb.Where(query+" "+op, whereVal...)
				}
			}
		}
	}
	whereStr, args := newDb.genWhere()
	d.err = newDb.err
	d.appendWhere(whereStr, "AND", args...)
	return d
}

// Union 合并查询
func (d *Sql) Union(gsList ...*Sql) (gs *Sql) {
	return d.unionGen("UNION", gsList...)
}

// UnionAll 合并查询
func (d *Sql) UnionAll(gsList ...*Sql) (gs *Sql) {
	return d.unionGen("UNION ALL", gsList...)
}

// unionGen 合并查询
func (d *Sql) unionGen(unionType string, txList ...*Sql) (gs *Sql) {
	if d.err != nil {
		return d
	}
	var res *Result
	var query string
	if len(txList) > 0 {
		for _, q := range txList {
			res, d.err = q.Select()
			if d.err != nil {
				return d
			}
			var unionInfo = &GenCtrl{
				Query: new(strings.Builder),
				Ct:    unionType,
				Args:  nil,
			}
			query, unionInfo.Args, d.err = d.genQuery(res.Sql.String(), res.Args)
			if d.err != nil {
				return d
			}
			unionInfo.Query.WriteString(query)
			d.union = append(d.union, unionInfo)
		}
	}
	return d
}

// Raw 原生查询
func (d *Sql) Raw(query string, args ...any) (gs *Sql) {
	query, d.rawSql.Args, d.err = d.genQuery(strings.TrimLeft(query, " \t\n\r"), args)
	if d.err != nil {
		return d
	}
	d.rawSql.Query.Reset()
	d.rawSql.Query.WriteString(query)
	return d
}

// RawExec 原生更新
func (d *Sql) RawExec(execSql string, args ...any) (r *Result, err error) {
	execSql, d.rawSql.Args, d.err = d.genQuery(strings.TrimLeft(execSql, " \t\n\r"), args)
	if d.err != nil {
		return r, d.err
	}
	if len(execSql) <= 6 {
		return r, errors.New("SQL语句错误")
	}
	d.rawSql.Query.Reset()
	d.rawSql.Query.WriteString(execSql)
	r = &Result{
		Sql:  d.rawSql.Query,
		Args: d.rawSql.Args,
	}

	// 处理虚拟表
	return d.cteQueryBuild(r)
}

// CteQuery 虚拟表查询
// WITH
//
//	临时表1名称 [(列名1, 列名2, ...)] AS (SELECT/INSERT/UPDATE/DELETE 语句),
//
// cte 默认 WITH
func (d *Sql) CteQuery(query *CteQuery, cte ...string) (gs *Sql) {
	if query == nil {
		return d
	}
	if query.err != nil {
		d.err = query.err
		return d
	}
	var queryCet = "WITH"
	if len(cte) > 0 {
		queryCet = cte[0]
	}
	query.SetCet(queryCet)
	d.cteQuery = query
	return d
}

// cteQueryBuild 构建虚拟表查询
func (d *Sql) cteQueryBuild(r *Result) (*Result, error) {
	if d.err != nil {
		return r, d.err
	}
	if d.cteQuery != nil && len(d.cteQuery.VirtualTable) > 0 {
		var cteQuery = new(strings.Builder)
		var cteQueryArgs = make([]any, 0)
		cteQuery.WriteString(d.cteQuery.Cte)
		cteQuery.WriteString(`
`)
		var vqMaxKey = len(d.cteQuery.VirtualTable) - 1
		var vq *Result
		for i, cq := range d.cteQuery.VirtualTable {
			cteQuery.WriteString("  ")
			cteQuery.WriteString(cq.TableAlias)
			cteQuery.WriteString(" AS (")
			vq, d.err = cq.Sql.Select()
			if d.err != nil {
				return r, d.err
			}
			cteQuery.WriteString(vq.Sql.String())
			cteQuery.WriteString(")")
			if i < vqMaxKey {
				cteQuery.WriteString(",")
			}
			cteQuery.WriteString(`
`)
			cteQueryArgs = append(cteQueryArgs, vq.Args...)
		}
		cteQuery.WriteString(r.Sql.String())
		cteQueryArgs = append(cteQueryArgs, r.Args...)
		r.Sql = cteQuery
		r.Args = cteQueryArgs
	}
	return r, d.err
}

// GetWhereLen 获取where 条件数量
func (d *Sql) GetWhereLen() int {
	return len(d.whereCtrl)
}

// GetJoinLen 获取join 条件数量
func (d *Sql) GetJoinLen() int {
	return len(d.joinCtrl)
}

// GetGroupLen 获取分组数量
func (d *Sql) GetGroupLen() int {
	return d.group.Query.Len()
}

// getSearchWhere 构建适配前端的查询
func (d *Sql) getSearchWhere(fieldCmpOp string, fileVal any) (op string, whereVal []any, err error) {
	op = ""
	whereVal = []any{fileVal}
	var val []any
	var ok bool
	//生成查询
	switch strings.ToLower(fieldCmpOp) {
	case "=", "equal":
		op = "= ?"
	case "<", "lt":
		op = "< ?"
	case ">", "gt":
		op = "> ?"
	case ">=", "gte":
		op = ">= ?"
	case "<=", "lte":
		op = "<= ?"
	case "<>", "notequal", "!=":
		op = "!= ?"
	case "in", "not in":
		whereVal = make([]any, 0)
		if val, ok = fileVal.([]any); ok {
			whereVal = append(whereVal, val)
		} else {
			if strVal, ok := fileVal.(string); ok {
				strArr := strToArr(strVal, ",", false)
				whereVal = append(whereVal, strArr)
			}
		}
		if len(whereVal) > 0 {
			op = strings.ToUpper(fieldCmpOp) + " (?)"
		}
	case "like":
		op = "LIKE ?"
		whereVal = []any{fmt.Sprintf("%%%v%%", fileVal)}
	case "between":
		whereVal = make([]any, 0)
		if val, ok = fileVal.([]any); ok {
			if len(val) == 2 {
				whereVal = append(whereVal, val...)
			}
		} else {
			if strVal, ok := fileVal.(string); ok {
				btVal := strToArr(strVal, ",", false)
				if len(btVal) == 2 {
					whereVal = append(whereVal, btVal[0], btVal[1])
				}
			}
		}
		if len(whereVal) == 2 {
			op = "BETWEEN ? AND ?"
		}
	default:
		return op, whereVal, errors.New("不支持搜索条件：" + fieldCmpOp)
	}
	return op, whereVal, nil
}

// genWhere 构建查询条件
func (d *Sql) genWhere() (whereStr string, args []any) {
	whereStr = ""
	args = make([]any, 0)
	if len(d.whereCtrl) == 0 {
		return whereStr, args
	}
	for _, v := range d.whereCtrl {
		var ct string
		if v.Ct != "" {
			ct = " " + v.Ct + " "
		}
		whereStr += ct + v.Query.String()
		args = append(args, v.Args...)
	}
	return whereStr, args
}

// genJoin 构建连表条件
func (d *Sql) genJoin() (joinStr string, args []any) {
	joinStr = ""
	args = make([]any, 0)
	if len(d.joinCtrl) == 0 {
		return joinStr, args
	}
	for _, v := range d.joinCtrl {
		joinStr += " " + v.Query.String()
		if len(v.Args) > 0 {
			args = append(args, v.Args...)
		}
	}
	joinStr = strings.TrimLeft(joinStr, " ")
	return joinStr, args
}

// genPrePil 构建预编译sql where id = ?
func (d *Sql) genPrePil(query string, pos, len int) (q string, err error) {
	return genPrePil(query, pos, len)
}

// genQuery 处理数组 匹配 ? | IN (?), [][]int{1,2,3} => IN (?, ?, ?), []int{1,2,3}
func (d *Sql) genQuery(query string, args []any) (string, []any, error) {
	var newArgs = make([]any, 0)
	var err error
	var genArgs = make([]any, 0)
	var tmpArgs = make([]any, 0)
	var arrList [][]any
	var phIndex = &atomic.Int32{}
	phIndex.Add(1)
	var argsLen int
	if len(args) > 0 {
		// 处理子表
		for k, v := range args {
			if db, ok := v.(*Sql); ok {
				r, e := db.Select()
				if e != nil {
					return "", nil, e
				}
				query, d.err = replaceIndex(query, k+1, r.Sql.String())
				if d.err != nil {
					return "", nil, d.err
				}
				tmpArgs = append(tmpArgs, r.Args...)
			} else {
				tmpArgs = append(tmpArgs, v)
			}
		}
		// 处理数组
		for _, arg := range tmpArgs {
			switch arr := arg.(type) {
			case []int:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []int8:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []int16:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []int32:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []int64:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []uint:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []uint8:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []uint16:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []uint32:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []uint64:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []string:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []float32:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []float64:
				query, genArgs, err = genQueryList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []bool:
				var arrInt = make([]int, len(arr))
				for i, v := range arr {
					if v {
						arrInt[i] = 1
					} else {
						arrInt[i] = 0
					}
				}
				query, genArgs, err = genQueryList(query, arrInt, phIndex)
				newArgs = append(newArgs, genArgs...)
			case []any:
				if conv.IsArrayOrSlice(arr[0]) {
					arrList, err = arrToArrList(arr)
					if err != nil {
						return "", nil, err
					}
					query, genArgs, err = genQueryGroupAnyList(query, arrList, phIndex)
					newArgs = append(newArgs, genArgs...)
				} else {
					argsLen = len(arr)
					query, err = d.genPrePil(query, int(phIndex.Load()), argsLen)
					if err != nil {
						return "", nil, err
					}
					phIndex.Add(int32(argsLen))
					newArgs = append(newArgs, arr...)
				}
			case [][]int:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]int8:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]int16:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]int32:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]int64:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]uint:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]uint8:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]uint16:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]uint32:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]uint64:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]string:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]float32:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]float64:
				query, genArgs, err = genQueryGroupList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]bool:
				var arrInt = make([][]int, len(arr))
				for i, v := range arr {
					for _, vv := range v {
						if vv {
							arrInt[i] = append(arrInt[i], 1)
						} else {
							arrInt[i] = append(arrInt[i], 0)
						}
					}
				}
				query, genArgs, err = genQueryGroupList(query, arrInt, phIndex)
				newArgs = append(newArgs, genArgs...)
			case [][]any:
				query, genArgs, err = genQueryGroupAnyList(query, arr, phIndex)
				newArgs = append(newArgs, genArgs...)
			case RawArg:
				phIndex.Add(1)
				newArgs = append(newArgs, arr.val)
			default:
				// 处理数组和数组切片混合的场景
				var refVal = reflect.ValueOf(arg)
				if refVal.Kind() == reflect.Array || refVal.Kind() == reflect.Slice {
					var elemRefType = refVal.Type().Elem()
					if elemRefType.Kind() == reflect.Slice || elemRefType.Kind() == reflect.Array {
						// 二维数组
						if refVal.Len() <= 0 {
							return "", nil, errors.New("invalid array child length")
						}
						var sliceList = make([][]any, refVal.Len())
						var elemRefVal reflect.Value
						var elemRefLen = refVal.Index(0).Len()
						for i := 0; i < refVal.Len(); i++ {
							elemRefVal = refVal.Index(i)
							var sliceChild = make([]any, elemRefLen)
							for j := 0; j < elemRefLen; j++ {
								sliceChild[j] = elemRefVal.Index(j).Interface()
							}
							sliceList[i] = sliceChild
						}
						query, genArgs, err = genQueryGroupAnyList(query, sliceList, phIndex)
						newArgs = append(newArgs, genArgs...)
					} else {
						// 一维数组
						var slice = make([]any, refVal.Len())
						argsLen = refVal.Len()
						for i := 0; i < argsLen; i++ {
							slice[i] = refVal.Index(i).Interface()
						}
						query, err = d.genPrePil(query, int(phIndex.Load()), argsLen)
						if err != nil {
							return "", nil, err
						}
						phIndex.Add(int32(argsLen))
						newArgs = append(newArgs, slice...)
					}
				} else {
					phIndex.Add(1)
					newArgs = append(newArgs, arg)
				}
			}
		}
	}
	return query, newArgs, err
}

// ppNum 占位符出现的次数
func (d *Sql) ppNum(query string) int {
	return len(strings.Split(query, "?")) - 1
}
