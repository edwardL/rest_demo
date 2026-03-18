package conv

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testFloat64Provider struct{}

func (p testFloat64Provider) Float64() float64 {
	return 12.5
}

type testFloat64EProvider struct{}

func (p testFloat64EProvider) Float64() (float64, error) {
	return 25.5, nil
}

type testFloat64EProviderErr struct{}

func (p testFloat64EProviderErr) Float64() (float64, error) {
	return 0, errors.New("float64 provider error")
}

func TestToString(t *testing.T) {
	var got = ToString(123)
	if got != "123" {
		t.Fatalf("ToString int failed, got=%s", got)
	}

	got = ToString(true)
	if got != "true" {
		t.Fatalf("ToString bool failed, got=%s", got)
	}

	var bytesVal = []byte("abc")
	got = ToString(bytesVal)
	if got != "abc" {
		t.Fatalf("ToString []byte failed, got=%s", got)
	}
}

func TestToIntSeries(t *testing.T) {
	var int64Val int64
	var err error
	int64Val, err = ToInt64("42")
	if err != nil {
		t.Fatalf("ToInt64 unexpected error: %v", err)
	}
	if int64Val != 42 {
		t.Fatalf("ToInt64 result mismatch, got=%d", int64Val)
	}

	var int32Val int32
	int32Val, err = ToInt32("12")
	if err != nil {
		t.Fatalf("ToInt32 unexpected error: %v", err)
	}
	if int32Val != 12 {
		t.Fatalf("ToInt32 result mismatch, got=%d", int32Val)
	}

	var intVal int
	intVal, err = ToInt("9")
	if err != nil {
		t.Fatalf("ToInt unexpected error: %v", err)
	}
	if intVal != 9 {
		t.Fatalf("ToInt result mismatch, got=%d", intVal)
	}

	_, err = ToInt64("abc")
	if err == nil {
		t.Fatalf("ToInt64 should fail for non-number")
	}
}

func TestToUint64(t *testing.T) {
	var got uint64
	var err error
	got, err = ToUint64("88")
	if err != nil {
		t.Fatalf("ToUint64 unexpected error: %v", err)
	}
	if got != 88 {
		t.Fatalf("ToUint64 result mismatch, got=%d", got)
	}

	_, err = ToUint64(-1)
	if err == nil {
		t.Fatalf("ToUint64 should fail for negative value")
	}
}

func TestToBool(t *testing.T) {
	var got bool
	var err error
	got, err = ToBool("yes")
	if err != nil {
		t.Fatalf("ToBool unexpected error: %v", err)
	}
	if !got {
		t.Fatalf("ToBool expected true for yes")
	}

	got, err = ToBool("0")
	if err != nil {
		t.Fatalf("ToBool unexpected error: %v", err)
	}
	if got {
		t.Fatalf("ToBool expected false for 0")
	}

	_, err = ToBool("maybe")
	if err == nil {
		t.Fatalf("ToBool should fail for invalid bool string")
	}
}

func TestToFloat64E(t *testing.T) {
	var got float64
	var err error
	got, err = ToFloat64E("3.14")
	if err != nil {
		t.Fatalf("ToFloat64E unexpected error: %v", err)
	}
	if got != 3.14 {
		t.Fatalf("ToFloat64E result mismatch, got=%v", got)
	}

	got, err = ToFloat64E(testFloat64Provider{})
	if err != nil {
		t.Fatalf("ToFloat64E provider unexpected error: %v", err)
	}
	if got != 12.5 {
		t.Fatalf("ToFloat64E provider result mismatch, got=%v", got)
	}

	got, err = ToFloat64E(testFloat64EProvider{})
	if err != nil {
		t.Fatalf("ToFloat64E providerE unexpected error: %v", err)
	}
	if got != 25.5 {
		t.Fatalf("ToFloat64E providerE result mismatch, got=%v", got)
	}

	_, err = ToFloat64E(testFloat64EProviderErr{})
	if err == nil {
		t.Fatalf("ToFloat64E should fail for providerE error")
	}
}

func TestToTime(t *testing.T) {
	var got time.Time
	var err error
	got, err = ToTime("2025-01-02 03:04:05")
	if err != nil {
		t.Fatalf("ToTime format unexpected error: %v", err)
	}
	if got.Year() != 2025 || got.Month() != 1 || got.Day() != 2 {
		t.Fatalf("ToTime format parse mismatch, got=%v", got)
	}

	got, err = ToTime("1700000000")
	if err != nil {
		t.Fatalf("ToTime unix unexpected error: %v", err)
	}
	if got.Unix() != 1700000000 {
		t.Fatalf("ToTime unix parse mismatch, got=%d", got.Unix())
	}

	var zero time.Time
	got, err = ToTime(nil)
	if err != nil {
		t.Fatalf("ToTime nil unexpected error: %v", err)
	}
	if !got.Equal(zero) {
		t.Fatalf("ToTime nil expected zero time")
	}

	_, err = ToTime("not-time")
	if err == nil {
		t.Fatalf("ToTime should fail for invalid text")
	}
}

func TestSliceStringHelpers(t *testing.T) {
	var vals = ParseValToStrSlice("[a, b, c]")
	var expected = []string{"a", "b", "c"}
	if !reflect.DeepEqual(vals, expected) {
		t.Fatalf("ParseValToStrSlice mismatch, got=%v", vals)
	}

	vals = SplitTrim(" 1 , 2, ,3 ", ",")
	expected = []string{"1", "2", "3"}
	if !reflect.DeepEqual(vals, expected) {
		t.Fatalf("SplitTrim mismatch, got=%v", vals)
	}
}

func TestToValType(t *testing.T) {
	var got any
	got = ToValType("1", TypeInt)
	if got.(int) != 1 {
		t.Fatalf("ToValType int mismatch, got=%v", got)
	}

	got = ToValType("true", TypeBool)
	if got.(bool) != true {
		t.Fatalf("ToValType bool mismatch, got=%v", got)
	}

	got = ToValType(12, TypeString)
	if got.(string) != "12" {
		t.Fatalf("ToValType string mismatch, got=%v", got)
	}
}

func TestSliceToSlice(t *testing.T) {
	var source = [][]string{{"id"}, {"1"}, {"2"}}
	var target []int
	var err = SliceToSlice(source, &target)
	if err != nil {
		t.Fatalf("SliceToSlice unexpected error: %v", err)
	}
	var expected = []int{1, 2}
	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("SliceToSlice int mismatch, got=%v", target)
	}

	var sourceStr = [][]string{{"name"}, {"a"}, {"b"}}
	var targetPtr []*string
	err = SliceToSlice(sourceStr, &targetPtr)
	if err != nil {
		t.Fatalf("SliceToSlice ptr unexpected error: %v", err)
	}
	if len(targetPtr) != 2 || *targetPtr[0] != "a" || *targetPtr[1] != "b" {
		t.Fatalf("SliceToSlice ptr mismatch, got len=%d", len(targetPtr))
	}

	err = SliceToSlice(source, target)
	if err == nil {
		t.Fatalf("SliceToSlice should fail when target is not pointer")
	}
}

func TestStrToTargetType(t *testing.T) {
	var v reflect.Value
	var err error
	v, err = StrToTargetType("7", reflect.TypeOf(int32(0)))
	if err != nil {
		t.Fatalf("StrToTargetType unexpected error: %v", err)
	}
	if v.Interface().(int32) != 7 {
		t.Fatalf("StrToTargetType int32 mismatch, got=%v", v.Interface())
	}

	v, err = StrToTargetType("abc", reflect.TypeOf(""))
	if err != nil {
		t.Fatalf("StrToTargetType string unexpected error: %v", err)
	}
	if v.Interface().(string) != "abc" {
		t.Fatalf("StrToTargetType string mismatch, got=%v", v.Interface())
	}

	_, err = StrToTargetType("1", reflect.TypeOf(true))
	if err == nil {
		t.Fatalf("StrToTargetType should fail for unsupported type")
	}
}

func TestToPtr(t *testing.T) {
	var p = ToPtr("abc")
	if p == nil || *p != "abc" {
		t.Fatalf("ToPtr mismatch")
	}
}
