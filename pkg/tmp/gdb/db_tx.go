package gdb

import (
	"database/sql"
	"errors"
	hhdb "nwgit.gzhhit.com/BD/hhitdb.git"
)

// DbTx 事务结构体
type DbTx struct {
	*Db
}

// Table 表名 或者 子查询
// db.Table("table_name")
// db.Table("(?) tb",db.Table("tb"))
// db.Table(*Struct{}) 模型结构体
func (t *DbTx) Table(table any, args ...any) *DbTx {
	t.Db = t.Db.tableClone(table, args...)
	return t
}

// Transaction 重写事务拦截嵌套
func (t *DbTx) Transaction(f func(db *DbTx) error) (err error) {
	return errors.New("不支持嵌套事务")
}

// GetTx 获取事务对象
func (t *DbTx) GetTx() *sql.Tx {
	return t.sqlTx
}

// Begin 开启事务
func Begin() (*sql.Tx, error) {
	var begin, err = hhdb.GetDBOP().GetDBObj().Begin()
	if err != nil {
		return nil, err
	}
	return begin, nil
}
