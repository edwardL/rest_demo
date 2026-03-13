package export_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"rest_demo/pkg/excel/export"
	"time"

	rand2 "github.com/opdss/go-helper/rand"
)

var testHeader export.Headers = []export.Header{
	{
		Field: "a",
		Title: "第一列",
	},
	{
		Field: "B",
		Title: "第2列",
	},
	{
		Field: "c",
		Title: "第333列",
		CellRender: func(rowData reflect.Value, v any, row int, col int) any {
			bb, _ := json.Marshal(rowData.Interface())
			return v.(time.Time).Format(time.TimeOnly) + string(bb)
		},
	},
	{
		Field: "d",
		Title: "四列",
		CellRender: func(rowData reflect.Value, v any, row int, col int) any {
			return fmt.Sprintf("%v:%d-%d", v, row, col)
		},
	},
}

func testGetFilename() string {
	return fmt.Sprintf("/Users/wuxin/worker/hobi/common/excel/export/test_%d", time.Now().UnixMilli())
}

func testRandSliceData(curr int, typ int) []any {
	n := 1000
	res := make([]any, n)
	for i := 0; i < n; i++ {
		a := fmt.Sprintf("%d-%d", curr, i)
		b := float64(rand2.Int(100000, 999999)) * 0.5
		c := time.Now()
		d := rand2.StringN(5)
		switch typ {
		case 1:
			res[i] = map[string]any{"a": a, "B": b, "c": c, "d": d}
		case 2:
			res[i] = testStruct{a, b, c, d}
		case 3:
			res[i] = []any{a, b, c, d}
		default:
			switch curr % 3 {
			case 0:
				res[i] = map[string]any{"a": a, "B": b, "c": c, "d": d}
			case 1:
				res[i] = testStruct{a, b, c, d}
			case 2:
				res[i] = []any{a, b, c, d}
			}
		}
	}
	return res
}

type testStruct struct {
	A  string
	B  float64
	C1 time.Time `export:"c"`
	D  string
}

type testTask struct {
	hasMore bool
	curr    int
	total   int
	typ     int
	sliceDp *export.SliceDataProvider
}

func (dp *testTask) Next() bool {
	if !dp.hasMore {
		return false
	}
	if dp.curr >= dp.total {
		return false
	}
	hasNext := dp.sliceDp.Next()
	if hasNext {
		return hasNext
	}
	dp.sliceDp = export.NewSliceDataProvider(testRandSliceData(dp.curr, dp.typ))
	return dp.sliceDp.Next()
}

func (dp *testTask) Value() any {
	defer func() {
		dp.curr++
	}()
	return dp.sliceDp.Value()
}

func newTestTask(total int, typ int) *testTask {
	return &testTask{
		total:   total,
		curr:    0,
		hasMore: true,
		typ:     typ,
		sliceDp: export.NewSliceDataProvider(make([]any, 0)),
	}
}
