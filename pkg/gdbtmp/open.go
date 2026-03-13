package gdbtmp

// GenWhere 开放条件构建函数
func (d *Db) GenWhere() (whereStr string, args []any) {
	return d.gs.genWhere()
}

// GenWhere 开放条件构建函数
func (d *Sql) GenWhere() (whereStr string, args []any) {
	return d.genWhere()
}
