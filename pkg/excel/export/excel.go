package export

import (
	"archive/zip"
	"context"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"path/filepath"
	"reflect"
	"sync"
)

// DefaultSheetName 默认操作表
const DefaultSheetName = "Sheet1"

var _ Exporter = (*Excel)(nil)

type Excel struct {
	options *options
	columns *columns
	dp      DataProvider
}

func NewExcel(h Headers, dp DataProvider, opts ...Option) *Excel {
	e := &Excel{
		dp:      dp,
		columns: newColumns(h),
		options: newOptions(opts...),
	}
	return e
}

// Export 导出到本地文件，返回本地文件路径
func (e *Excel) Export() (string, error) {
	ef, err := e.export()
	if err != nil {
		return "", err
	}
	defer func() {
		if err = ef.Close(); err != nil {
			log.Println("excel Export() close err:", err)
		}
	}()
	return ef.Save()
}

// ExportTo 导出到io.Writer
func (e *Excel) ExportTo(w io.Writer) (n int64, err error) {
	ef, err := e.export()
	if err != nil {
		return
	}
	defer func() {
		if err = ef.Close(); err != nil {
			log.Println("excel Export() close err:", err)
		}
	}()
	return ef.WriteTo(w)
}

// ExportToStorage 导出到文件存储，返回下载地址
func (e *Excel) ExportToStorage(fs FileStorage) (string, error) {
	ef, err := e.export()
	if err != nil {
		return "", err
	}
	defer func() {
		if err = ef.Close(); err != nil {
			log.Println("excel Export() close err:", err)
		}
	}()
	fk := filepath.Base(ef.Filepath())
	fr, fw := io.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(2)
	var _err error
	go func() {
		defer wg.Done()
		if _, _err = ef.WriteTo(fw); _err != nil {
			log.Println("io pipe write error", _err.Error())
		}
		_ = fw.Close()
	}()
	go func() {
		defer wg.Done()
		if _err = fs.PutStream(context.Background(), fk, fr); _err != nil {
			log.Println("io pipe read error", _err.Error())
		}
		_ = fr.Close()
	}()
	wg.Wait()
	if _err != nil {
		return "", _err
	}
	return fs.Url(fk), nil
}

func (e *Excel) export() (ef exportFile, err error) {
	if e.options.forceZip {
		return e.exportZip(nil)
	}
	//先导出第一个文件
	firstFile := excelize.NewFile()
	hasMore, err := e.exportToExcelize(firstFile)
	if err != nil {
		_ = firstFile.Close()
		return nil, err
	}
	if !hasMore {
		return newExportExcel(getFilename(e.options.filename, 0, ExcelSuffix), firstFile), nil
	}
	defer func() {
		_ = firstFile.Close()
	}()
	//导出zip
	return e.exportZip(firstFile)
}

func (e *Excel) exportZip(firstFile *excelize.File) (exportFile, error) {
	var idx int
	ef, err := newExportTmpFile(getFilename(e.options.filename, idx, ZipSuffix))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = ef.Close()
		}
	}()
	zw := zip.NewWriter(ef)
	defer func() {
		_ = zw.Close()
	}()
	var w io.Writer
	//把外面传进来的加进去
	if firstFile != nil {
		w, err = newZipWriter(zw, getFilename(e.options.filename, 0, ExcelSuffix))
		if err != nil {
			return nil, err
		}
		if err = firstFile.Write(w); err != nil {
			return nil, err
		}
		idx++
	}
	//读取数据
	hasMore := true
	for {
		if !hasMore {
			break
		}
		w, err = newZipWriter(zw, getFilename(e.options.filename, idx, ExcelSuffix))
		if err != nil {
			return nil, err
		}
		idx++
		fw := excelize.NewFile()
		hasMore, err = e.exportToExcelize(fw)
		if err != nil {
			_ = fw.Close()
			return nil, err
		}
		//写入zip
		if err = fw.Write(w); err != nil {
			_ = fw.Close()
			return nil, err
		}
		_ = fw.Close()
	}
	return ef, nil
}

func (e *Excel) exportToExcelize(fp *excelize.File) (hasMore bool, err error) {
	fw, err := fp.NewStreamWriter(DefaultSheetName)
	if err != nil {
		return
	}
	row := e.options.rowStart + 1
	col := e.options.colStart + 1
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return
	}
	if err = fw.SetRow(cell, e.columns.titles); err != nil {
		return
	}
	for {
		row++
		if !e.dp.Next() {
			break
		}
		_v := e.dp.Value()
		values := e.processRow(reflect.ValueOf(_v), row)
		cell, err = excelize.CoordinatesToCellName(col, row)
		if err != nil {
			log.Println(err)
			break
		}
		err = fw.SetRow(cell, values)
		if err != nil {
			log.Println(err)
			break
		}
		if !e.options.forceSingleFile && row-e.options.rowStart-1 >= e.options.singleFileMaxRows {
			hasMore = true
			break
		}
	}
	if err = fw.Flush(); err != nil {
		return
	}
	return
}

func (e *Excel) TestProcessRow(rowData reflect.Value, row int) []any {
	return e.processRow(rowData, row)
}

func (e *Excel) processRow(rowData reflect.Value, row int) []any {
	switch rowData.Type().Kind() {
	case reflect.Ptr:
		return e.processRow(rowData.Elem(), row)
	case reflect.Map:
		return e.processRowFromMap(rowData, row)
	case reflect.Struct:
		return e.processRowFromStruct(rowData, row)
	case reflect.Slice:
		return e.processRowFromSlice(rowData, row)
	default:
		return make([]any, e.columns.nums)
	}
}

func (e *Excel) processRowFromMap(rowData reflect.Value, row int) []any {
	_rowData := make([]any, e.columns.nums)
	for i := range e.columns.fields {
		field := e.columns.fields[i]
		val := rowData.MapIndex(reflect.ValueOf(field))
		_rowData[i] = e.processCell(field, val, rowData, row, e.options.colStart+i+1)
	}
	return _rowData
}

func (e *Excel) processRowFromStruct(rowData reflect.Value, row int) []any {
	_rowData := make([]any, e.columns.nums)
	typ := rowData.Type()
	gets := 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		key := field.Name
		//读取tag
		if k, ok := typ.Field(i).Tag.Lookup(TagName); ok {
			key = k
		}
		if idx, ok := e.columns.keyIndex[key]; ok {
			gets++
			_rowData[idx] = e.processCell(key, rowData.Field(i), rowData, row, e.options.colStart+idx+1)
		}
		if gets == e.columns.nums {
			break
		}
	}
	return _rowData
}

func (e *Excel) processRowFromSlice(rowData reflect.Value, row int) []any {
	_rowData := make([]any, e.columns.nums)
	for i := 0; i < e.columns.nums; i++ {
		_rowData[i] = e.processCell(e.columns.fields[i], rowData.Index(i), rowData, row, e.options.colStart+i+1)
	}
	return _rowData
}

func (e *Excel) processCell(field string, val reflect.Value, rowData reflect.Value, row, col int) any {
	var v any
	if val.IsValid() {
		v = val.Interface()
	}
	if e.columns.columnRenders[field] != nil {
		return e.columns.columnRenders[field](rowData, v, row, col)
	}
	return v
}
