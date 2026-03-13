package gdbtmp

import (
	"fmt"
	"os"
	"testing"
)

func TestSql_SaveInBatches(t *testing.T) {
	args := []string{"$$", "&&"}
	one := "one"
	r, err := NewSql("table").
		Where("aa = ?", one).
		WhereOr("dd IN (?)", args).
		Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", []int{1}, append(args, "dds"), []float64{1, 2, 3}).
		WhereOr("aa = ?", one).
		WhereOr("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args).
		WhereGroup(func(tx *Sql) (err error) {
			tx.Where("aa = ?", one).
				Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args)
			return nil
		}).WhereGroupOr(func(tx *Sql) (err error) {
		tx.Where("aa = ?", one).
			Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args)
		return nil
	}).
		Select()
	fmt.Println(err)
	fmt.Println(r.Sql, r.Args)
	os.Exit(0)
	data := []map[string]any{
		{
			"id":   1,
			"name": "test",
			"age":  18,
		},
		{
			"id":   2,
			"name": "test2",
			"age":  19,
		},
	}

	create, err := NewSql("tb_name").OmitFields("age").Create(data[0])
	if err != nil {
		fmt.Println("Create", err)
	} else {
		fmt.Println("Create", create.Sql, create.Args)
	}

	update, err := NewSql("tb_name").Where("id >?", 0).Updates(data[0])
	if err != nil {
		fmt.Println("update", err)
	} else {
		fmt.Println("update", update.Sql, update.Args)
	}

	createBat, err := NewSql("tb_name").OmitFields("age").CreateInBatches(data)
	if err != nil {
		fmt.Println("CreateInBatches", err)
	} else {
		fmt.Println("CreateInBatches", createBat.Sql, createBat.Args)
	}

	save, err := NewSql("tb_name").Save(data[0])
	if err != nil {
		fmt.Println("Save", err)
	} else {
		fmt.Println("Save", save.Sql, save.Args)
	}

	saveBat, err := NewSql("tb_name").SaveInBatches(data)
	if err != nil {
		fmt.Println("SaveInBatches", err)
	} else {
		fmt.Println("SaveInBatches", saveBat.Sql, saveBat.Args)
	}
}
func TestSql_Select(t *testing.T) {
	args := []string{"$$", "&&"}
	one := "one"
	s, err := NewSql("table").
		Where("dd IN (?) AND name = ? AND (ts,id) IN (?) AND id IN (?) AND name = ? AND id IN (?) ", []int{1, 2, 3}, "sss", [][]int{{4, 6}, {5, 6}}, []int{1, 2, 3}, "ddd", []int{1, 2, 3}).
		Where("aa = ?", one).
		WhereOr("dd IN (?)", args).
		Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", []int{1}, append(args, "dds"), []float64{1, 2, 3}).
		WhereOr("aa = ?", one).
		WhereOr("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args).
		WhereGroup(func(tx *Sql) (err error) {
			tx.Where("aa = ?", one).
				Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args)
			return nil
		}).WhereGroupOr(func(tx *Sql) (err error) {
		tx.Where("aa = ?", one).
			Where("(aa IN (?) AND bb IN (?) AND cc IN (?))", args, args, args)
		return nil
	}).Select()
	fmt.Println(err)
	fmt.Println(s.Sql)
	fmt.Println(s.CompSql())
}

func TestSql_WhereCombination(t *testing.T) {
	var iaa = [][]int{{1, 2}, {3, 4}}
	var ia = []int{5, 6, 7}
	var str = "aaa"
	var ian []any = []any{[]any{8, "b"}, []any{"d", "b"}}
	var iann [][]any = [][]any{{9, "g"}, {"h", "j"}}
	s, err := NewSql("table").
		Where("dd IN (?) AND (ts,id) IN (?) AND (ts,id) IN (?) AND name = ? AND (ts,id) IN (?) AND id IN (?) AND name = ? AND id IN (?) ",
			ia, ian, iann, str, iaa, ia, str, ia).
		Select()
	fmt.Println(err)
	fmt.Println(s.Sql)
	fmt.Println(s.CompSql())
}

type Aaa struct {
	Aaa string
}

func (Aaa) TableName() string {
	return "tb_name"
}

type Bbb struct {
	Bbb string
}

func (*Bbb) TableName() string {
	return "tb_bb_name"
}

func TestSql_getSliceModelTableName(t *testing.T) {
	fmt.Println(NewSql("Tb").Select())
	fmt.Println(NewSql().Table("(?)", NewSql("tb")).Select())
	var a1 []Aaa = make([]Aaa, 0)
	fmt.Println(NewSql(a1).Select())
	var a2 []Aaa = []Aaa{{}}
	fmt.Println(NewSql(a2).Select())
	var a3 []Aaa
	fmt.Println(NewSql(a3).Select())
	var a4 []*Aaa = make([]*Aaa, 0)
	fmt.Println(NewSql(a4).Select())
	var a5 []*Aaa = []*Aaa{&Aaa{}}
	fmt.Println(NewSql(a5).Select())
	var a6 []*Aaa
	fmt.Println(NewSql(a6).Select())
	fmt.Println(NewSql(Aaa{}).Select())
	fmt.Println(NewSql(&Aaa{}).Select())
	var a7 Aaa
	var a8 *Aaa
	fmt.Println(NewSql(a7).Select())
	fmt.Println(NewSql(a8).Select())

	var b1 []Bbb = make([]Bbb, 0)
	fmt.Println(NewSql(b1).Select())
	var b2 []Bbb = []Bbb{{}}
	fmt.Println(NewSql(b2).Select())
	var b3 []Bbb
	fmt.Println(NewSql(b3).Select())
	var b4 []*Bbb = make([]*Bbb, 0)
	fmt.Println(NewSql(b4).Select())
	var b5 []*Bbb = []*Bbb{&Bbb{}}
	fmt.Println(NewSql(b5).Select())
	var b6 []*Bbb
	fmt.Println(NewSql(b6).Select())
	fmt.Println(NewSql(Bbb{}).Select())
	fmt.Println(NewSql(&Bbb{}).Select())
	var b7 Bbb
	var b8 *Bbb
	fmt.Println(NewSql(b7).Select())
	fmt.Println(NewSql(b8).Select())
}
