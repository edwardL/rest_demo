package gdbtmp

import (
	"testing"
)

func TestDb_Init(t *testing.T) {
	Init(WithTraceIdFunc(traceFunc),
		WithLogLevel(DebugLogLevel),
		WithWriteHhDbLog(false),
		WithWriteLog(true),
		WithAppendDbKeywords([]string{"a", "b"}))

	Dump(conf)
	Dump(mysqlKeywords)
}
