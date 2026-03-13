package repository

import (
	"context"
	"rest_demo/pkg/db"

	"gorm.io/gorm"
)

const ctxTxKey = "TxKey"

type Repository struct {
	db *db.MsDb
}

func NewRepository(
	db *db.MsDb,
) *Repository {
	return &Repository{
		db: db,
	}
}

// DB return tx
// If you need to create a Transaction, you must call DB(ctx) and Transaction(ctx,fn)
func (r *Repository) DB(ctx context.Context) *gorm.DB {
	v := ctx.Value(ctxTxKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx
		}
	}
	return r.db.Master().WithContext(ctx)
}

func (r *Repository) ReadDB(ctx context.Context) *gorm.DB {
	return r.db.Slave().WithContext(ctx)
}

func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.Master().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, ctxTxKey, tx)
		return fn(ctx)
	})
}
