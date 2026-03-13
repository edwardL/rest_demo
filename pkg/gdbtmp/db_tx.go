package gdbtmp

import (
	"database/sql"
	"errors"
)

type Tx struct {
	*sql.Tx
}

// DbTx 事务结构体
type DbTx struct {
	*Db
}

// Transaction 重写事务拦截嵌套
func (t *DbTx) Transaction(f func(db *DbTx) error) (err error) {
	return errors.New("不支持嵌套事务")
}

// GetTx 获取事务对象
func (t *DbTx) GetTx() *Tx {
	return t.sqlTx
}

// Begin 开启事务
func Begin() (*Tx, error) {
	if err != nil {
		return nil, err
	}
	return &Tx{
		Tx: begin,
	}, nil
}
