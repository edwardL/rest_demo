package orm

import "database/sql"

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type count[T number] struct {
	v   *sql.Null[T]
	col string
}

func COUNT[T number](col string, v *sql.Null[T]) *count[T] {
	return &count[T]{v: v, col: "COUNT(" + col + ")"}
}

func (c *count[T]) Mapping() []*Mapping {
	return []*Mapping{
		{
			Column: c.col,
			Result: c.v,
		},
	}
}

type sum[T number] struct {
	v   *sql.Null[T]
	col string
}

func SUM[T number](col string, v *sql.Null[T]) *sum[T] {
	return &sum[T]{
		v: v, col: "SUM(" + col + ")",
	}
}

func (s *sum[T]) Mapping() []*Mapping {
	return []*Mapping{
		{
			Column: s.col,
			Result: s.v,
		},
	}
}

type avg[T number] struct {
	v   *sql.Null[T]
	col string
}

func AVG[T number](col string, v *sql.Null[T]) *avg[T] {
	return &avg[T]{
		v:   v,
		col: "AVG(" + col + ")",
	}
}
func (a *avg[T]) Mapping() []*Mapping {
	return []*Mapping{
		{
			Column: a.col,
			Result: a.v,
		},
	}
}

type min[T number] struct {
	v   *sql.Null[T]
	col string
}

func MIN[T number](col string, v *sql.Null[T]) *min[T] {
	return &min[T]{
		v:   v,
		col: "MIN(" + col + ")",
	}
}
func (m *min[T]) Mapping() []*Mapping {
	return []*Mapping{
		{
			Column: m.col,
			Result: m.v,
		},
	}
}

type max[T number] struct {
	v   *sql.Null[T]
	col string
}

func MAX[T number](col string, v *sql.Null[T]) *max[T] {
	return &max[T]{
		v:   v,
		col: "MAX(" + col + ")",
	}
}
func (m *max[T]) Mapping() []*Mapping {
	return []*Mapping{
		{
			Column: m.col,
			Result: m.v,
		},
	}
}
