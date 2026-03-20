package dialect

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetDialect_Registered(t *testing.T) {
	cases := []string{"sqlite3", "mysql", "postgres", "postgresql"}
	for _, name := range cases {
		d, ok := GetDialect(name)
		if !ok || d == nil {
			t.Fatalf("expected dialect %q to be registered", name)
		}
	}
}

func TestSQLite3_TableExistSql(t *testing.T) {
	d, ok := GetDialect("sqlite3")
	if !ok {
		t.Fatal("sqlite3 dialect not registered")
	}

	sql, args := d.TableExistSql("users")
	if !strings.Contains(sql, "sqlite_master") {
		t.Fatalf("unexpected sqlite table exist sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "users" {
		t.Fatalf("unexpected sqlite table exist args: %#v", args)
	}
}

func TestMySQL_DataTypeOf(t *testing.T) {
	d, ok := GetDialect("mysql")
	if !ok {
		t.Fatal("mysql dialect not registered")
	}

	if got := d.DataTypeOf(reflect.ValueOf("abc")); got != "varchar(255)" {
		t.Fatalf("mysql string type mismatch, got: %s", got)
	}
	if got := d.DataTypeOf(reflect.ValueOf(time.Now())); got != "datetime" {
		t.Fatalf("mysql time type mismatch, got: %s", got)
	}
}

func TestPostgres_DataTypeOfAndTableExistSql(t *testing.T) {
	d, ok := GetDialect("postgres")
	if !ok {
		t.Fatal("postgres dialect not registered")
	}

	if got := d.DataTypeOf(reflect.ValueOf(float64(1))); got != "double precision" {
		t.Fatalf("postgres float64 type mismatch, got: %s", got)
	}
	if got := d.DataTypeOf(reflect.ValueOf(time.Now())); got != "timestamp with time zone" {
		t.Fatalf("postgres time type mismatch, got: %s", got)
	}

	sql, args := d.TableExistSql("users")
	if !strings.Contains(sql, "$1") || !strings.Contains(sql, "pg_tables") {
		t.Fatalf("unexpected postgres table exist sql: %s", sql)
	}
	if len(args) != 1 || args[0] != "users" {
		t.Fatalf("unexpected postgres table exist args: %#v", args)
	}
}
