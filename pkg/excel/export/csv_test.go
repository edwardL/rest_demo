package export_test

import (
	"fmt"
	"rest_demo/pkg/excel/export"
	"testing"
)

func TestCsv_ExportWithForceZip(t *testing.T) {
	ep := export.NewCsv(testHeader, newTestTask(3000, 0),
		export.WithFilename(testGetFilename()), export.WithForceZip())
	fmt.Println(ep.Export())

	ep1 := export.NewCsv(testHeader, newTestTask(3000, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestCsv_Export(t *testing.T) {
	ep := export.NewCsv(testHeader, newTestTask(300, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep.Export())
}

func TestCsv_ExportWithSingleFileMaxRows(t *testing.T) {
	ep := export.NewCsv(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()), export.WithSingleFileMaxRows(140))
	fmt.Println(ep.Export())

	ep1 := export.NewCsv(testHeader, newTestTask(500, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestCsv_WithForceSingleFile(t *testing.T) {
	ep := export.NewCsv(testHeader, newTestTask(210000, 0),
		export.WithFilename(testGetFilename()), export.WithForceSingleFile())
	fmt.Println(ep.Export())

	ep1 := export.NewCsv(testHeader, newTestTask(210000, 0),
		export.WithFilename(testGetFilename()))
	fmt.Println(ep1.Export())
}

func TestCsvExports(t *testing.T) {
	//f1, _ := os.OpenFile("./csv.go", os.O_CREATE|os.O_RDWR, os.ModePerm)
	//defer f1.Close()
	//z1, _ := os.OpenFile("./z1.zip", os.O_CREATE|os.O_RDWR, os.ModePerm)
	//defer z1.Close()
	//w1 := zip.NewWriter(z1)
	//w1.
	//	f1.WriteTo(w1)
}
