package gdbtmp

import (
	"context"
)

// WithLog 配置日志
func WithLog(log GdbLog) Option {
	return func(c *Conf) {
		c.Log = log
	}
}

// WithLogLevel 配置日志级别
func WithLogLevel(lv LogLevel) Option {
	return func(c *Conf) {
		c.LogLevel = lv
	}
}

// WithWriteHhDbLog 打印hhDb日志
func WithWriteHhDbLog(wl bool) Option {
	return func(c *Conf) {
		c.WriteHhDbLog = wl
	}
}

// WithWriteLog 打印gdb日志
func WithWriteLog(wl bool) Option {
	return func(c *Conf) {
		c.WriteLog = wl
	}
}

// WithTraceIdFunc 配置traceId处理函数
func WithTraceIdFunc(f func(ctx context.Context) string) Option {
	return func(c *Conf) {
		if f == nil {
			return
		}
		c.TraceIdFunc = f
	}
}

// WithAppendDbKeywords 配置数据库关键词
func WithAppendDbKeywords(kw []string) Option {
	return func(c *Conf) {
		c.DbKeywords = kw
	}
}

// WithAppendZeroValIgnoreField 配置零值忽略字段
func WithAppendZeroValIgnoreField(ig []string) Option {
	return func(c *Conf) {
		c.ZeroValIgnoreField = ig
	}
}

// WithMapAssignment 配置map赋值处理函数 用于自定义类型
func WithMapAssignment(f MapAssignmentFunc) Option {
	return func(c *Conf) {
		c.MapAssignment = f
	}
}

// WithStructAssignment 配置struct赋值处理函数 用于自定义类型
func WithStructAssignment(f StructAssignmentFunc) Option {
	return func(c *Conf) {
		c.StructAssignment = f
	}
}

// WithWriteErrSql 是否打印错误sql 避免打印敏感信息
func WithWriteErrSql(we bool) Option {
	return func(c *Conf) {
		c.WriteErrSql = we
	}
}

// WithWriteCompSql 是否打印完整sql 避免打印敏感信息
func WithWriteCompSql(wc bool) Option {
	return func(c *Conf) {
		c.WriteCompSql = wc
	}
}
