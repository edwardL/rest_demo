package gdbtmp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// mysqlKeywords sql查询关键词会使用 `包裹`
	mysqlKeywords = map[string]struct{}{
		"add":    struct{}{},
		"alter":  struct{}{},
		"case":   struct{}{},
		"create": struct{}{},
		"delete": struct{}{},
		"drop":   struct{}{},
		"from":   struct{}{},
		"group":  struct{}{},
		"insert": struct{}{},
		"select": struct{}{},
		"update": struct{}{},
		"where":  struct{}{},
		"status": struct{}{},
		"name":   struct{}{},
		"port":   struct{}{},
		"desc":   struct{}{},
		"key":    struct{}{},
		"index":  struct{}{},
		"skip":   struct{}{},
		"order":  struct{}{},
		"tenant": struct{}{},
		"user":   struct{}{},
		"count":  struct{}{},
	}
	// zeroDelField // 0值忽略的字段
	zeroValIgnoreField = []string{"id", "ts", "create_time", "update_time"}
)

// Sql 构建sql操作对象
type Sql struct {
	err           error             // 错误信息
	table         string            // 表名
	args          []any             // 参数
	fields        GenCtrl           // 字段参数
	rawSql        GenCtrl           // 原生sql参数
	union         []GenCtrl         // union参数
	whereCtrl     []GenCtrl         // where参数
	joinCtrl      []GenCtrl         // 连表参数
	limitCtrl     GenCtrl           // 分页参数
	offsetCtrl    GenCtrl           // 分页参数
	group         GenCtrl           // 分组参数
	order         GenCtrl           // 排序参数
	fieldReplace  map[string]string // 更新和新增字段名替换
	fieldSelect   map[string]string // 更新和新增指定的字段
	fieldOmit     map[string]string // 不更新和新增指定字段
	saveDuplicate map[string]bool   // 更新时，如果字段值重复，则更新
}

// NewSql 创建sql对象
// 不支持变量替换 只有第一个参数生效
func NewSql(table ...any) *Sql {
	gs := &Sql{}
	if len(table) > 0 {
		gs.Table(table[0])
	}
	return gs
}

const (
	// FieldRemove val 填这个会忽略这个字段
	FieldRemove = "_field_remove_"
)

// RawBody 原样拼接
type RawBody string

// Raw 保持原样拼接 适配场景 age = age+1
func Raw(q string) RawBody {
	return RawBody(q)
}

// GenCtrl 构造条件
type GenCtrl struct {
	Query string
	Ct    string // and or
	Args  []any
}

// Result 返回结果
type Result struct {
	Sql  string
	Args []any
}

// CompSql 获取完整SQL
func (r Result) CompSql() string {
	if r.Args == nil {
		return r.Sql
	}
	var err error
	compSql := r.Sql
	var argsStr string
	for _, v := range r.Args {
		argsStr = ""
		if v == nil {
			argsStr = "NULL"
		} else {
			switch vt := v.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
				argsStr = ToString(vt)
			case bool:
				if vt {
					argsStr = "1"
				} else {
					argsStr = "0"
				}
			case []byte:
				argsStr = "'" + string(vt) + "'"
			default:
				argsStr = "'" + ToString(vt) + "'"
			}
		}
		compSql, err = replaceIndex(compSql, 1, argsStr)
	}
	if err != nil {
		defLog.CtxError(context.Background(), "创建完整sql错误：", err)
	}
	return compSql
}

// WhereFunc 闭包where 给where 加 ()
type WhereFunc func(gs *Sql) (err error)

// Select 获取查询语句
func (d *Sql) Select() (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "",
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理字段
	if d.fields.Query == "" {
		r.Sql = "SELECT * FROM"
	} else {
		r.Sql = "SELECT " + d.fields.Query + " FROM"
		r.Args = append(r.Args, d.fields.Args...)
	}

	// 处理表
	r.Sql += " " + d.table
	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	}

	// 处理 GROUP BY
	if d.group.Query != "" {
		r.Sql += " " + d.group.Query
		if d.group.Args != nil {
			r.Args = append(r.Args, d.group.Args...)
		}
	}

	// 处理 ORDER BY
	if d.order.Query != "" {
		r.Sql += " " + d.order.Query
		if d.order.Args != nil {
			r.Args = append(r.Args, d.order.Args...)
		}
	}

	// 处理 LIMIT OFFSET
	if d.limitCtrl.Query != "" {
		r.Sql += " " + d.limitCtrl.Query
		if d.limitCtrl.Args != nil {
			r.Args = append(r.Args, d.limitCtrl.Args...)
		}
	}
	if d.offsetCtrl.Query != "" {
		r.Sql += " " + d.offsetCtrl.Query
		if d.offsetCtrl.Args != nil {
			r.Args = append(r.Args, d.offsetCtrl.Args...)
		}
	}

	// 处理 union
	if len(d.union) > 0 {
		for _, v := range d.union {
			r.Sql += " " + v.Ct + " " + v.Query
			r.Args = append(r.Args, v.Args...)
		}
	}
	return r, d.err
}

// Count 获取统计语句 f 为COUNT(f[0])
func (d *Sql) Count(f ...string) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	cf := "*"
	if len(f) > 0 {
		cf = f[0]
	}
	r = &Result{
		Sql:  "SELECT COUNT(" + cf + ") FROM",
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理表
	r.Sql += " " + d.table
	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理连表
	query, args := d.genJoin()
	if query != "" {
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	}

	// 处理 GROUP BY
	if d.group.Query != "" {
		r.Sql += " " + d.group.Query
		if d.group.Args != nil {
			r.Args = append(r.Args, d.group.Args...)
		}
	}

	return r, d.err
}

// Update 更新某个字段
func (d *Sql) Update(column string, val any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "UPDATE " + d.table,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
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
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	// 处理更新字段
	if vr, ok := val.(RawBody); ok {
		r.Sql += " SET " + column + " = " + string(vr)
	} else {
		r.Sql += " SET " + column + " = ?"
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
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql = ""
		r.Args = make([]any, 0)
		d.err = errors.New("缺少更新条件")
	}

	return r, d.err
}

// Updates 更新多个字段 data = map[string]any 或者 Struct
func (d *Sql) Updates(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "UPDATE " + d.table,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理更新字段
	var upField = ""
	var upFieldMap = map[string]string{}
	var m map[string]any
	var igFn string
	m, d.err = toMap(data)
	for _, igFn = range zeroValIgnoreField {
		if _, ok := m[igFn]; ok && IsEmpty(m[igFn]) {
			delete(m, igFn)
		}
	}
	if m != nil {
		_, _, upFieldMap = d.getInsField([]map[string]any{m})
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
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	if upField != "" {
		r.Sql += " SET " + upField
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql = ""
		r.Args = make([]any, 0)
		d.err = errors.New("缺少更新条件")
	}
	return r, d.err
}

// UpdateIgnore 更新忽略 data = map[string]any 或者 Struct
func (d *Sql) UpdateIgnore(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "UPDATE IGNORE " + d.table,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理更新字段
	var upField = ""
	var upFieldMap = map[string]string{}
	var m map[string]any
	m, d.err = toMap(data)
	if m != nil {
		_, _, upFieldMap = d.getInsField([]map[string]any{m})
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
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	if upField != "" {
		r.Sql += " SET " + upField
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	}
	return r, d.err
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
	if d.saveDuplicate == nil {
		d.saveDuplicate = map[string]bool{}
	}
	r = &Result{
		Sql:  "",
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	var insFileMap = map[string]string{}
	var mArr []map[string]any
	mArr, d.err = toMaps(data)
	if len(mArr) > 0 {
		_, _, insFileMap = d.getInsField(mArr)
		//字段名
		var oldFieldName []string
		var fieldName []string
		fieldIgnoreSql := " ON DUPLICATE KEY UPDATE "
		for k, v := range mArr {
			placeholder := make([]string, 0)
			if k == 0 {
				for k2, _ := range v {
					if fn, ok := insFileMap[k2]; ok {
						oldFieldName = append(oldFieldName, k2)
						//	处理主键
						if !d.saveDuplicate[k2] {
							fieldIgnoreSql += fmt.Sprintf(" %s = VALUES( %s ),", fn, fn)
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
			r.Sql += "(" + strings.Join(placeholder, ",") + "),"
		}
		//处理尾部逗号
		fieldIgnoreSql = fieldIgnoreSql[:len(fieldIgnoreSql)-1]
		r.Sql = r.Sql[:len(r.Sql)-1]
		r.Sql = "INSERT INTO " + d.table + " (" + strings.Join(fieldName, ",") + ") VALUES" + r.Sql + fieldIgnoreSql + ";"
	}
	return r, d.err
}

// Delete 删除 DELETE {t} From table
func (d *Sql) Delete(t ...string) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	var genSql = "DELETE FROM " + d.table
	if len(t) > 0 {
		genSql = "DELETE " + strings.Join(t, ",") + " FROM " + d.table
	}
	r = &Result{
		Sql:  genSql,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	// 处理参数
	if d.args != nil {
		r.Args = append(r.Args, d.args...)
	}

	var query string
	var args []any

	// 处理连表
	query, args = d.genJoin()
	if query != "" {
		r.Sql += " " + query
		r.Args = append(r.Args, args...)
	}

	// 处理where
	if len(d.whereCtrl) > 0 {
		query, args = d.genWhere()
		if query != "" {
			r.Sql += " WHERE " + query
			r.Args = append(r.Args, args...)
		}
	} else {
		r.Sql = ""
		r.Args = make([]any, 0)
		d.err = errors.New("缺少删除条件")
	}

	// 处理limit
	// 处理 LIMIT OFFSET
	if d.limitCtrl.Query != "" {
		r.Sql += " " + d.limitCtrl.Query
		if d.limitCtrl.Args != nil {
			r.Args = append(r.Args, d.limitCtrl.Args...)
		}
	}
	if d.offsetCtrl.Query != "" {
		r.Sql += " " + d.offsetCtrl.Query
		if d.offsetCtrl.Args != nil {
			r.Args = append(r.Args, d.offsetCtrl.Args...)
		}
	}

	return r, d.err
}

// Create 创建 data = map[string]any 或者 Struct
func (d *Sql) Create(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "INSERT INTO " + d.table,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
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
		if d.fields.Query != "" {
			r.Sql += " (" + d.fields.Query + ")"
			r.Args = append(r.Args, d.fields.Args...)
		}
		r.Sql += " " + childRes.Sql
		r.Args = append(r.Args, childRes.Args...)
		return r, d.err
	}

	var insField = ""
	var insFieldMap = map[string]string{}
	var insArgs = make([]any, 0)
	var m map[string]any
	m, d.err = toMap(data)

	if m != nil {
		_, _, insFieldMap = d.getInsField([]map[string]any{m})
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
		r.Sql += " ( " + insField + " )"
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
		r.Sql += " VALUES ( " + val + " )"
	}

	return r, d.err
}

// CreateInBatches 批量创建
// data = []map[string]any 或者 []Struct
func (d *Sql) CreateInBatches(data any) (r *Result, err error) {
	if d.err != nil {
		return nil, d.err
	}
	r = &Result{
		Sql:  "INSERT INTO " + d.table,
		Args: make([]any, 0),
	}

	if d.rawSql.Query != "" {
		r.Sql = d.rawSql.Query
		r.Args = d.rawSql.Args
		return r, d.err
	}

	var insField = ""
	var insFileArr = make([]string, 0)
	var insArgs = make([][]any, 0)
	var mArr []map[string]any
	mArr, d.err = toMaps(data)
	if len(mArr) > 0 {
		insField, insFileArr, _ = d.getInsField(mArr)
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
		r.Sql += " ( " + insField + " )"
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
		r.Sql += " VALUES " + valList
	}
	return r, d.err
}

// SaveDuplicate 更新条件
func (d *Sql) SaveDuplicate(fieldIgnore map[string]bool) *Sql {
	d.saveDuplicate = fieldIgnore
	return d
}

// getInsField 获取允许新增的字段
func (d *Sql) getInsField(mArr []map[string]any) (string, []string, map[string]string) {
	var insField = ""
	var fsLen = len(d.fieldSelect)
	var ofLen = len(d.fieldOmit)
	var rfLen = len(d.fieldReplace)
	var insFileArr = make([]string, 0)
	var insFileMap = map[string]string{}
	var igFn string
	for f, v := range mArr[0] {
		for _, igFn = range zeroValIgnoreField {
			if f == igFn && IsEmpty(v) {
				continue
			}
		}
		mapKey := f
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
				if kfOk {
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
	d.joinCtrl = append(d.joinCtrl, GenCtrl{
		Query: query,
		Args:  args,
	})
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
	return d.appendWhere(query, "AND", args...)
}

// WhereOr 或查询条件
// db.WhereOr("name = ?", "xxx")
// db.WhereOr("name = ? AND id IN (?)", "xxx", []int{1,2,3}).Where("age <> ?", "20")
func (d *Sql) WhereOr(query string, args ...any) (gs *Sql) {
	return d.appendWhere(query, "OR", args...)
}

// WhereReset 重置查询条件
func (d *Sql) WhereReset() (gs *Sql) {
	d.whereCtrl = make([]GenCtrl, 0)
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
		d.whereCtrl = make([]GenCtrl, 0)
	}

	query, args, d.err = d.genQuery(query, args)
	if d.err != nil {
		return d
	}
	d.whereCtrl = append(d.whereCtrl, GenCtrl{
		Query: query,
		Ct:    ct,
		Args:  args,
	})
	return d
}

// WhereGroup 查询条件组 AND (id = ? AND name = ?)
func (d *Sql) WhereGroup(wf WhereFunc) (gs *Sql) {
	return d.appendWhereGroup(wf, "AND")
}

// WhereGroupOr 查询条件组会加括号 OR (id = ? AND name = ?)
func (d *Sql) WhereGroupOr(wf WhereFunc) (gs *Sql) {
	return d.appendWhereGroup(wf, "OR")
}

// appendWhereGroup 构建查询条件组
func (d *Sql) appendWhereGroup(wf WhereFunc, ct string) (gs *Sql) {
	if d.err != nil {
		return d
	}
	if len(d.whereCtrl) <= 0 {
		ct = ""
		d.whereCtrl = make([]GenCtrl, 0)
	}
	db := NewSql()
	_ = wf(db)
	whereStr, args := db.genWhere()
	if whereStr == "" {
		return d
	}
	d.err = db.err
	d.whereCtrl = append(d.whereCtrl, GenCtrl{
		Query: "( " + whereStr + " )",
		Ct:    ct,
		Args:  args,
	})
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
	var table, err = d.getTableName(tb)
	if err != nil || table == "" {
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
				d.table, d.err = replaceIndex(d.table, k+1, r.Sql)
				d.args = append(d.args, r.Args...)
			} else {
				d.args = append(d.args, v)
			}
		}
	}
	return d
}

func (d *Sql) getTableName(tb any) (string, error) {
	if d.err != nil {
		return "", d.err
	}
	var err error
	var table string
	switch tbt := tb.(type) {
	case string:
		return tbt, nil
	case ModelFace:
		var val = reflect.ValueOf(tbt)
		if !isNilValue(val) {
			return tbt.TableName(), nil
		}
	}
	table, err = getTableNameRecursive(reflect.ValueOf(tb))
	if err != nil || table == "" {
		d.err = errors.New("未解析到表名")
		return "", d.err
	}
	return table, nil
}

// Clone 复制一个实例
func (d *Sql) Clone() (gs *Sql) {
	return &Sql{
		err:           d.err,
		table:         d.table,
		args:          d.args,
		fields:        d.fields,
		rawSql:        d.rawSql,
		union:         d.union,
		whereCtrl:     d.whereCtrl,
		joinCtrl:      d.joinCtrl,
		limitCtrl:     d.limitCtrl,
		offsetCtrl:    d.offsetCtrl,
		group:         d.group,
		order:         d.order,
		fieldSelect:   d.fieldSelect,
		fieldOmit:     d.fieldOmit,
		fieldReplace:  d.fieldReplace,
		saveDuplicate: d.saveDuplicate,
	}
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
		d.fields.Query = field
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
				d.fields.Query = strings.TrimLeft(newF, ", ")
			}
		}
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(d.fields.Query) != len(args) {
		d.err = errors.New("[ " + d.fields.Query + " ] 参数和占位符不匹配")
		return d
	}
	d.fields.Query, d.fields.Args, d.err = d.genQuery(d.fields.Query, args)
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
	d.limitCtrl = GenCtrl{
		Query: "LIMIT " + limit,
		Ct:    "",
	}
	d.limitCtrl.Query, d.limitCtrl.Args, d.err = d.genQuery(d.limitCtrl.Query, args)
	return d
}

// Offset .
func (d *Sql) Offset(offset int) (gs *Sql) {
	if d.err != nil {
		return d
	}
	d.offsetCtrl = GenCtrl{
		Query: "OFFSET ?",
		Ct:    "",
		Args:  []any{offset},
	}
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
	d.limitCtrl = GenCtrl{}
	d.offsetCtrl = GenCtrl{}
	return d
}

// Group .
// f[string | []string]
func (d *Sql) Group(f any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	groupField := ""
	switch field := f.(type) {
	case string:
		groupField = field
	case []string:
		if field != nil {
			newF := ""
			for _, v := range field {
				newF += ", " + v
			}
			if newF != "" {
				groupField = strings.TrimLeft(newF, ", ")
			}
		}
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(groupField) != len(args) {
		d.err = errors.New("[ " + groupField + " ] 参数和占位符不匹配")
		return d
	}
	d.group = GenCtrl{
		Query: "GROUP BY " + groupField,
		Ct:    "",
		Args:  make([]any, 0),
	}
	d.group.Query, d.group.Args, d.err = d.genQuery(d.group.Query, args)
	return d
}

// Order .
// f[string | []string]
func (d *Sql) Order(f any, args ...any) (gs *Sql) {
	if d.err != nil {
		return d
	}
	groupField := ""
	switch field := f.(type) {
	case string:
		groupField = field
	case []string:
		if field != nil {
			newF := ""
			for _, v := range field {
				newF += ", " + v
			}
			if newF != "" {
				groupField = strings.TrimLeft(newF, ", ")
			}
		}
	default:
		d.err = errors.New("字段只能是 string 或 []string")
	}
	if d.ppNum(groupField) != len(args) {
		d.err = errors.New("[ " + groupField + " ] 参数和占位符不匹配")
		return d
	}
	d.order = GenCtrl{
		Query: "ORDER BY " + groupField,
		Ct:    "",
		Args:  make([]any, 0),
	}
	d.order.Query, d.order.Args, d.err = d.genQuery(d.order.Query, args)
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
	orderArr := strings.Split(orderStr, ",")
	orderByArr := make([]string, 0)
	for _, v := range orderArr {
		oArr := strToArr(v, " ", true)
		if len(oArr) == 2 {
			var f = oArr[0]
			var orderDep = strings.ToUpper(oArr[1])
			if orderDep == "DESC" || orderDep == "ASC" {
				if filter != nil {
					if newF, ok := filter[f]; ok {
						if newF != "" {
							f = newF
						}
						// 开始排序
						orderByArr = append(orderByArr, f+" "+orderDep)
					} else {
						if !igErr {
							d.err = errors.New("字段：" + f + " 不能用于排序")
							return d
						}
					}
				}
			} else {
				if !igErr {
					d.err = errors.New("排序规则必须是：asc | desc")
					return d
				}
			}
		} else {
			if !igErr {
				d.err = errors.New("排序规则必须是：asc | desc")
				return d
			}
		}
	}
	d.Order(orderByArr)
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
	if len(txList) > 0 {
		for _, q := range txList {
			res, d.err = q.Select()
			if d.err != nil {
				return d
			}
			var unionInfo = GenCtrl{
				Query: "",
				Ct:    unionType,
				Args:  nil,
			}
			unionInfo.Query, unionInfo.Args, d.err = d.genQuery(res.Sql, res.Args)
			if d.err != nil {
				return d
			}
			d.union = append(d.union, unionInfo)
		}
	}
	return d
}

// Raw 原生查询
func (d *Sql) Raw(query string, args ...any) (gs *Sql) {
	d.rawSql.Query, d.rawSql.Args, d.err = d.genQuery(query, args)
	if d.err != nil {
		return d
	}
	return d
}

// GetWhereLen 获取where 条件数量
func (d *Sql) GetWhereLen() int {
	return len(d.whereCtrl)
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
		whereStr += ct + v.Query
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
		joinStr += " " + v.Query
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
	var tmpArgs = make([]any, 0)
	var arrList [][]any
	var phIndex = 1
	var argsLen int
	if len(args) > 0 {
		// 处理子表
		for k, v := range args {
			if db, ok := v.(*Sql); ok {
				r, e := db.Select()
				if e != nil {
					return "", nil, e
				}
				query, d.err = replaceIndex(query, k+1, r.Sql)
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
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []int8:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []int16:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []int32:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []int64:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []uint:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []uint8:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []uint16:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []uint32:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []uint64:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []string:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []float32:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []float64:
				query, err = genQueryList(query, arr, &phIndex, &newArgs)
			case []bool:
				var arrInt = make([]int, len(arr))
				for i, v := range arr {
					if v {
						arrInt[i] = 1
					} else {
						arrInt[i] = 0
					}
				}
				query, err = genQueryList(query, arrInt, &phIndex, &newArgs)
			case []any:
				if IsArrayOrSlice(arr[0]) {
					arrList, err = arrToArrList(arr)
					if err != nil {
						return "", nil, err
					}
					query, err = genQueryGroupAnyList(query, arrList, &phIndex, &newArgs)
				} else {
					argsLen = len(arr)
					query, err = d.genPrePil(query, phIndex, argsLen)
					if err != nil {
						return "", nil, err
					}
					phIndex += argsLen
					newArgs = append(newArgs, arr...)
				}
			case [][]int:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]int8:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]int16:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]int32:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]int64:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]uint:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]uint8:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]uint16:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]uint32:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]uint64:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]string:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]float32:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
			case [][]float64:
				query, err = genQueryGroupList(query, arr, &phIndex, &newArgs)
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
				query, err = genQueryGroupList(query, arrInt, &phIndex, &newArgs)
			case [][]any:
				query, err = genQueryGroupAnyList(query, arr, &phIndex, &newArgs)
			default:
				phIndex++
				newArgs = append(newArgs, arg)
			}
		}
	}
	return query, newArgs, err
}

// ppNum 占位符出现的次数
func (d *Sql) ppNum(query string) int {
	return len(strings.Split(query, "?")) - 1
}
