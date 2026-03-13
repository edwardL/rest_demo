package export

import (
	"context"
	"io"
	"time"
)

const TagName = "export"         // export 导出字段的tag
const SingleFileMaxRows = 100000 //单个文件最大数据量
const MaxRows = 1000000          //最大导出数据
const Timeout = time.Minute * 30 //最大导出执行时间

// ExcelSuffix CsvSuffix 导出文件后缀
const ExcelSuffix = "xlsx"
const CsvSuffix = "csv"
const ZipSuffix = "zip"

type FileStorage interface {
	PutStream(ctx context.Context, filename string, rs io.Reader) error
	Url(fileKey string) string
}

// Exporter 导出接口
type Exporter interface {
	// Export 导出到本地文件，返回本地文件路径
	Export() (string, error)
	// ExportTo 导出到io.Writer
	ExportTo(at io.Writer) (int64, error)
	// ExportToStorage 导出到文件存储，返回下载地址
	ExportToStorage(fileStorage FileStorage) (string, error)
}

// ToExcelStream 导出excel的快捷方法
func ToExcelStream(h Headers, dp DataProvider, w io.Writer, opt ...Option) (int64, error) {
	return NewExcel(h, dp, opt...).ExportTo(w)
}

// ToExcelFile 导出excel的快捷方法
func ToExcelFile(h Headers, dp DataProvider, opt ...Option) (string, error) {
	return NewExcel(h, dp, opt...).Export()
}

// ToExcelStorage 导出excel到oss的快捷方法
func ToExcelStorage(h Headers, dp DataProvider, fs FileStorage, opt ...Option) (string, error) {
	return NewExcel(h, dp, opt...).ExportToStorage(fs)
}

// ToCsvStream 导出csv的快捷方法
func ToCsvStream(h Headers, dp DataProvider, w io.Writer, opt ...Option) (int64, error) {
	return NewCsv(h, dp, opt...).ExportTo(w)
}

// ToCsvFile 导出csv的快捷方法
func ToCsvFile(h Headers, dp DataProvider, opt ...Option) (string, error) {
	return NewCsv(h, dp, opt...).Export()
}

// ToCsvStorage 导出csv到oss的快捷方法
func ToCsvStorage(h Headers, dp DataProvider, fs FileStorage, opt ...Option) (string, error) {
	return NewCsv(h, dp, opt...).ExportToStorage(fs)
}
