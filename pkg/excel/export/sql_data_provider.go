package export

import (
	"gorm.io/gorm"
)

type SqlDataProvider struct {
	tx      *gorm.DB
	offset  int
	limit   int
	hasMore bool
	sliceDp *SliceDataProvider
}

func NewSqlDataProvider(db *gorm.DB, selectSql string) *SqlDataProvider {
	return &SqlDataProvider{
		tx:      db.Raw(selectSql),
		offset:  0,
		hasMore: true,
		limit:   2000,
		sliceDp: NewSliceDataProvider([]any{}),
	}
}

func (dp *SqlDataProvider) Next() bool {
	if !dp.hasMore {
		return false
	}
	hasNext := dp.sliceDp.Next()
	if hasNext {
		return hasNext
	}
	res := make([]map[string]any, 0)
	err := dp.tx.Offset(dp.offset).Limit(dp.limit).Scan(&res).Error
	if err != nil {
		dp.hasMore = false
		return false
	}
	if len(res) == 0 {
		dp.hasMore = false
		return false
	}
	dp.offset += dp.limit
	_data := make([]any, len(res))
	for i, v := range res {
		_data[i] = v
	}
	dp.sliceDp = NewSliceDataProvider(_data)
	return dp.sliceDp.Next()
}

func (dp *SqlDataProvider) Value() any {
	return dp.sliceDp.Value()
}
