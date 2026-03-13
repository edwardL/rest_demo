package middleware

import (
	v1 "rest_demo/api/v1"
	"rest_demo/api/v1/errcode"
	redis2 "rest_demo/pkg/redis"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func LoginRateLimit(j *redis.Client) gin.HandlerFunc {
	limiter := redis2.NewRateLimit(redis2.RateLimiterConfig{
		Limit:       1000,
		LimitWindow: time.Second * 30,
		Prefix:      "adm:",
	}, j)
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/login" {
			if !limiter.AllowRequest(ctx.Request) {
				v1.Error(ctx, errcode.ErrUserLoginDisabled)
				return
			}
		}
		// ctx.Next()
	}
}
