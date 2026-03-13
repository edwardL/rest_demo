package export_test

import (
	"fmt"
	"reflect"
	"rest_demo/pkg/excel/export"
	"testing"
)

func TestExcel_ExportWithForceZip(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(3000, 0),
		export.WithFilename(testGetFilename()), export.WithForceZip())
	fmt.Println(ep.Export())

	ep1 := export.NewExcel(testHeader, newTestTask(3000, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestExcel_Export(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(300, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep.Export())
}

func TestExcel_ExportWithSingleFileMaxRows(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()), export.WithSingleFileMaxRows(140))
	fmt.Println(ep.Export())

	ep1 := export.NewExcel(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestExcel_WithRowStart(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()), export.WithRowStart(3), export.WithSingleFileMaxRows(140))
	fmt.Println(ep.Export())

	ep1 := export.NewExcel(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()), export.WithRowStart(3), export.WithColStart(2))
	fmt.Println(ep1.Export())
}

func TestExcel_WithForceSingleFile(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(210000, 0),
		export.WithFilename(testGetFilename()), export.WithForceSingleFile())
	fmt.Println(ep.Export())

	ep1 := export.NewExcel(testHeader, newTestTask(210000, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestExcelProcessRow(t *testing.T) {
	ep := export.NewExcel(testHeader, newTestTask(100, 0))

	a := map[string]any{
		"a": 1212,
		"b": "ffff",
		"C": 23.434,
		"d": 66666,
	}
	fmt.Println("a value = ", ep.TestProcessRow(reflect.ValueOf(a), 1))

	fmt.Println("a addr = ", ep.TestProcessRow(reflect.ValueOf(&a), 1))

	b := struct {
		A string
		B int `excel:"b"`
		C float64
		D string
	}{
		A: "aaa",
		B: 222,
		C: 232.54,
		D: "gggg",
	}

	fmt.Println("b value = ", ep.TestProcessRow(reflect.ValueOf(b), 1))

	fmt.Println("b addr = ", ep.TestProcessRow(reflect.ValueOf(&b), 1))

	c := []any{
		2323.434,
		"bbbb",
		2323,
		99999,
	}

	fmt.Println("c value = ", ep.TestProcessRow(reflect.ValueOf(c), 1))

	fmt.Println("c addr = ", ep.TestProcessRow(reflect.ValueOf(&c), 1))

	fmt.Println("d  = ", ep.TestProcessRow(reflect.ValueOf(""), 1))
	fmt.Println("d  = ", ep.TestProcessRow(reflect.ValueOf(2323), 1))

}
