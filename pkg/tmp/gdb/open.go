package gdb

// GenWhere 开放条件构建函数
func (d *Db) GenWhere() (whereStr string, args []any) {
	return d.gs.GenWhere()
}

// QueryWrapper 开放条件构建函数
func (d *Db) QueryWrapper() *Wrapper {
	var qw = &Wrapper{Db: d}
	return qw
}

// GenWhere 条件构建函数
func (d *Sql) GenWhere() (whereStr string, args []any) {
	return d.genWhere()
}

// QueryWrapper 条件构建函数
func (d *Sql) QueryWrapper() *Wrapper {
	var qw = QueryWrapper()
	qw.gs = d
	return qw
}
