package orm

import (
	"context"
	"fmt"
)

type SelectModels[T any, P PModel[T]] struct {
	*selec[SelectModels[T, P]]
	row  *T
	rows *[]*T
}

func SELECT2[T any, P PModel[T]](rows *[]*T) *SelectModels[T, P] {
	s := &SelectModels[T, P]{
		selec: &selec[SelectModels[T, P]]{
			dbCommon: dbCommon{},
		},
		row:  new(T),
		rows: rows,
	}
	mappings := P(s.row).Mapping()
	for _, v := range mappings {
		s.cols = append(s.cols, v.Column)
		s.res = append(s.res, v.Result)
	}
	s.setT(s)
	return s
}

func (d *SelectModels[T, P]) Query(ctx context.Context, db Execer) error {
	if d.err != nil {
		return d.err
	}

	sqlText := d.SQL()
	d.debugPrint(ctx, sqlText)
	rows, err := db.QueryContext(ctx, sqlText, d.args...)
	if err != nil {
		return fmt.Errorf("stmt.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(d.res...); err != nil {
			return fmt.Errorf("rows.Scan: %w", err)
		}
		cp := *d.row
		*d.rows = append(*d.rows, &cp)
	}
	return rows.Err()
}
