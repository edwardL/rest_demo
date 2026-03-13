package export

import (
	"github.com/xuri/excelize/v2"
	"reflect"
)

// DataProvider 数据提供者
type DataProvider interface {
	Next() bool
	Value() any
}

type CellRender func(rowData reflect.Value, v any, row int, col int) any

// Header 表头
type Header struct {
	Field      string
	Title      string
	CellRender CellRender      //单元格数据处理
	ColStyle   *excelize.Style //列格式,导出excel时支持
}

// Headers 表头
type Headers []Header
