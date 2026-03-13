package export

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"github.com/spf13/cast"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

var _ Exporter = (*Csv)(nil)

type Csv struct {
	dp      DataProvider
	options *options
	columns *columns
}

func NewCsv(h Headers, dp DataProvider, opts ...Option) *Csv {
	e := &Csv{
		dp:      dp,
		columns: newColumns(h),
		options: newOptions(opts...),
	}
	return e
}

func (c *Csv) Export() (filename string, err error) {
	ef, err := c.export()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = ef.Close()
	}()
	return ef.Save()
}

func (c *Csv) ExportTo(w io.Writer) (n int64, err error) {
	ef, err := c.export()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = ef.Close()
	}()
	return ef.WriteTo(w)
}

func (c *Csv) ExportToStorage(fs FileStorage) (filename string, err error) {
	ef, err := c.export()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = ef.Close()
	}()
	fileKey := filepath.Base(ef.Filepath())
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
		if _err = fs.PutStream(context.Background(), fileKey, fr); _err != nil {
			log.Println("io pipe read error", _err.Error())
		}
		_ = fr.Close()
	}()
	wg.Wait()
	if _err != nil {
		return "", _err
	}
	return fs.Url(fileKey), nil
}

func (c *Csv) export() (exportFile, error) {
	if c.options.forceZip {
		return c.exportZip(nil)
	}
	//先导出第一个文件
	firstEf, err := newExportTmpFile(getFilename(c.options.filename, 0, CsvSuffix))
	if err != nil {
		return nil, err
	}
	hasMore, err := c.exportToWrite(firstEf)
	if err != nil {
		_ = firstEf.Close()
		return nil, err
	}
	if !hasMore {
		return firstEf, nil
	}
	defer func() {
		_ = firstEf.Close()
		_ = os.Remove(firstEf.Filepath())
	}()
	//导出zip
	return c.exportZip(firstEf)
}

func (c *Csv) exportZip(firstFile exportFile) (exportFile, error) {
	var idx int
	ef, err := newExportTmpFile(getFilename(c.options.filename, 0, ZipSuffix))
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
		w, err = newZipWriter(zw, firstFile.Filepath())
		if err != nil {
			return nil, err
		}
		if _, err = firstFile.WriteTo(w); err != nil {
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
		w, err = newZipWriter(zw, getFilename(c.options.filename, idx, CsvSuffix))
		if err != nil {
			return nil, err
		}
		idx++
		hasMore, err = c.exportToWrite(w)
		if err != nil {
			return nil, err
		}
	}
	return ef, nil
}

func (c *Csv) exportToWrite(fp io.Writer) (bool, error) {
	fw := csv.NewWriter(fp)
	defer func() {
		fw.Flush()
	}()
	// 写入CSV头部
	err := fw.Write(c.columns.titleStr)
	if err != nil {
		return false, err
	}
	row := 0
	for {
		row++
		if !c.dp.Next() {
			break
		}
		_v := c.dp.Value()
		values := c.processRow(reflect.ValueOf(_v), row)
		err = fw.Write(values)
		if err != nil {
			break
		}
		if !c.options.forceSingleFile && row >= c.options.singleFileMaxRows {
			return true, nil
		}
	}
	return false, err
}

func (c *Csv) processRow(rowData reflect.Value, row int) []string {
	switch rowData.Type().Kind() {
	case reflect.Ptr:
		return c.processRow(rowData.Elem(), row)
	case reflect.Map:
		return c.processRowFromMap(rowData, row)
	case reflect.Struct:
		return c.processRowFromStruct(rowData, row)
	case reflect.Slice:
		return c.processRowFromSlice(rowData, row)
	default:
		return make([]string, c.columns.nums)
	}
}

func (c *Csv) processRowFromMap(rowData reflect.Value, row int) []string {
	_rowData := make([]string, c.columns.nums)
	for i := range c.columns.fields {
		field := c.columns.fields[i]
		_rowData[i] = c.processCell(field, rowData.MapIndex(reflect.ValueOf(field)), rowData, row, i+1)
	}
	return _rowData
}

func (c *Csv) processRowFromStruct(rowData reflect.Value, row int) []string {
	_rowData := make([]string, c.columns.nums)
	typ := rowData.Type()
	gets := 0
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		key := field.Name
		//读取tag
		if k, ok := typ.Field(i).Tag.Lookup(TagName); ok {
			key = k
		}
		if idx, ok := c.columns.keyIndex[key]; ok {
			gets++
			_rowData[idx] = c.processCell(key, rowData.Field(i), rowData, row, idx+1)
		}
		if gets == c.columns.nums {
			break
		}
	}
	return _rowData
}

func (c *Csv) processRowFromSlice(rowData reflect.Value, row int) []string {
	_rowData := make([]string, c.columns.nums)
	for i := 0; i < c.columns.nums; i++ {
		_rowData[i] = c.processCell(c.columns.fields[i], rowData, rowData.Index(i), row, i+1)
	}
	return _rowData
}

func (c *Csv) processCell(field string, val reflect.Value, rowData reflect.Value, row, col int) string {
	var v any
	if val.IsValid() {
		v = val.Interface()
	}
	if c.columns.columnRenders[field] != nil {
		v = c.columns.columnRenders[field](rowData, v, row, col)
	}
	return cast.ToString(v)
}
