package gdb

import (
	"context"
	"time"
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

// WithAppendDbFieldChar 配置数据库字段允许的字符
func WithAppendDbFieldChar(fc []byte) Option {
	return func(c *Conf) {
		c.DbFieldChar = append(c.DbFieldChar, fc...)
	}
}

// WithAppendZeroValIgnoreField 配置零值忽略字段
func WithAppendZeroValIgnoreField(ig []string) Option {
	return func(c *Conf) {
		c.ZeroValIgnoreField = ig
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

// WithEmptyError 查询为空返回错误
func WithEmptyError(err DbError) Option {
	return func(c *Conf) {
		err.isNotFound = true
		c.EmptyError = &err
	}
}

// WithDbConvInitPtr 配置conv_struct是否初始化指针
func WithDbConvInitPtr(csip bool) Option {
	return func(c *Conf) {
		c.DbConvInitPtr = csip
	}
}

// WithTimeLocation 配置时间时区
func WithTimeLocation(tl *time.Location) Option {
	return func(c *Conf) {
		c.TimeLocation = tl
	}
}

// WithDriveType 配置驱动类型
func WithDriveType(dt string) Option {
	return func(c *Conf) {
		c.DriveType = dt
	}
}

// WithDriveMap 配置驱动类型
func WithDriveMap(dt string, newFun NewWrapperFunc) Option {
	return func(c *Conf) {
		c.DriveMap[dt] = newFun
	}
}
