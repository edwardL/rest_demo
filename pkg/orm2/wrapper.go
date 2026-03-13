package orm2

import (
	"reflect"
	"strings"
)

const (
	And       = "AND"
	Eq        = "="
	NotEq     = "!="
	Gt        = ">"
	Ge        = ">="
	Lt        = "<"
	Le        = "<="
	LIKE      = "LIKE"
	In        = "IN"
	Between   = "BETWEEN"
	IsNotNull = "IS NOT NULL"
)

// Wrapper 查询构建
type Wrapper struct {
	selects       []string // SQL查询
	filters       []string // SQL查询过滤
	from          string   // SQL查询来源
	wheres        []string // Sql条件
	orders        []string // Sql排序
	groups        []string // Sql分组
	defaultOrders []string // Sql默认排序
	last          string   // Sql最后语句
	args          []any    // Sql参数值
}

// QueryWrapper 简单查询
func QueryWrapper() *Wrapper {
	return &Wrapper{}
}

// Selects 添加查询
func (wrapper *Wrapper) Selects(selects ...string) *Wrapper {
	wrapper.selects = append(wrapper.selects, selects...)
	return wrapper
}

// Filters 添加查询过滤
func (wrapper *Wrapper) Filters(filters ...string) *Wrapper {
	wrapper.filters = append(wrapper.filters, filters...)
	return wrapper
}

// From 添加查询来源
func (wrapper *Wrapper) From(from string) *Wrapper {
	wrapper.from = from
	return wrapper
}

// Group 添加分组
func (wrapper *Wrapper) Group(groups ...string) *Wrapper {
	wrapper.groups = append(wrapper.groups, groups...)
	return wrapper
}

// Order 添加排序
func (wrapper *Wrapper) Order(orders ...string) *Wrapper {
	wrapper.orders = append(wrapper.orders, orders...)
	return wrapper
}

// DefaultOrder 添加默认排序
func (wrapper *Wrapper) DefaultOrder(defaultOrders ...string) *Wrapper {
	wrapper.defaultOrders = append(wrapper.defaultOrders, defaultOrders...)
	return wrapper
}

// Last 在最后添加语句
func (wrapper *Wrapper) Last(last string) *Wrapper {
	if len(last) != 0 {
		wrapper.last = last
	}
	return wrapper
}

// And 添加一个条件
func (wrapper *Wrapper) And(where string, value ...any) *Wrapper {
	if len(strings.TrimSpace(where)) == 0 {
		return wrapper
	}
	wrapper.wheres = append(wrapper.wheres, where)
	wrapper.args = append(wrapper.args, value...)
	return wrapper
}

// Eq Equal
func (wrapper *Wrapper) Eq(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Eq)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// NotEq Not Equal
func (wrapper *Wrapper) NotEq(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(NotEq)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// Like LIKE "%%"
func (wrapper *Wrapper) Like(column string, value string) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(LIKE)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, "%"+strings.ReplaceAll(strings.ReplaceAll(value, "_", "\\_"), "%", "\\%")+"%")
	return wrapper
}

// In IN()
func (wrapper *Wrapper) In(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(In)
	sqlBuild.WriteString(" (")
	var p []string
	valueOf := reflect.ValueOf(value)
	values := valueOf.Slice(0, valueOf.Len())
	for i := 0; i < values.Len(); i++ {
		p = append(p, "?")
		wrapper.args = append(wrapper.args, values.Index(i).Interface())
	}
	sqlBuild.WriteString(strings.Join(p, ", "))
	sqlBuild.WriteString(")")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	return wrapper
}

// Lt 小于
func (wrapper *Wrapper) Lt(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Lt)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// Le 小于等于
func (wrapper *Wrapper) Le(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Le)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// Gt 大于
func (wrapper *Wrapper) Gt(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Gt)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// Ge 大于等于
func (wrapper *Wrapper) Ge(column string, value any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Ge)
	sqlBuild.WriteString(" ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value)
	return wrapper
}

// Between 之间
func (wrapper *Wrapper) Between(column string, value1 any, value2 any) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(Between)
	sqlBuild.WriteString(" ? AND ?")
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	wrapper.args = append(wrapper.args, value1, value2)
	return wrapper
}

// IsNotNull 不为空
func (wrapper *Wrapper) IsNotNull(column string) *Wrapper {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString(column)
	sqlBuild.WriteString(" ")
	sqlBuild.WriteString(IsNotNull)
	wrapper.wheres = append(wrapper.wheres, sqlBuild.String())
	return wrapper
}

// GetSql 获取SQL语句
func (wrapper *Wrapper) GetSql() string {
	sqlBuild := strings.Builder{}
	sqlBuild.WriteString("SELECT ")
	sqlBuild.WriteString(wrapper.GetSelectSql())
	sqlBuild.WriteString(" FROM ")
	sqlBuild.WriteString(wrapper.GetFromSql())
	if len(wrapper.wheres) != 0 {
		sqlBuild.WriteString(" WHERE ")
		sqlBuild.WriteString(wrapper.GetWhereSql())
	}
	if len(wrapper.groups) != 0 {
		sqlBuild.WriteString(" GROUP BY ")
		sqlBuild.WriteString(wrapper.GetGroupSql())
	}
	if len(wrapper.orders) != 0 {
		sqlBuild.WriteString(" ORDER BY ")
		sqlBuild.WriteString(wrapper.GetOrderSql())
	}
	if len(wrapper.last) != 0 {
		sqlBuild.WriteString(" ")
		sqlBuild.WriteString(wrapper.last)
	}
	return sqlBuild.String()
}

// GetSelectSql 获取查询语句
func (wrapper *Wrapper) GetSelectSql() string {
	return strings.Join(wrapper.selects, ", ")
}

// GetFilters 获取查询过滤
func (wrapper *Wrapper) GetFilters() []string {
	return wrapper.filters
}

// GetFromSql 获取 FROM 语句
func (wrapper *Wrapper) GetFromSql() string {
	return wrapper.from
}

// GetWhereSql 获取条件语句
func (wrapper *Wrapper) GetWhereSql() string {
	return strings.Join(wrapper.wheres, " "+And+" ")
}

// GetGroupSql 获取分组语句
func (wrapper *Wrapper) GetGroupSql() string {
	return strings.Join(wrapper.groups, ", ")
}

// GetOrderSql 获取排序语句
func (wrapper *Wrapper) GetOrderSql() string {
	if len(wrapper.orders) != 0 {
		return strings.Join(wrapper.orders, ", ")
	}
	if len(wrapper.defaultOrders) != 0 {
		return strings.Join(wrapper.defaultOrders, ", ")
	}
	return ""
}

// GetArgs 获取参数
func (wrapper *Wrapper) GetArgs() []any {
	return wrapper.args
}
