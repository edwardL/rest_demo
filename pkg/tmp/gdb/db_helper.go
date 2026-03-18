package gdb

import (
	"context"
	"database/sql"
)

// Exec 执行更新语句
func Exec(execSql string, args ...any) (affected int64, insertId int64, err error) {
	var ctx = context.Background()
	var dbRes = New().Ctx(ctx).LogCallDepth(1).RawExec(execSql, args...)
	if dbRes.Error != nil {
		return 0, 0, dbRes.Error
	}
	return dbRes.ExecResult.AffectRecs, dbRes.ExecResult.LastAffectId, nil
}

// TExec 事务执行更新语句
func TExec(tx *sql.Tx, execSql string, args ...any) (affected int64, insertId int64, err error) {
	var ctx = context.Background()
	var dbRes = New().Ctx(ctx).Tx(tx).LogCallDepth(1).RawExec(execSql, args...)
	if dbRes.Error != nil {
		return 0, 0, dbRes.Error
	}
	return dbRes.ExecResult.AffectRecs, dbRes.ExecResult.LastAffectId, nil
}

// ExecContext 执行更新语句
func ExecContext(ctx context.Context, execSql string, args ...any) (affected int64, insertId int64, err error) {
	var dbRes = New().Ctx(ctx).LogCallDepth(1).RawExec(execSql, args...)
	if dbRes.Error != nil {
		return 0, 0, dbRes.Error
	}
	return dbRes.ExecResult.AffectRecs, dbRes.ExecResult.LastAffectId, nil
}

// TExecContext 执行更新语句
func TExecContext(ctx context.Context, tx *sql.Tx, execSql string, args ...any) (affected int64, insertId int64, err error) {
	var dbRes = New().Ctx(ctx).Tx(tx).LogCallDepth(1).RawExec(execSql, args...)
	if dbRes.Error != nil {
		return 0, 0, dbRes.Error
	}
	return dbRes.ExecResult.AffectRecs, dbRes.ExecResult.LastAffectId, nil
}

// Query 原生查询 带类型
func Query[T any](query string, args ...any) ([]*T, error) {
	var ctx = context.Background()
	return Model[T]().Ctx(ctx).LogCallDepth(1).Raw(query, args...).Select()
}

// TQuery 原生查询 带类型
func TQuery[T any](tx *sql.Tx, query string, args ...any) ([]*T, error) {
	var ctx = context.Background()
	return Model[T]().Tx(tx).Ctx(ctx).LogCallDepth(1).Raw(query, args...).Select()
}

// QueryContext 原生查询 带类型
func QueryContext[T any](ctx context.Context, query string, args ...any) ([]*T, error) {
	return Model[T]().Ctx(ctx).LogCallDepth(1).Raw(query, args...).Select()
}

// TQueryContext 原生查询 带类型
func TQueryContext[T any](ctx context.Context, tx *sql.Tx, query string, args ...any) ([]*T, error) {
	return Model[T]().Tx(tx).Ctx(ctx).LogCallDepth(1).Raw(query, args...).Select()
}

// QueryOne 原生查询 带类型
func QueryOne[T any](query string, args ...any) (*T, error) {
	var ctx = context.Background()
	return Model[T]().Ctx(ctx).LogCallDepth(1).Raw(query, args...).One()
}

// TQueryOne 原生查询 带类型
func TQueryOne[T any](tx *sql.Tx, query string, args ...any) (*T, error) {
	var ctx = context.Background()
	return Model[T]().Tx(tx).Ctx(ctx).LogCallDepth(1).Raw(query, args...).One()
}

// QueryOneContext 原生查询 带类型
func QueryOneContext[T any](ctx context.Context, query string, args ...any) (*T, error) {
	return Model[T]().Ctx(ctx).LogCallDepth(1).Raw(query, args...).One()
}

// TQueryOneContext 原生查询 带类型
func TQueryOneContext[T any](ctx context.Context, tx *sql.Tx, query string, args ...any) (*T, error) {
	return Model[T]().Tx(tx).Ctx(ctx).LogCallDepth(1).Raw(query, args...).One()
}

// Create 新增数据
// data = Struct
func Create[T any](data *T) (*T, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Create(data)
	return res.Model, res.Error
}

// CreateContext 新增数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func CreateContext[T any](ctx context.Context, data *T) (*T, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Create(data)
	return res.Model, res.Error
}

// TCreate 新增数据
// data = Struct
func TCreate[T any](tx *sql.Tx, data *T) (*T, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Create(data)
	return res.Model, res.Error
}

// TCreateContext 新增数据
// data = Struct
func TCreateContext[T any](ctx context.Context, tx *sql.Tx, data *T) (*T, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Create(data)
	return res.Model, res.Error
}

// CreateInBatches 新增数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func CreateInBatches[T any](data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).CreateInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// CreateInBatchesContext 新增数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func CreateInBatchesContext[T any](ctx context.Context, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).CreateInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TCreateInBatches 新增数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func TCreateInBatches[T any](tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).CreateInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TCreateInBatchesContext 新增数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func TCreateInBatchesContext[T any](ctx context.Context, tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).CreateInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// Replace 新增或替换数据
// data = Struct
func Replace[T any](data *T) (*T, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Replace(data)
	return res.Model, res.Error
}

// ReplaceContext 新增或替换数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func ReplaceContext[T any](ctx context.Context, data *T) (*T, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Replace(data)
	return res.Model, res.Error
}

// TReplace 新增或替换数据
// data = Struct
func TReplace[T any](tx *sql.Tx, data *T) (*T, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Replace(data)
	return res.Model, res.Error
}

// TReplaceContext 新增或替换数据
// data = Struct
func TReplaceContext[T any](ctx context.Context, tx *sql.Tx, data *T) (*T, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Replace(data)
	return res.Model, res.Error
}

// ReplaceInBatches 新增或替换数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func ReplaceInBatches[T any](data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).ReplaceInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// ReplaceInBatchesContext 新增或替换数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func ReplaceInBatchesContext[T any](ctx context.Context, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).ReplaceInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TReplaceInBatches 新增或替换数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func TReplaceInBatches[T any](tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).ReplaceInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TReplaceInBatchesContext 新增或替换数据
// data = Struct
// data = *Db insert init ... select  从查询新增
func TReplaceInBatchesContext[T any](ctx context.Context, tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).ReplaceInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// Update 更新数据
// data = Struct
// whereField 默认 [id,ts]
func Update[T any](data *T, whereField ...[]string) (*T, error) {
	if len(whereField) == 0 || len(whereField[0]) == 0 {
		whereField = [][]string{
			{"id", "ts"},
		}
	}
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Updates(data, whereField...)
	return res.Model, res.Error
}

// UpdateContext 更新数据
// data = Struct
func UpdateContext[T any](ctx context.Context, data *T, whereField ...[]string) (*T, error) {
	if len(whereField) == 0 || len(whereField[0]) == 0 {
		whereField = [][]string{
			{"id", "ts"},
		}
	}
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Updates(data, whereField...)
	return res.Model, res.Error
}

// TUpdate 更新数据
// data = Struct
func TUpdate[T any](tx *sql.Tx, data *T, whereField ...[]string) (*T, error) {
	if len(whereField) == 0 || len(whereField[0]) == 0 {
		whereField = [][]string{
			{"id", "ts"},
		}
	}
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Updates(data, whereField...)
	return res.Model, res.Error
}

// TUpdateContext 更新数据
// data = Struct
func TUpdateContext[T any](ctx context.Context, tx *sql.Tx, data *T, whereField ...[]string) (*T, error) {
	if len(whereField) == 0 || len(whereField[0]) == 0 {
		whereField = [][]string{
			{"id", "ts"},
		}
	}
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Updates(data, whereField...)
	return res.Model, res.Error
}

// Save 更新多个字段 Struct
func Save[T any](data *T) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Save(data)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// SaveContext 更新多个字段 Struct
func SaveContext[T any](ctx context.Context, data *T) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).Save(data)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TSave 更新多个字段 Struct
func TSave[T any](tx *sql.Tx, data *T) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Save(data)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TSaveContext 更新多个字段 Struct
func TSaveContext[T any](ctx context.Context, tx *sql.Tx, data *T) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).Save(data)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// SaveInBatches 更新多个字段 Struct
func SaveInBatches[T any](data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).SaveInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// SaveInBatchesContext 更新多个字段 Struct
func SaveInBatchesContext[T any](ctx context.Context, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).LogCallDepth(1).SaveInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TSaveInBatches 更新多个字段 Struct
func TSaveInBatches[T any](tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var ctx = context.Background()
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).SaveInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}

// TSaveInBatchesContext 更新多个字段 Struct
func TSaveInBatchesContext[T any](ctx context.Context, tx *sql.Tx, data []*T, batchSize ...int) (int64, int64, error) {
	var res = Model[T]().Ctx(ctx).Tx(tx).LogCallDepth(1).SaveInBatches(data, batchSize...)
	return res.ExecResult.AffectRecs, res.ExecResult.LastAffectId, res.Error
}
