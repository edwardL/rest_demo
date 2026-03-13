package orm

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Insert[T Model] struct {
	dbCommon

	cols     []string // 查询字段
	size     int      // 插入数据条数
	conflict string
	updates  []string
}

func INSERT[T Model](rows ...T) *Insert[T] {
	insert := &Insert[T]{
		dbCommon: dbCommon{},
	}
	insert.insert(rows)
	return insert
}

func INSERT1() *Insert[*emptyModel] {
	return INSERT[*emptyModel]()
}

func (d *Insert[T]) Debug() *Insert[T] {
	d.debug = true
	return d
}

// insert into ab (a, b) values (?, ?)

func (d *Insert[T]) insert(rows []T) {
	switch len(rows) {
	case 1:
		d.size = 1
		mapping := rows[0].Mapping()
		for _, v := range mapping {
			if v.Column == "id" {
				continue
			}
			d.cols = append(d.cols, v.Column)
			d.args = append(d.args, v.Value)
		}
	default:
		d.size = len(rows)
		for i, row := range rows {
			mapping := row.Mapping()
			for _, v := range mapping {
				if v.Column == "id" {
					continue
				}
				if i == 0 {
					d.cols = append(d.cols, v.Column)
				}
				d.args = append(d.args, v.Value)
			}
		}
	}
}

func (d *Insert[T]) INTO(table string) *Insert[T] {
	d.table = table
	return d
}

func (d *Insert[T]) COLUMNS(cols ...string) *Insert[T] {
	if d.size != 0 {
		d.err = fmt.Errorf("columns and models can only be used once")
		return d
	}
	if len(cols) == 0 {
		d.err = fmt.Errorf("columns is empty")
		return d
	}
	d.cols = cols
	return d
}

func (d *Insert[T]) VALUES(args ...any) *Insert[T] {
	if d.size != 0 {
		d.err = fmt.Errorf("columns and models can only be used once")
		return d
	}
	if len(args) == 0 {
		d.err = fmt.Errorf("args is empty")
		return d
	}
	d.args = append(d.args, args...)
	d.size = 1
	return d
}

func (d *Insert[T]) ON(conflict string) *Insert[T] {
	d.conflict = conflict
	return d
}

func (d *Insert[T]) UPDATE(conds map[string]any) *Insert[T] {
	keys := slices.Sorted(maps.Keys(conds))
	for _, k := range keys { // 按键排序
		d.updates = append(d.updates, k)
		if v := conds[k]; v != nil {
			d.args = append(d.args, v)
		}
	}
	return d
}

func (d *Insert[T]) Exec(ctx context.Context, db Execer) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}

	sqlText := d.SQL()
	d.debugPrint(ctx, sqlText)
	res, err := db.ExecContext(ctx, sqlText, d.args...)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return res.LastInsertId()
}

func (d *Insert[T]) SQL() string {
	var builder strings.Builder
	builder.WriteString("INSERT INTO " + d.table + " (" + strings.Join(d.cols, ", ") + ") VALUES ")
	for i := range d.size {
		builder.WriteString("(" + strings.Repeat("?, ", len(d.cols)-1) + "?)")
		if i < d.size-1 {
			builder.WriteString(", ")
		}
	}
	if d.conflict != "" {
		builder.WriteString(" ON " + d.conflict + " UPDATE ")
		builder.WriteString(strings.Join(d.updates, ", "))
	}
	return builder.String()
}

func (d *Insert[T]) DryRun(ctx context.Context) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}
	sqlText := d.SQL()
	d.print(ctx, sqlText)
	return 0, nil
}
