package slices

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/opdss/go-helper/funcs"
)

func TestIntersect(t *testing.T) {
	a1 := []int{1, 2, 3, 5, 4, 5}
	a2 := []int{4, 5, 5, 6, 7, 8}
	v := Intersect(a1, a2)
	want := []int{4, 5}
	if !Equal(v, want) {
		t.Errorf("Intersect(%v, %v) = %v, want %v", a1, a2, v, want)
	}
}

func TestUnion(t *testing.T) {
	a1 := []int{1, 2, 3, 4, 4, 5}
	a2 := []int{4, 5, 5, 6, 7, 8}
	v := Union(a1, a2)
	want := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if !Equal(v, want) {
		t.Errorf("Union(%v, %v) = %v, want %v", a1, a2, v, want)
	}
}

func TestDifference(t *testing.T) {
	a1 := []int{1, 2, 3, 5, 4, 5}
	a2 := []int{4, 5, 5, 6, 7, 8}
	v := Difference(a1, a2)
	want := []int{1, 2, 3, 6, 7, 8}
	if !Equal(v, want) {
		t.Errorf("Difference(%v, %v) = %v, want %v", a1, a2, v, want)
	}
}

func TestMap(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	go func() {
		funcs.RunTicker(context.Background(), time.Second, func() error {
			return nil
		})
	}()
	//fmt.Println(22222)
	s := Map(arr, func(v int, k int) string {
		time.Sleep(time.Second)
		fmt.Println(v, k)
		return fmt.Sprintf("%d", v)
	})
	if strings.Join(s, ",") != "1,2,3,4,5" {
		t.Errorf("Map(%v, fn) = %v, want %v", arr, s, strings.Join(s, ","))
	}

	s = MapAsync(arr, func(v int, k int) string {
		time.Sleep(time.Second)
		fmt.Println(v, k)
		return fmt.Sprintf("%d", v)
	})
	if strings.Join(s, ",") != "1,2,3,4,5" {
		t.Errorf("Map(%v, fn) = %v, want %v", arr, s, strings.Join(s, ","))
	}

	s = MapAsyncN(arr, func(v int, k int) string {
		time.Sleep(time.Second)
		fmt.Println(v, k)
		return fmt.Sprintf("%d", v)
	}, 3)
	if strings.Join(s, ",") != "1,2,3,4,5" {
		t.Errorf("Map(%v, fn) = %v, want %v", arr, s, strings.Join(s, ","))
	}
}

func Test(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9}
	r1 := Reduce(arr, func(res *int, item int, k int) {
		*res = *res + item
	})
	fmt.Println(r1)
	r2 := Reduce(arr, func(res *map[int]int, item int, k int) {
		if *res == nil {
			*res = make(map[int]int)
		}
		(*res)[item] = item
	})
	fmt.Println(r2)
	r3 := Reduce(arr, func(res *[]int, item int, k int) {
		*res = append(*res, item*item)
	})
	fmt.Println(r3)
	type Tage struct {
		age int
	}
	r4 := Reduce(arr, func(res *Tage, item int, k int) {
		res.age += item
	})
	fmt.Println(r4)
}
