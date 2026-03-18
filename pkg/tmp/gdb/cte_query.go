package gdb

import "errors"

// VirtualTable 虚拟表 条件
type VirtualTable struct {
	TableAlias string
	Sql        *Sql
}

// CteQuery 公共表表达式
type CteQuery struct {
	Cte          string
	err          error
	VirtualTable []*VirtualTable
}

type DbFace interface {
	GetGSql() *Sql
}

// NewCteQuery 创建虚拟查询
func NewCteQuery() *CteQuery {
	return &CteQuery{
		VirtualTable: make([]*VirtualTable, 0),
	}
}

// SetCet 虚拟表查询
func (vt *CteQuery) SetCet(cte string) *CteQuery {
	vt.Cte = cte
	return vt
}

// SetCteTable 虚拟表查询
// WITH
//
//	临时表1名称 [(列名1, 列名2, ...)] AS (SELECT/INSERT/UPDATE/DELETE 语句),
//	{tableAlias} AS ({dbQuery *Db|*Sql|*DbT[T]}),
func (vt *CteQuery) SetCteTable(tableAlias string, dbQuery any) *CteQuery {
	if vt.err != nil {
		return vt
	}
	var dbSql *Sql
	switch t := dbQuery.(type) {
	case *Sql:
		dbSql = t
	case DbFace:
		dbSql = t.GetGSql()
	default:
		vt.err = errors.New("dbQuery Type must be *Db|*Sql|*DbT[T]")
		return vt
	}
	vt.VirtualTable = append(vt.VirtualTable, &VirtualTable{
		TableAlias: tableAlias,
		Sql:        dbSql,
	})
	return vt
}
