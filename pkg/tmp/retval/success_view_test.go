package retval

import (
	hhtools "nwgit.gzhhit.com/BD/hhitcomm.git"
	"nwgit.gzhhit.com/BD/hhitframe.git/types"
	ftypes "nwgit.gzhhit.com/BD/hhitframe.git/types"
	"testing"
)

type viewRow struct {
	Id       int    `json:"id"`
	Port     int    `json:"port"`
	Severity string `json:"severity"`
	Memo     string `json:"memo"`
}

func buildCliInfo(fields map[string]types.DataViewField) *types.SitkWebCliInfo {
	var info = &types.SitkWebCliInfo{
		ReqViewInfo: &types.SitkWebViewInfo{
			SearchViewInfo: types.DataViewSearchInfo{FieldList: fields},
		},
	}
	return info
}

func TestSuccessPageView_FilterFields(t *testing.T) {
	var info = buildCliInfo(map[string]types.DataViewField{
		"id":   {FieldAlias: "id"},
		"port": {FieldAlias: "port"},
	})
	var rows = []viewRow{{Id: 1, Port: 2, Severity: "低危", Memo: "m"}}

	var tp uint8
	var resp any
	tp, resp = SuccessPageView(info, 65535, rows)
	if tp != 1 {
		t.Fatalf("SuccessPageView 返回类型错误，期望1，实际%d", tp)
	}

	var ret = resp.(ftypes.ResultData01)
	var data = ret.Data
	if len(data) != 1 {
		t.Fatalf("分页返回外层长度错误: %d", len(data))
	}
	if int(data[0]["total_nums"].(int)) != 65535 {
		t.Fatalf("total_nums 转换错误: %#v", data[0]["total_nums"])
	}

	var pageData = data[0]["page_data"].([]map[string]any)
	if len(pageData) != 1 {
		t.Fatalf("page_data 长度错误: %d", len(pageData))
	}
	if len(pageData[0]) != 2 {
		t.Fatalf("字段过滤失败: %#v", pageData[0])
	}
	if _, ok := pageData[0]["id"]; !ok {
		t.Fatalf("字段过滤后缺少 id")
	}
	if _, ok := pageData[0]["port"]; !ok {
		t.Fatalf("字段过滤后缺少 port")
	}
	if _, ok := pageData[0]["severity"]; ok {
		t.Fatalf("字段过滤后不应包含 severity")
	}
}

func TestSuccessView_FilterFields(t *testing.T) {
	var info = buildCliInfo(map[string]types.DataViewField{
		"id":   {FieldAlias: "id"},
		"port": {FieldAlias: "port"},
	})

	var tp uint8
	var resp any
	tp, resp = SuccessView(info, viewRow{Id: 7, Port: 9, Severity: "高危", Memo: "x"})
	if tp != 1 {
		t.Fatalf("SuccessView 返回类型错误，期望1，实际%d", tp)
	}

	var ret = resp.(ftypes.ResultData01)
	var data = ret.Data
	if len(data) != 1 {
		t.Fatalf("列表长度错误: %d", len(data))
	}
	if len(data[0]) != 2 {
		t.Fatalf("字段过滤失败: %#v", data[0])
	}
	if _, ok := data[0]["id"]; !ok {
		t.Fatalf("字段过滤后缺少 id")
	}
	if _, ok := data[0]["port"]; !ok {
		t.Fatalf("字段过滤后缺少 port")
	}
	if _, ok := data[0]["severity"]; ok {
		t.Fatalf("字段过滤后不应包含 severity")
	}
}

func TestSuccessViewAndPageView_Error(t *testing.T) {
	var info = buildCliInfo(map[string]types.DataViewField{"id": {FieldAlias: "id"}})

	var tp1 uint8
	var r1 any
	tp1, r1 = SuccessView(info, make(chan int))
	if tp1 != 1 {
		t.Fatalf("SuccessView 错误返回类型错误: %d", tp1)
	}
	if r1.(ftypes.ResultData01).Code != hhtools.CodeReDataFormatError {
		t.Fatalf("SuccessView 错误码异常")
	}

	var tp2 uint8
	var r2 any
	tp2, r2 = SuccessPageView(info, 1, []chan int{make(chan int)})
	if tp2 != 1 {
		t.Fatalf("SuccessPageView 错误返回类型错误: %d", tp2)
	}
	if r2.(ftypes.ResultData01).Code != hhtools.CodeReDataFormatError {
		t.Fatalf("SuccessPageView 错误码异常")
	}
}

func TestFilterMapListByViewBranches(t *testing.T) {
	var src = []map[string]any{{"id": 1, "port": 2, "severity": "low"}}

	var out = filterMapListByView(nil, src)
	if len(out[0]) != 3 {
		t.Fatalf("nil info 不应过滤: %#v", out)
	}

	var infoNoReq = &types.SitkWebCliInfo{}
	out = filterMapListByView(infoNoReq, src)
	if len(out[0]) != 3 {
		t.Fatalf("nil ReqViewInfo 不应过滤: %#v", out)
	}

	var infoEmpty = buildCliInfo(map[string]types.DataViewField{})
	out = filterMapListByView(infoEmpty, src)
	if len(out[0]) != 3 {
		t.Fatalf("空 FieldList 不应过滤: %#v", out)
	}

	var infoAllow = buildCliInfo(map[string]types.DataViewField{
		"id":   {FieldAlias: "id"},
		"port": {FieldAlias: "port"},
	})
	out = filterMapListByView(infoAllow, []map[string]any{{"id": 1, "port": 2, "severity": "low"}})
	if len(out[0]) != 2 {
		t.Fatalf("有 FieldList 时应过滤: %#v", out)
	}
}
