package redis

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var script string = `
if redis.call("zcount",KEYS[1],ARGV[1],ARGV[2]) >= tonumber(ARGV[3]) then

else
	redis.call("zadd",KEYS[1],ARGV[2],ARGV[2])	
	redis.call("expire",KEYS[1],ARGV[4])
	redis.call("zremrangebyscore",KEYS[1],0,ARGV[1])
	return 0
end
`

type RateLimiterConfig struct {
	Limit            int64         `help:"窗口限制值,为0时不限制" default:"50"`
	LimitWindow      time.Duration `help:"窗口限制时间" default:"5s"`
	Prefix           string        `help:"key前缀" default:"rrl"`
	ExceptionDefault bool          `help:"异常情况下是否允许的默认值,如redis挂掉情况下" default:"true"`
}

type RateLimiterGetKey func(ctx *http.Request) string

// RateLimiter 基于redis的分布式窗口时间速率控制
type RateLimiter struct {
	limitScript      *redis.Script
	rdb              *redis.Client
	keyPre           string
	limit            int64
	limitWindow      time.Duration
	ExceptionDefault bool
}

// NewRateLimit
func NewRateLimit(conf RateLimiterConfig, rdb *redis.Client) *RateLimiter {
	r := &RateLimiter{
		limitScript:      redis.NewScript(script),
		keyPre:           conf.Prefix,
		rdb:              rdb,
		limit:            conf.Limit,
		limitWindow:      conf.LimitWindow,
		ExceptionDefault: conf.ExceptionDefault,
	}
	if r.keyPre == "" {
		r.keyPre = "rrl:"
	}
	if r.limitWindow == 0 {
		r.limitWindow = time.Second * 5
	}
	return r
}

func (r *RateLimiter) AllowRequest(req *http.Request) bool {
	key := KeyGenByURIFromRequest(req)
	return r.allow(key)
}

func (r *RateLimiter) Allow(key string) bool {
	return r.allow(key)
}

func (r *RateLimiter) allow(key string) bool {
	if r.limit == 0 || key == "" {
		return true
	}
	key = r.keyPre + key
	curTs := time.Now().UnixMilli()
	winTs := curTs - int64(r.limitWindow.Milliseconds())
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()
	val, err := r.limitScript.Run(ctx, r.rdb, []string{key}, winTs, curTs, r.limit, r.limitWindow.Seconds()).Int64()
	if err != nil {
		return r.ExceptionDefault
	}
	if val == 1 {
		return false
	}
	return true
}

func KeyGenByUidAndApiFromRequest(req *http.Request) string {
	uid := req.Header.Get("X-Uid")
	if uid == "" {
		return ""
	}
	return uid + ":" + req.URL.Path
}

func KeyGenByIpFromRequest(req *http.Request) string {
	arr := strings.Split(req.RemoteAddr, ":")
	if len(arr) > 0 {
		return arr[0]
	}
	return req.RemoteAddr
}

func KeyGenByURIFromRequest(req *http.Request) string {
	return req.URL.Path
}
