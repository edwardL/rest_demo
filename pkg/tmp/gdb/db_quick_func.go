package gdb

// ----- Db

// Eq Equal
func (d *Db) Eq(column string, value any) *Db {
	d.Where(column+" = ?", value)
	return d
}

// NotEq Not Equal
func (d *Db) NotEq(column string, value any) *Db {
	d.Where(column+" != ?", value)
	return d
}

// Like LIKE "%%"
func (d *Db) Like(column string, value string) *Db {
	d.LikeRaw(column+" LIKE ?", "%"+value+"%")
	return d
}

// LikeRaw LIKE ""
func (d *Db) LikeRaw(column string, value string) *Db {
	d.Where(column+" LIKE ?", value)
	return d
}

// In IN()
func (d *Db) In(column string, value any) *Db {
	d.Where(column+" IN (?)", value)
	return d
}

// Lt 小于
func (d *Db) Lt(column string, value any) *Db {
	d.Where(column+" < ?", value)
	return d
}

// Le 小于等于
func (d *Db) Le(column string, value any) *Db {
	d.Where(column+" <= ?", value)
	return d
}

// Gt 大于
func (d *Db) Gt(column string, value any) *Db {
	d.Where(column+" > ?", value)
	return d
}

// Ge 大于等于
func (d *Db) Ge(column string, value any) *Db {
	d.Where(column+" >= ?", value)
	return d
}

// Between 之间
func (d *Db) Between(column string, value1 any, value2 any) *Db {
	d.Where(column+" BETWEEN ? AND ?", value1, value2)
	return d
}

// IsNotNull 不为空
func (d *Db) IsNotNull(column string) *Db {
	d.Where(column + " IS NOT NULL")
	return d
}

// ------Db[T]

// Eq Equal
func (d *DbT[T]) Eq(column string, value any) *DbT[T] {
	d.Db.Eq(column, value)
	return d
}

// NotEq Not Equal
func (d *DbT[T]) NotEq(column string, value any) *DbT[T] {
	d.Db.NotEq(column, value)
	return d
}

// Like LIKE "%%"
func (d *DbT[T]) Like(column string, value string) *DbT[T] {
	d.Db.Like(column, value)
	return d
}

// LikeRaw LIKE ""
func (d *DbT[T]) LikeRaw(column string, value string) *DbT[T] {
	d.Db.LikeRaw(column, value)
	return d
}

// In IN()
func (d *DbT[T]) In(column string, value any) *DbT[T] {
	d.Db.In(column, value)
	return d
}

// Lt 小于
func (d *DbT[T]) Lt(column string, value any) *DbT[T] {
	d.Db.Lt(column, value)
	return d
}

// Le 小于等于
func (d *DbT[T]) Le(column string, value any) *DbT[T] {
	d.Db.Le(column, value)
	return d
}

// Gt 大于
func (d *DbT[T]) Gt(column string, value any) *DbT[T] {
	d.Db.Gt(column, value)
	return d
}

// Ge 大于等于
func (d *DbT[T]) Ge(column string, value any) *DbT[T] {
	d.Db.Ge(column, value)
	return d
}

// Between 之间
func (d *DbT[T]) Between(column string, value1 any, value2 any) *DbT[T] {
	d.Db.Between(column, value1, value2)
	return d
}

// IsNotNull 不为空
func (d *DbT[T]) IsNotNull(column string) *DbT[T] {
	d.Db.IsNotNull(column)
	return d
}
