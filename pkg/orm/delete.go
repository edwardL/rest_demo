package orm

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Delete struct {
	dbCommon
	where []string // 查询条件
}

func DELETE() *Delete {
	del := &Delete{
		dbCommon: dbCommon{},
	}
	return del
}

func (d *Delete) Debug() *Delete {
	d.debug = true
	return d
}

func (d *Delete) FROM(table string) *Delete {
	// 检查table 是否为空
	if table == "" {
		d.err = fmt.Errorf("table name is empty")
	}
	d.table = table
	return d
}

func (d *Delete) WHERE(where map[string]any) *Delete {
	if len(where) > 0 {
		d.where = append(d.where, "1 = 1")
	}
	keys := slices.Sorted(maps.Keys(where))
	for _, k := range keys {
		d.where = append(d.where, k)
		d.args = append(d.args, where[k])
	}
	return d
}

func (d *Delete) SQL() string {
	sqlText := "DELETE FROM " + d.table + " WHERE " + strings.Join(d.where, ", ")
	return sqlText
}

func (d *Delete) Exec(ctx context.Context, db Execer) (int64, error) {
	// 检查是否有错误
	if d.err != nil {
		return 0, d.err
	}
	sqlText := d.SQL()
	d.debugPrint(ctx, sqlText)
	res, err := db.ExecContext(ctx, sqlText, d.args...)
	if err != nil {
		return 0, fmt.Errorf("db..Exec: %w", err)
	}
	return res.RowsAffected()
}

func (d *Delete) DryRun(ctx context.Context, db Execer) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}
	sqlText := d.SQL()
	d.print(ctx, sqlText)
	return 0, nil
}
