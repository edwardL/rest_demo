package gdb

import (
	"context"
	"time"
)

// 暴漏内部变量给构建对象

// GetCtx 获取上下文context
func (d *Sql) GetCtx() context.Context {
	return d.ctx
}

// GetErr 获取错误信息
func (d *Sql) GetErr() error {
	return d.err
}

// GetTableAlias 获取表别名
func (d *Sql) GetTableAlias() string {
	return d.tableAlias
}

// GetArgs 获取SQL参数列表
func (d *Sql) GetArgs() []any {
	return d.args
}

// GetFields 获取字段参数GenCtrl
func (d *Sql) GetFields() *GenCtrl {
	return d.fields
}

// GetRawSql 获取原生SQL参数GenCtrl
func (d *Sql) GetRawSql() *GenCtrl {
	return d.rawSql
}

// GetCteQuery 获取CTE复杂查询参数
func (d *Sql) GetCteQuery() *CteQuery {
	return d.cteQuery
}

// GetUnion 获取Union参数列表
func (d *Sql) GetUnion() []*GenCtrl {
	return d.union
}

// GetWhereCtrl 获取Where条件参数列表
func (d *Sql) GetWhereCtrl() []*GenCtrl {
	return d.whereCtrl
}

// GetJoinCtrl 获取连表参数列表
func (d *Sql) GetJoinCtrl() []*GenCtrl {
	return d.joinCtrl
}

// GetLimitCtrl 获取分页Limit参数
func (d *Sql) GetLimitCtrl() *GenCtrl {
	return d.limitCtrl
}

// GetOffsetCtrl 获取分页Offset参数
func (d *Sql) GetOffsetCtrl() *GenCtrl {
	return d.offsetCtrl
}

// GetGroup 获取分组参数
func (d *Sql) GetGroup() *GenCtrl {
	return d.group
}

// GetOrder 获取排序参数
func (d *Sql) GetOrder() *GenCtrl {
	return d.order
}

// GetDbConvInitPtr 获取初始化指针标记
func (d *Sql) GetDbConvInitPtr() bool {
	return d.dbConvInitPtr
}

// GetTimeLocation 获取时间时区
func (d *Sql) GetTimeLocation() *time.Location {
	return d.timeLocation
}

// GetFieldReplace 获取字段名替换映射
func (d *Sql) GetFieldReplace() map[string]string {
	return d.fieldReplace
}

// GetFieldSelect 获取指定更新/新增字段映射
func (d *Sql) GetFieldSelect() map[string]string {
	return d.fieldSelect
}

// GetFieldOmit 获取忽略更新/新增字段映射
func (d *Sql) GetFieldOmit() map[string]string {
	return d.fieldOmit
}

// GetIgnoreDuplicate 获取忽略重复字段标记映射
func (d *Sql) GetIgnoreDuplicate() map[string]bool {
	return d.ignoreDuplicate
}

// GetSetDuplicate 获取自定义重复更新规则映射
func (d *Sql) GetSetDuplicate() map[string]string {
	return d.setDuplicate
}
