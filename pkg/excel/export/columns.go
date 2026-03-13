package export

type columns struct {
	fields        []string
	titles        []any
	titleStr      []string
	nums          int
	keyIndex      map[string]int
	columnRenders map[string]CellRender
}

func newColumns(headers Headers) *columns {
	size := len(headers)
	if size == 0 {
		panic("header is empty")
	}
	c := &columns{
		fields:        make([]string, size),
		titles:        make([]any, size),
		titleStr:      make([]string, size),
		nums:          size,
		keyIndex:      make(map[string]int),
		columnRenders: make(map[string]CellRender),
	}
	for i := 0; i < size; i++ {
		c.fields[i] = headers[i].Field
		c.titles[i] = headers[i].Title
		c.titleStr[i] = headers[i].Title
		c.keyIndex[headers[i].Field] = i
		c.columnRenders[headers[i].Field] = headers[i].CellRender
	}
	return c
}
