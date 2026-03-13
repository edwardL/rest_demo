package orm

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Update struct {
	dbCommon
	cols  []string
	where []string // 查询语法条件 例如：["AND id = ?", "OR account = ?"]
}

func UPDATE(tableName string) *Update {
	update := &Update{
		dbCommon: dbCommon{
			table: tableName,
		},
	}
	return update
}

func (d *Update) Debug() *Update {
	d.debug = true
	return d
}

// 不需要占位符的 SET 方法
func (d *Update) SET(set map[string]any) *Update {
	if len(set) == 0 {
		d.err = errors.New("empty set")
		return d
	}

	keys := slices.Sorted(maps.Keys(set))
	for _, k := range keys { // 按键排序
		d.cols = append(d.cols, k)
		d.args = append(d.args, set[k])
	}
	return d
}

func (d *Update) SET1(row Model) *Update {
	if row == nil {
		d.err = errors.New("empty model")
		return d
	}

	mapping := row.Mapping()
	for _, v := range mapping {
		if v.Column == "id" {
			continue
		}
		//if util.IsZero(v.Value) {
		//	continue
		//}
		d.cols = append(d.cols, v.Column)
		d.args = append(d.args, v.Value)
	}

	if len(d.cols) == 0 { // 没有更新字段
		return d
	}
	return d
}

func (d *Update) WHERE(where map[string]any) *Update {
	if len(where) > 0 {
		d.where = append(d.where, "1 = 1")
	}
	// 按键排序
	keys := slices.Sorted(maps.Keys(where))
	for _, k := range keys {
		d.where = append(d.where, k)
		d.args = append(d.args, where[k])
	}
	return d
}

func (d *Update) SQL() string {
	sqlText := "UPDATE " + d.table + " SET " + strings.Join(d.cols, " = ?, ") + " = ? WHERE " + strings.Join(d.where, " ")
	return sqlText
}

func (d *Update) Exec(ctx context.Context, db Execer) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}
	sqlText := d.SQL()
	d.debugPrint(ctx, sqlText)
	res, err := db.ExecContext(ctx, sqlText, d.args...)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return res.RowsAffected()
}

func (d *Update) DryRun(ctx context.Context) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}
	sqlText := d.SQL()
	d.print(ctx, sqlText)
	return 0, nil
}
