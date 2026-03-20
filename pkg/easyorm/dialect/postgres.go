package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type postgres struct{}

var _ Dialect = (*postgres)(nil)

func init() {
	RegisterDialect("postgres", &postgres{})
	RegisterDialect("postgresql", &postgres{})
}

func (p *postgres) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int8, reflect.Uint8:
		return "smallint"
	case reflect.Int16, reflect.Uint16:
		return "smallint"
	case reflect.Int32, reflect.Uint32:
		return "integer"
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Uintptr:
		return "bigint"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "double precision"
	case reflect.String:
		return "text"
	case reflect.Slice, reflect.Array:
		return "bytea"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "timestamp with time zone"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (p *postgres) TableExistSql(tableName string) (string, []interface{}) {
	return "SELECT tablename FROM pg_tables WHERE schemaname = current_schema() AND tablename = $1", []interface{}{tableName}
}
