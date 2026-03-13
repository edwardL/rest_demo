package export

import (
	"context"
	"gorm.io/gorm"
	"time"
)

// GormDataProvider Gorm查询数据迭代器
type GormDataProvider struct {
	tx       *gorm.DB
	offset   int
	limit    int
	hasMore  bool
	callback GormDataProviderCallback
	sliceDp  *SliceDataProvider
}

type GormDataProviderCallback func([]map[string]any) []any

type GormDataProviderOption func(provider *GormDataProvider)

func WithGormDataProviderCallback(cb GormDataProviderCallback) GormDataProviderOption {
	return func(provider *GormDataProvider) {
		if cb != nil {
			provider.callback = cb
		}
	}
}

func NewGormDataProvider(tx *gorm.DB, opts ...GormDataProviderOption) *GormDataProvider {
	g := &GormDataProvider{
		tx:       tx,
		offset:   0,
		hasMore:  true,
		limit:    2000,
		callback: defaultCallback,
		sliceDp:  NewSliceDataProvider([]any{}),
	}
	for i := range opts {
		opts[i](g)
	}
	return g
}

func (dp *GormDataProvider) Next() bool {
	if !dp.hasMore {
		return false
	}
	hasNext := dp.sliceDp.Next()
	if hasNext {
		return hasNext
	}
	res := make([]map[string]any, 0)
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

func (dp *GormDataProvider) Value() any {
	return dp.sliceDp.Value()
}

func defaultCallback(d []map[string]any) []any {
	_data := make([]any, len(d))
	for i, v := range d {
		_data[i] = v
	}
	return _data
}
