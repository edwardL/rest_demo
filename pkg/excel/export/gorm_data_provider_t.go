package export

import (
	"context"
	"gorm.io/gorm"
	"time"
)

// GormDataProviderT Gorm查询数据迭代器
type GormDataProviderT[T any] struct {
	tx       *gorm.DB
	offset   int
	limit    int
	hasMore  bool
	callback GormDataProviderTCallback[T]
	sliceDp  *SliceDataProvider
}

type GormDataProviderTCallback[T any] func([]T) []any

type GormDataProviderTOption[T any] func(provider *GormDataProviderT[T])

func WithGormDataProviderTCallback[T any](cb GormDataProviderTCallback[T]) GormDataProviderTOption[T] {
	return func(provider *GormDataProviderT[T]) {
		if cb != nil {
			provider.callback = cb
		}
	}
}

func NewGormDataProviderT[T any](tx *gorm.DB, opts ...GormDataProviderTOption[T]) *GormDataProviderT[T] {
	g := &GormDataProviderT[T]{
		tx:       tx,
		offset:   0,
		hasMore:  true,
		limit:    2000,
		callback: defaultCallbackT[T],
		sliceDp:  NewSliceDataProvider([]any{}),
	}
	for i := range opts {
		opts[i](g)
	}
	return g
}

func (dp *GormDataProviderT[T]) Next() bool {
	if !dp.hasMore {
		return false
	}
	hasNext := dp.sliceDp.Next()
	if hasNext {
		return hasNext
	}
	res := make([]T, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	err := dp.tx.WithContext(ctx).Offset(dp.offset).Limit(dp.limit).Scan(&res).Error
	if err != nil {
		dp.hasMore = false
		return false
	}
	if len(res) == 0 {
		dp.hasMore = false
		return false
	}
	dp.offset += dp.limit
	dp.sliceDp = NewSliceDataProvider(dp.callback(res))
	return dp.sliceDp.Next()
}

func (dp *GormDataProviderT[T]) Value() any {
	return dp.sliceDp.Value()
}

func defaultCallbackT[T any](d []T) []any {
	_data := make([]any, len(d))
	for i, v := range d {
		_data[i] = v
	}
	return _data
}
