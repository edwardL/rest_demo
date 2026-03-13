package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string
	TableExistSql(tableName string) (string, []interface{})
}

// RegisterDialect register a dialect to the global variable
func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

// Get the dialect from global variable if it exists
func GetDialect(name string) (Dialect, bool) {
	d, ok := dialectsMap[name]
	return d, ok
}
