package cache

import (
	"context"
	"fmt"
	"rest_demo/pkg/redis"
	"time"
)

const NoExpiration time.Duration = 0

type Cache interface {
	//Add 不存在才能添加成功
	Add(ctx context.Context, key string, value any, t time.Duration) bool
	Decrement(ctx context.Context, key string, value ...int) (int, error)
	Forever(ctx context.Context, key string, value any) bool
	Forget(ctx context.Context, key string) bool
	Flush(ctx context.Context) bool
	Get(ctx context.Context, key string, def ...any) any
	GetBool(ctx context.Context, key string, def ...bool) bool
	GetInt(ctx context.Context, key string, def ...int) int
	GetInt64(ctx context.Context, key string, def ...int64) int64
	GetString(ctx context.Context, key string, def ...string) string
	Has(ctx context.Context, key string) bool
	Increment(ctx context.Context, key string, value ...int) (int, error)
	//Put 是否存在都会添加
	Put(ctx context.Context, key string, value any, t time.Duration) error
	Pull(ctx context.Context, key string, def ...any) any
	Remember(ctx context.Context, key string, seconds time.Duration, callback func() (any, error)) (any, error)
	RememberForever(ctx context.Context, key string, callback func() (any, error)) (any, error)
}

type Config struct {
	Prefix string `help:"缓存键前缀" default:"admin:"`
	Driver string `help:"默认缓存驱动，可选[redis,memory]" default:"memory"`
	Redis  redis.Config
	Memory MemoryConfig
}

func NewCache(conf *Config) (Cache, error) {
	if conf.Driver == "redis" {
		c, err := redis.NewRedis(&conf.Redis)
		if err != nil {
			return nil, err
		}
		return NewRedis(conf.Prefix, c)
	} else if conf.Driver == "memory" {
		return NewMemory(conf.Prefix, &conf.Memory)
	}
	return nil, fmt.Errorf("cache driver[%s] not support", conf.Driver)
}
