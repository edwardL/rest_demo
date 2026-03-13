package orm2

import (
	"fmt"
	"rest_demo/pkg/utils/arrays"
)

// Page 分页对象
type Page[T any] struct {
	CurrPage  int  // 当前页数
	PageNums  int  // 分页数量
	TotalNums int  // 总条数
	PageData  []*T // 分页数据
}

// NewPage 创建分页对象
func NewPage[T any](currPage int, pageNums int) Page[T] {
	return Page[T]{
		CurrPage:  currPage,
		PageNums:  pageNums,
		TotalNums: 0,
		PageData:  nil,
	}
}

// MapPage 转换分页对象
func MapPage[T any, R any](t Page[T], f func(*T) *R) Page[R] {
	return Page[R]{
		CurrPage:  t.CurrPage,
		PageNums:  t.PageNums,
		TotalNums: t.TotalNums,
		PageData:  arrays.Map(t.PageData, f),
	}
}

// getLimitSql 获取分页SQL
func (page Page[T]) getLimitSql() string {
	pointer := (page.CurrPage - 1) * page.PageNums
	return fmt.Sprintf("LIMIT %d, %d", pointer, page.PageNums)
}
