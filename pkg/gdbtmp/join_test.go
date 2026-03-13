package gdbtmp

import (
	"fmt"
	"testing"
)

// UPDATE T1, T2,
// [INNER JOIN | LEFT JOIN] T1 ON T1.C1 = T2. C1
// SET T1.C2 = T2.C2,
// T2.C3 = expr
// WHERE condition

// DELETE T1, T2
// FROM T1
// INNER JOIN T2 ON T1.key = T2.key
// WHERE condition;
func TestExecJoin(t *testing.T) {
	var res, _ = NewSql("hhit_asset_update_queue q").
		LeftJoin("hhit_asset a", "q.uid = a.uid").
		Where("q.uid != ''").
		Updates(map[string]any{
			"q.uid": Raw("a.uid"),
			"q.ts":  Raw("a.ts"),
		})
	fmt.Println(res.Sql, res.Args)
	res, _ = NewSql("user u").
		Where("u.aa = ?", 2).
		Updates(map[string]any{
			"u.aa": Raw("u2.aa"),
			"u.bb": Raw("u2.bb"),
		})
	fmt.Println(res.Sql, res.Args)

	res, _ = NewSql("user u").
		LeftJoin("user2 u2", "u.a = u2.a").
		Where("u.aa = ?", 2).
		Update("u.aa", Raw("u2.bb"))
	fmt.Println(res.Sql, res.Args)
	res, _ = NewSql("user u").
		Where("u.aa = ?", 2).
		Update("u.aa", Raw("u2.bb"))
	fmt.Println(res.Sql, res.Args)

	res, _ = NewSql("user u").
		LeftJoin("user2 u2", "u.a = u2.a").
		Where("u.aa = ?", 2).
		UpdateIgnore(map[string]any{
			"u.aa": Raw("u2.aa"),
			"u.bb": Raw("u2.bb"),
		})
	fmt.Println(res.Sql, res.Args)
	res, _ = NewSql("user u").
		Where("u.aa = ?", 2).
		UpdateIgnore(map[string]any{
			"u.aa": Raw("u2.aa"),
			"u.bb": Raw("u2.bb"),
		})
	fmt.Println(res.Sql, res.Args)

	res, _ = NewSql("hhit_asset_update_queue q").
		LeftJoin("hhit_asset a", "q.uid = a.uid").
		Where("q.uid = ?", "ssss").
		Delete("q")
	fmt.Println(res.Sql, res.Args)
	res, _ = NewSql("user u").
		Where("u.aa = ?", 2).
		Where("u.aa != u2.aa").
		Delete()
	fmt.Println(res.Sql, res.Args)
}
