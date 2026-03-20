package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type mysql struct{}

var _ Dialect = (*mysql)(nil)

func init() {
	RegisterDialect("mysql", &mysql{})
}

func (m *mysql) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int8, reflect.Uint8:
		return "tinyint"
	case reflect.Int16, reflect.Uint16:
		return "smallint"
	case reflect.Int32, reflect.Uint32:
		return "int"
	case reflect.Int, reflect.Uint, reflect.Int64, reflect.Uint64, reflect.Uintptr:
		return "bigint"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.String:
		return "varchar(255)"
	case reflect.Slice, reflect.Array:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (m *mysql) TableExistSql(tableName string) (string, []interface{}) {
	return "SELECT TABLE_NAME FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", []interface{}{tableName}
}
