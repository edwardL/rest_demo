package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/spf13/cast"
)

var _ Cache = &Redis{}

type Redis struct {
	prefix   string
	instance *redis.Client
}

func NewRedis(prefix string, client *redis.Client) (*Redis, error) {
	return &Redis{
		prefix:   prefix,
		instance: client,
	}, nil
}

// Add Driver an item in the cache if the key does not exist.
func (r *Redis) Add(ctx context.Context, key string, value any, t time.Duration) bool {
	val, err := r.instance.SetNX(ctx, r.key(key), value, t).Result()
	if err != nil {
		return false
	}
	return val
}

func (r *Redis) Decrement(ctx context.Context, key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	res, err := r.instance.DecrBy(ctx, r.key(key), int64(value[0])).Result()
	return int(res), err
}

// Forever Driver an item in the cache indefinitely.
func (r *Redis) Forever(ctx context.Context, key string, value any) bool {
	if err := r.Put(ctx, key, value, 0); err != nil {
		return false
	}
	return true
}

// Forget Remove an item from the cache.
func (r *Redis) Forget(ctx context.Context, key string) bool {
	_, err := r.instance.Del(ctx, r.key(key)).Result()
	return err == nil
}

// Flush Remove all items from the cache.
func (r *Redis) Flush(ctx context.Context) bool {
	res, err := r.instance.FlushAll(ctx).Result()
	if err != nil || res != "OK" {
		return false
	}
	return true
}

// Get Retrieve an item from the cache by key.
func (r *Redis) Get(ctx context.Context, key string, def ...any) any {
	val, err := r.instance.Get(ctx, r.key(key)).Result()
	if err != nil {
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

	return val
}

func (r *Redis) GetBool(ctx context.Context, key string, def ...bool) bool {
	if len(def) == 0 {
		def = append(def, false)
	}
	res := r.Get(ctx, key, def[0])
	if val, ok := res.(string); ok {
		return val == "1"
	}
	return cast.ToBool(res)
}

func (r *Redis) GetInt(ctx context.Context, key string, def ...int) int {
	if len(def) == 0 {
		def = append(def, 1)
	}
	res := r.Get(ctx, key, def[0])
	if val, ok := res.(string); ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return def[0]
		}
		return i
	}
	return cast.ToInt(res)
}

func (r *Redis) GetInt64(ctx context.Context, key string, def ...int64) int64 {
	if len(def) == 0 {
		def = append(def, 1)
	}
	res := r.Get(ctx, key, def[0])
	if val, ok := res.(string); ok {
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return def[0]
		}
		return i
	}
	return cast.ToInt64(res)
}

func (r *Redis) GetString(ctx context.Context, key string, def ...string) string {
	if len(def) == 0 {
		def = append(def, "")
	}
	return cast.ToString(r.Get(ctx, key, def[0]))
}

// Has Check an item exists in the cache.
func (r *Redis) Has(ctx context.Context, key string) bool {
	value, err := r.instance.Exists(ctx, r.key(key)).Result()
	if err != nil || value == 0 {
		return false
	}
	return true
}

func (r *Redis) Increment(ctx context.Context, key string, value ...int) (int, error) {
	if len(value) == 0 {
		value = append(value, 1)
	}
	res, err := r.instance.IncrBy(ctx, r.key(key), int64(value[0])).Result()
	return int(res), err
}

// Put Driver an item in the cache for a given time.
func (r *Redis) Put(ctx context.Context, key string, value any, t time.Duration) error {
	err := r.instance.Set(ctx, r.key(key), value, t).Err()
	if err != nil {
		return err
	}

	return nil
}

// Pull Retrieve an item from the cache and delete it.
func (r *Redis) Pull(ctx context.Context, key string, def ...any) any {
	var res any
	if len(def) == 0 {
		res = r.Get(ctx, key)
	} else {
		res = r.Get(ctx, key, def[0])
	}
	r.Forget(ctx, key)

	return res
}

// Remember Get an item from the cache, or execute the given Closure and store the result.
func (r *Redis) Remember(ctx context.Context, key string, seconds time.Duration, callback func() (any, error)) (any, error) {
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
func (r *Redis) RememberForever(ctx context.Context, key string, callback func() (any, error)) (any, error) {
	val := r.Get(ctx, key, nil)

	if val != nil {
		return val, nil
	}

	var err error
	val, err = callback()
	if err != nil {
		return nil, err
	}

	if err := r.Put(ctx, key, val, 0); err != nil {
		return nil, err
	}

	return val, nil
}

func (r *Redis) key(key string) string {
	return r.prefix + key
}
