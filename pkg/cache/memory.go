package cache

import (
	"context"
	"time"

	cache2 "github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
)

var _ Cache = &Memory{}

type MemoryConfig struct {
	DefaultExpiration time.Duration `help:"默认过期时间" default:"5m"`
	CleanupInterval   time.Duration `help:"自动清理时间" default:"10m"`
}

type Memory struct {
	prefix   string
	instance *cache2.Cache
}

func NewMemory(prefix string, conf *MemoryConfig) (*Memory, error) {
	memory := cache2.New(conf.DefaultExpiration, conf.CleanupInterval)
	return &Memory{
		prefix:   prefix,
		instance: memory,
	}, nil
}

// Add Driver an item in the cache if the key does not exist.
func (r *Memory) Add(ctx context.Context, key string, value any, t time.Duration) bool {
	if t == NoExpiration {
		t = cache2.NoExpiration
	}
	err := r.instance.Add(r.key(key), value, t)
	return err == nil
}

func (r *Memory) Decrement(ctx context.Context, key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	r.Add(ctx, key, 0, cache2.NoExpiration)
	return r.instance.DecrementInt(r.key(key), value[0])
}

// Forever Driver an item in the cache indefinitely.
func (r *Memory) Forever(ctx context.Context, key string, value any) bool {
	if err := r.Put(ctx, key, value, cache2.NoExpiration); err != nil {
		return false
	}
	return true
}

// Forget Remove an item from the cache.
func (r *Memory) Forget(ctx context.Context, key string) bool {
	r.instance.Delete(r.key(key))
	return true
}

// Flush Remove all items from the cache.
func (r *Memory) Flush(ctx context.Context) bool {
	r.instance.Flush()

	return true
}

// Get Retrieve an item from the cache by key.
func (r *Memory) Get(ctx context.Context, key string, def ...any) any {
	val, exist := r.instance.Get(r.key(key))
	if exist {
		return val
	}
	if len(def) == 0 {
		return nil
	}

	switch s := def[0].(type) {
	case func() any:
		return s()
	default:
		return s
	}
}

func (r *Memory) GetBool(ctx context.Context, key string, def ...bool) bool {
	if len(def) == 0 {
		def = append(def, false)
	}
	res := r.Get(ctx, key, def[0])

	return cast.ToBool(res)
}

func (r *Memory) GetInt(ctx context.Context, key string, def ...int) int {
	if len(def) == 0 {
		def = append(def, 0)
	}

	return cast.ToInt(r.Get(ctx, key, def[0]))
}

func (r *Memory) GetInt64(ctx context.Context, key string, def ...int64) int64 {
	if len(def) == 0 {
		def = append(def, 0)
	}

	return cast.ToInt64(r.Get(ctx, key, def[0]))
}

func (r *Memory) GetString(ctx context.Context, key string, def ...string) string {
	if len(def) == 0 {
		def = append(def, "")
	}

	return cast.ToString(r.Get(ctx, key, def[0]))
}

// Has Check an item exists in the cache.
func (r *Memory) Has(ctx context.Context, key string) bool {
	_, exist := r.instance.Get(r.key(key))

	return exist
}

func (r *Memory) Increment(ctx context.Context, key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	r.Add(ctx, key, 0, cache2.NoExpiration)

	return r.instance.IncrementInt(r.key(key), value[0])
}

// Pull Retrieve an item from the cache and delete it.
func (r *Memory) Pull(ctx context.Context, key string, def ...any) any {
	var res any
	if len(def) == 0 {
		res = r.Get(ctx, key)
	} else {
		res = r.Get(ctx, key, def[0])
	}
	r.Forget(ctx, key)

	return res
}

// Put Driver an item in the cache for a given number of seconds.
func (r *Memory) Put(ctx context.Context, key string, value any, t time.Duration) error {
	r.instance.Set(r.key(key), value, t)

	return nil
}

// Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Memory) Remember(ctx context.Context, key string, seconds time.Duration, callback func() (any, error)) (any, error) {
	val := r.Get(ctx, key, nil)
	if val != nil {
		return val, nil
	}

	var err error
	val, err = callback()
	if err != nil {
		return nil, err
	}

	if err := r.Put(ctx, key, val, seconds); err != nil {
		return nil, err
	}

	return val, nil
}

// RememberForever Get an item from the cache, or execute the given Closure and store the result forever.
func (r *Memory) RememberForever(ctx context.Context, key string, callback func() (any, error)) (any, error) {
	val := r.Get(ctx, key, nil)
	if val != nil {
		return val, nil
	}

	var err error
	val, err = callback()
	if err != nil {
		return nil, err
	}
	if err := r.Put(ctx, key, val, cache2.NoExpiration); err != nil {
		return nil, err
	}
	return val, nil
}

func (r *Memory) key(key string) string {
	return r.prefix + key
}
