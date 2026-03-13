package orm2

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"rest_demo/pkg/log"
	"rest_demo/pkg/utils/arrays"
	"rest_demo/pkg/utils/str"
	"strings"
)

const (
	TagFlag            = "orm"         // ORM Tag标记
	TagPrimaryKeyFlag  = "primary_key" // ORM 主键标记
	TableNameMethod    = "TableName"   // ORM 表名方法
	TableAliasMethod   = "TableAlias"  // ORM 表别名方法
	TagFlagCreateFalse = "c:false"     // 禁止插入
	TagFlagUpdateFalse = "u:false"     // 禁止修改
	TagColName         = "col_name"    // ORM 取db字段时用的名称，否则对字段名进行转换（注意，不取json tag）
)

// GetDB 获取DB对象
func GetDB() *sql.DB {
	return hhdb.GetDBOP().GetDBObj()
}

// Begin 开启事务
func Begin() (*sql.Tx, error) {
	begin, err := GetDB().Begin()
	if err != nil {
		log.Error.Println(err)
	}
	return begin, err
}

// Query 查询记录
func Query[T any](query string, args ...any) ([]*T, error) {
	rows, err := GetDB().Query(query, args...)
	if err != nil {
		log.Error.Print(err)
		return []*T{}, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	columns, err := rows.Columns()
	if err != nil {
		log.Error.Print(err)
		return []*T{}, err
	}
	var result []*T
	for rows.Next() {
		item := new(T)
		_, dest := GetFieldPointerFilter(item, columns...)
		err = rows.Scan(dest...)
		if err != nil {
			log.Error.Print(err)
			return []*T{}, err
		}
		result = append(result, item)
	}
	log.Debug.Printf("<==      Total: %d", len(result))
	return result, nil
}

// QueryOne 查询一条记录
func QueryOne[T any](query string, args ...any) (*T, error) {
	rows, err := GetDB().Query(query, args...)
	if err != nil {
		log.Error.Print(err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)
	columns, err := rows.Columns()
	if err != nil {
		log.Error.Print(err)
		return nil, err
	}
	if rows.Next() {
		item := new(T)
		_, dest := GetFieldPointerFilter(item, columns...)
		err = rows.Scan(dest...)
		if err != nil {
			log.Error.Print(err)
			return nil, err
		}
		log.Debug.Printf("<==      Total: %d", 1)
		return item, nil
	}
	log.Debug.Printf("<==      Total: %d", 0)
	return nil, nil
}

// QueryByWrapper 查询记录
func QueryByWrapper[T any](w *Wrapper) ([]*T, error) {
	return Query[T](w.GetSql(), w.GetArgs()...)
}

// Exec 执行SQL
func Exec(sql string, args ...any) (int64, int64, error) {
	return TExec(nil, sql, args...)
}

// TExec 执行SQL
func TExec(tx *sql.Tx, sqlString string, args ...any) (int64, int64, error) {
	var result sql.Result
	var err error
	if tx != nil {
		result, err = tx.Exec(sqlString, args...)
	} else {
		result, err = GetDB().Exec(sqlString, args...)
	}
	if err != nil {
		log.Error.Print(err)
		return 0, 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		log.Error.Print(err)
		return affected, 0, err
	}
	insertId, err := result.LastInsertId()
	if err != nil {
		log.Error.Print(err)
		return affected, insertId, err
	}
	log.Debug.Printf("<==   Affected: %v", affected)
	return affected, insertId, err
}

// FindByKey 通过主键获取记录
func FindByKey[T any](keys ...any) (*T, error) {
	model := new(T)
	f, _ := GetFieldPointer(model)
	k, _ := getPrimaryKey(model)
	if len(k) != len(keys) {
		return nil, fmt.Errorf("sql: expected %d destination keys, not %d", len(k), len(keys))
	}
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("SELECT ")
	sqlBuild.WriteString(strings.Join(f, ", "))
	sqlBuild.WriteString(" FROM ")
	sqlBuild.WriteString(GetTableNameAlias(model))
	sqlBuild.WriteString(" WHERE ")
	wrapper := QueryWrapper()
	for i, column := range k {
		wrapper.Eq(column, keys[i])
	}
	sqlBuild.WriteString(wrapper.GetWhereSql())
	return QueryOne[T](sqlBuild.String(), keys...)
}

// FindOne 查询一条记录
func FindOne[T any](wrapper *Wrapper) (*T, error) {
	if len(wrapper.selects) == 0 {
		f, _ := GetFieldPointer(new(T))
		wrapper.Selects(f...)
	}
	if len(wrapper.from) == 0 {
		wrapper.From(GetTableNameAlias(new(T)))
	}
	return QueryOne[T](wrapper.GetSql(), wrapper.args...)
}

// FindPage 分页查询
func FindPage[T any](page *Page[T], wrapper *Wrapper) (*Page[T], error) {
	if page.CurrPage < 1 {
		return page, fmt.Errorf("sql: CurrPage must be greater than 0, not %d", page.CurrPage)
	}
	if page.PageNums < 1 {
		return page, fmt.Errorf("sql: PageNums must be greater than 0, not %d", page.PageNums)
	}
	model := new(T)
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("SELECT ")
	if len(wrapper.GetSelectSql()) != 0 {
		sqlBuild.WriteString(wrapper.GetSelectSql())
	} else {
		var f []string
		if len(wrapper.filters) == 0 {
			f, _ = GetFieldPointer(model)
		} else {
			f, _ = GetFieldPointerFilter(model, wrapper.filters...)
		}
		sqlBuild.WriteString(strings.Join(f, ", "))
	}
	sqlBuild.WriteString(" FROM ")
	if len(wrapper.GetFromSql()) != 0 {
		sqlBuild.WriteString(wrapper.GetFromSql())
	} else {
		sqlBuild.WriteString(GetTableNameAlias(model))
	}
	if len(wrapper.GetWhereSql()) != 0 {
		sqlBuild.WriteString(" WHERE ")
		sqlBuild.WriteString(wrapper.GetWhereSql())
	}
	if len(wrapper.GetGroupSql()) != 0 {
		sqlBuild.WriteString(" GROUP BY ")
		sqlBuild.WriteString(wrapper.GetGroupSql())
	}

	// 统计分页总数量
	count := fmt.Sprintf("SELECT COUNT(1) AS total_nums FROM (%s) AS T1", sqlBuild.String())
	one, err := QueryOne[Page[T]](count, wrapper.args...)
	if err != nil {
		return page, err
	}

	page.TotalNums = one.TotalNums
	if one.TotalNums > 0 {
		orderSql := wrapper.GetOrderSql()
		if orderSql != "" {
			sqlBuild.WriteString(" ORDER BY ")
			sqlBuild.WriteString(wrapper.GetOrderSql())
		}
		sqlBuild.WriteString(" ")
		sqlBuild.WriteString(page.getLimitSql())
		query, err := Query[T](sqlBuild.String(), wrapper.args...)
		if err != nil {
			return page, err
		}
		page.PageData = query
	} else {
		page.PageData = make([]*T, 0)
	}
	return page, nil
}

// FindAll 分页查询
func FindAll[T any](wrapper *Wrapper) ([]*T, error) {
	model := new(T)
	f, _ := GetFieldPointer(new(T))
	if len(wrapper.selects) == 0 {
		wrapper.Selects(f...)
	}
	if len(wrapper.from) == 0 {
		wrapper.From(GetTableNameAlias(model))
	}
	return QueryByWrapper[T](wrapper)
}

// Create 创建一条记录
func Create[T any](model *T) (*T, error) {
	return TCreate(nil, model)
}

// TCreate 创建一条记录
func TCreate[T any](tx *sql.Tx, model *T) (*T, error) {
	filter, _ := getFieldByTag(model, TagFlagCreateFalse) // 过滤禁止插入的字段
	f, fv := GetFieldValue(model, false, filter...)
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("INSERT INTO")
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(GetTableName(model))
	sqlBuild.WriteString(" (")
	sqlBuild.WriteString(strings.Join(f, ", "))
	sqlBuild.WriteString(") VALUES (")
	for i := 0; i < len(fv); i++ {
		sqlBuild.WriteString("?")
		if i != len(fv)-1 {
			sqlBuild.WriteString(", ")
		}
	}
	sqlBuild.WriteString(")")
	affected, insertId, err := TExec(tx, sqlBuild.String(), fv...)
	if err != nil {
		return model, err
	}
	if affected != 1 {
		log.Error.Printf("sql: Affected %d", affected)
		return model, err
	}
	_, v := getFieldReflectByTag(model, TagPrimaryKeyFlag)
	if len(v) > 0 {
		v[0].SetInt(insertId)
	}
	return model, err
}

// Update 修改记录
func Update[T any](model *T, wrapper *Wrapper) (int64, error) {
	return TUpdate[T](nil, model, wrapper)
}

// TUpdate 修改记录
func TUpdate[T any](tx *sql.Tx, model *T, wrapper *Wrapper) (int64, error) {
	filter, _ := getFieldByTag(model, TagFlagUpdateFalse) // 过滤禁止修改的字段
	f, fv := GetFieldValue(model, true, filter...)
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("UPDATE ")
	sqlBuild.WriteString(GetTableNameAlias(model))
	sqlBuild.WriteString(" SET ")
	for i := 0; i < len(f); i++ {
		sqlBuild.WriteString(f[i])
		sqlBuild.WriteString("=?")
		if i != len(f)-1 {
			sqlBuild.WriteString(", ")
		}
	}
	sqlBuild.WriteString(" WHERE ")
	sqlBuild.WriteString(wrapper.GetWhereSql())
	affected, _, err := TExec(tx, sqlBuild.String(), append(fv, wrapper.args...)...)
	return affected, err
}

// UpdateByKey 通过主键修改一条记录
func UpdateByKey[T any](model *T) (*T, error) {
	return TUpdateByKey(nil, model)
}

// TUpdateByKey 通过主键修改一条记录
func TUpdateByKey[T any](tx *sql.Tx, model *T) (*T, error) {
	k, kv := getPrimaryKey(model)
	query := QueryWrapper()
	for i := 0; i < len(k); i++ {
		query.Eq(k[i], kv[i])
	}
	_, err := TUpdate(tx, model, query)
	if err != nil {
		return model, err
	}
	return model, nil
}

// BatchCreate 批量创建
func BatchCreate[T any](modelList []*T, size int) (int64, int64, error) {
	return TBatchCreate(nil, modelList, size)
}

// TBatchCreate 批量创建
func TBatchCreate[T any](tx *sql.Tx, modelList []*T, size int) (int64, int64, error) {
	if len(modelList) == 0 {
		return 0, 0, fmt.Errorf("sql: Must have an element %s", "")
	}
	var totalAffected, totalInsertId int64
	for _, itemList := range arrays.Split(modelList, size) {
		filter, _ := getFieldByTag(itemList[0], TagFlagCreateFalse) // 过滤禁止插入的字段
		f, _ := GetFieldValue(itemList[0], false, filter...)
		var fv []any
		sqlBuild := strings.Builder{}
		sqlBuild.WriteString("INSERT INTO ")
		sqlBuild.WriteString(GetTableName(itemList[0]))
		sqlBuild.WriteString(" (")
		sqlBuild.WriteString(strings.Join(f, ", "))
		sqlBuild.WriteString(") VALUES")
		for i, model := range itemList {
			sqlBuild.WriteString(" (")
			_, v := GetFieldValue(model, false, filter...)
			fv = append(fv, v...)
			for j := range f {
				sqlBuild.WriteString("?")
				if j != len(f)-1 {
					sqlBuild.WriteString(", ")
				}
			}
			sqlBuild.WriteString(")")
			if i != len(itemList)-1 {
				sqlBuild.WriteString(",")
			}
		}
		affected, insertId, err := TExec(tx, sqlBuild.String(), fv...)
		if err != nil {
			return totalAffected, totalInsertId, err
		}
		totalAffected += affected
		totalInsertId = insertId
	}
	return totalAffected, totalInsertId, nil
}

// ReplaceInto 替换记录
func ReplaceInto[T any](modelList []*T) (int64, int64, error) {
	return TReplaceInto(nil, modelList)
}

// BatchReplaceInto 替换记录
func BatchReplaceInto[T any](modelList []*T, size int) (int64, int64, error) {
	var totalAffected, totalInsertId int64
	for _, itemList := range arrays.Split(modelList, size) {
		affected, insertId, err := TReplaceInto(nil, itemList)
		if err != nil {
			return totalAffected, totalInsertId, err
		}
		totalAffected += affected
		totalInsertId = insertId
	}
	return totalAffected, totalInsertId, nil
}

// TReplaceInto 替换记录
func TReplaceInto[T any](tx *sql.Tx, modelList []*T) (int64, int64, error) {
	if len(modelList) == 0 {
		return 0, 0, fmt.Errorf("sql: Must have an element %s", "")
	}
	filter, _ := getFieldByTag(modelList[0], TagFlagCreateFalse) // 过滤禁止插入的字段
	f, _ := GetFieldValue(modelList[0], false, filter...)
	var fv []any
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("REPLACE INTO ")
	sqlBuild.WriteString(GetTableName(modelList[0]))
	sqlBuild.WriteString(" (")
	sqlBuild.WriteString(strings.Join(f, ", "))
	sqlBuild.WriteString(") VALUES")
	for i, model := range modelList {
		sqlBuild.WriteString(" (")
		_, v := GetFieldValue(model, false, filter...)
		fv = append(fv, v...)
		for j := range f {
			sqlBuild.WriteString("?")
			if j != len(f)-1 {
				sqlBuild.WriteString(", ")
			}
		}
		sqlBuild.WriteString(")")
		if i != len(modelList)-1 {
			sqlBuild.WriteString(",")
		}
	}
	return TExec(tx, sqlBuild.String(), fv...)
}

// Delete 删除记录
func Delete[T any](wrapper *Wrapper) (int64, error) {
	return TDelete[T](nil, wrapper)
}

// TDelete 删除记录
func TDelete[T any](tx *sql.Tx, wrapper *Wrapper) (int64, error) {
	model := new(T)
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("DELETE ")
	if len(GetTableAlias(model)) > 0 {
		sqlBuild.WriteString(GetTableAlias(model))
		sqlBuild.WriteString(" ")
	}
	sqlBuild.WriteString("FROM")
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(GetTableNameAlias(model))
	sqlBuild.WriteString(" WHERE ")
	sqlBuild.WriteString(wrapper.GetWhereSql())
	affected, _, err := TExec(tx, sqlBuild.String(), wrapper.args...)
	return affected, err
}

// DeleteByKey 通过主键删除记录
func DeleteByKey[T any](keys ...any) (int64, error) {
	return TDeleteByKey[T](nil, keys...)
}

// TDeleteByKey 通过主键删除记录
func TDeleteByKey[T any](tx *sql.Tx, keys ...any) (int64, error) {
	model := new(T)
	k, _ := getPrimaryKey(model)
	if len(k) != len(keys) {
		return 0, fmt.Errorf("sql: expected %d destination keys, not %d", len(k), len(keys))
	}
	query := QueryWrapper()
	for i := 0; i < len(k); i++ {
		query.Eq(k[i], keys[i])
	}
	return TDelete[T](tx, query)
}

// GetTableName 通过模型指针获取表名
func GetTableName(model any) string {
	var name string
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Pointer {
		log.Error.Printf("orm: check type error not Pointer")
		return name
	}
	nameMethod := v.MethodByName(TableNameMethod)
	if nameMethod.Kind() == 0 {
		log.Error.Printf("orm: " + v.String() + " not found " + TableNameMethod)
		return name
	}
	return nameMethod.Call(nil)[0].String()
}

// GetTableAlias 通过模型指针获取表别名
func GetTableAlias(model any) string {
	v := reflect.ValueOf(model)
	var alias string
	aliasMethod := v.MethodByName(TableAliasMethod)
	if aliasMethod.Kind() != 0 {
		alias = aliasMethod.Call(nil)[0].String()
	}
	return alias
}

// GetTableNameAlias 通过模型指针获取表名加别名
func GetTableNameAlias(model any) string {
	name := GetTableName(model)
	alias := GetTableAlias(model)
	if len(alias) == 0 {
		return fmt.Sprintf("%s", name)
	}
	return fmt.Sprintf("%s %s", name, alias)
}

// GetFieldValue 通过模型指针获取字段及字段值
func GetFieldValue(model any, hasAlias bool, filter ...string) ([]string, []any) {
	return getChildFieldValue(model, model, hasAlias, filter...)
}

// getChildFieldValue 通过模型指针获取字段及字段值
func getChildFieldValue(model any, parent any, hasAlias bool, filter ...string) ([]string, []any) {
	t, err := getReflectType(model)
	if err != nil {
		return []string{}, []any{}
	}
	v, err := getReflectValue(model)
	if err != nil {
		return []string{}, []any{}
	}
	var alias string
	if hasAlias {
		alias = GetTableAlias(parent)
	}
	var f []string
	var p []any
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			childF, childP := getChildFieldValue(v.Field(i).Addr().Interface(), parent, hasAlias, filter...)
			f = append(f, childF...)
			p = append(p, childP...)
		} else {
			fieldName := getColumnName(alias, t.Field(i))
			if !containsColumn(filter, fieldName) {
				f = append(f, fieldName)
				p = append(p, v.Field(i).Interface())
			}
		}
	}
	return f, p
}

// GetFieldPointer 通过模型指针获取字段及指针
func GetFieldPointer(model any) ([]string, []any) {
	return GetChildFieldPointer(model, model)
}

// GetChildFieldPointer 通过模型指针获取字段及指针
func GetChildFieldPointer(model any, parent any) ([]string, []any) {
	t, err := getReflectType(model)
	if err != nil {
		return []string{}, []any{}
	}
	v, err := getReflectValue(model)
	if err != nil {
		return []string{}, []any{}
	}
	alias := GetTableAlias(parent)
	var f []string
	var p []any
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			childF, childP := GetChildFieldPointer(v.Field(i).Addr().Interface(), parent)
			f = append(f, childF...)
			p = append(p, childP...)
		} else {
			f = append(f, getColumnName(alias, t.Field(i)))
			p = append(p, v.Field(i).Addr().Interface())
		}
	}
	return f, p
}

// GetFieldPointerFilter 通过模型指针获取字段及指针
func GetFieldPointerFilter(model any, filters ...string) ([]string, []any) {
	var filteredF []string
	var filteredP []any
	f, p := GetFieldPointer(model)
	for _, filter := range filters {
		for i := 0; i < len(f); i++ {
			if equalColumn(f[i], filter) {
				filteredF = append(filteredF, f[i])
				filteredP = append(filteredP, p[i])
			}
		}
	}
	return filteredF, filteredP
}

// getPrimaryKey 通过模型指针获取主键
func getPrimaryKey(model any) ([]string, []any) {
	return getFieldByTag(model, TagPrimaryKeyFlag)
}

// getFieldByTag 根据模型获取拥有指定Tag的字段
func getFieldByTag(model any, tags ...string) ([]string, []any) {
	return getChildFieldByTag(model, model, tags...)
}

// getChildFieldByTag 根据模型获取拥有指定Tag的字段
func getChildFieldByTag(model any, parent any, tags ...string) ([]string, []any) {
	t, err := getReflectType(model)
	if err != nil {
		return []string{}, []any{}
	}
	v, err := getReflectValue(model)
	if err != nil {
		return []string{}, []any{}
	}
	alias := GetTableAlias(parent)
	var f []string
	var p []any
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			childF, childP := getChildFieldByTag(v.Field(i).Addr().Interface(), parent, tags...)
			f = append(f, childF...)
			p = append(p, childP...)
		} else {
			for _, tag := range tags {
				if strings.Contains(t.Field(i).Tag.Get(TagFlag), tag) {
					f = append(f, getColumnName(alias, t.Field(i)))
					p = append(p, v.Field(i).Interface())
				}
			}
		}
	}
	return f, p
}

// getFieldReflectByTag 根据模型获取拥有指定Tag的字段反射
func getFieldReflectByTag(model any, tags ...string) ([]reflect.StructField, []reflect.Value) {
	t, err := getReflectType(model)
	if err != nil {
		return []reflect.StructField{}, []reflect.Value{}
	}
	v, err := getReflectValue(model)
	if err != nil {
		return []reflect.StructField{}, []reflect.Value{}
	}
	var rf []reflect.StructField
	var rv []reflect.Value
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.Kind() == reflect.Struct {
			childF, childP := getFieldReflectByTag(v.Field(i).Addr().Interface(), tags...)
			rf = append(rf, childF...)
			rv = append(rv, childP...)
		} else {
			for _, tag := range tags {
				if strings.Contains(t.Field(i).Tag.Get(TagFlag), tag) {
					rf = append(rf, t.Field(i))
					rv = append(rv, v.Field(i))
				}
			}
		}
	}
	return rf, rv
}

// getColumnName 获取字段对应的数据库列名
func getColumnName(alias string, f reflect.StructField) string {
	colName, ok := f.Tag.Lookup(TagColName)
	if !ok {
		colName = str.ToSnakeLower(f.Name)
	}
	if len(alias) == 0 {
		return fmt.Sprintf("`%s`", colName)
	}
	return fmt.Sprintf("`%s`.`%s`", alias, colName)
}

// equalColumn 判断两列名是否相等
func equalColumn(column1 string, column2 string) bool {
	split1 := str.Split(column1, ".")
	split2 := str.Split(column2, ".")
	final1 := split1[len(split1)-1]
	final2 := split2[len(split2)-1]
	column1 = strings.ToLower(strings.Trim(final1, "`"))
	column2 = strings.ToLower(strings.Trim(final2, "`"))
	return column1 == column2
}

// containsColumn 判断数组中是否存在某字段
func containsColumn(columnList []string, column string) bool {
	for _, item := range columnList {
		if equalColumn(item, column) {
			return true
		}
	}
	return false
}

// getReflectType 获取模型反射类型
func getReflectType(model any) (reflect.Type, error) {
	p := reflect.TypeOf(model)
	if p.Kind() != reflect.Pointer {
		log.Error.Printf("orm: check type error not Pointer")
		return nil, errors.New("orm: check type error not Pointer")
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		log.Error.Printf("orm: check type error not Struct")
		return nil, errors.New("orm: check type error not Pointer")
	}
	return s, nil
}

// getReflectValue 获取模型反射值
func getReflectValue(model any) (reflect.Value, error) {
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Pointer {
		log.Error.Printf("orm: check type error not Pointer")
		return v, errors.New("orm: check type error not Pointer")
	}
	e := v.Elem()
	if e.Kind() != reflect.Struct {
		log.Error.Printf("orm: check type error not Struct")
		return e, errors.New("orm: check type error not Pointer")
	}
	return e, nil
}
