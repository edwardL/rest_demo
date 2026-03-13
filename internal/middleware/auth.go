package middleware

import (
	"rest_demo/internal/constant"
	"rest_demo/pkg/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Auth(j *jwt.JWT, log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.Request.Header.Get("Authorization")
		if len(tokenStr) > 0 && tokenStr[0:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		//头里面没有获取到，可能是ws
		if tokenStr == "" {
			//websocket
			tokenStr = ctx.Request.Header.Get("Sec-WebSocket-Protocol")
		}
		if tokenStr == "" {
			log.Error("Authorization token is empty", zap.String("url", ctx.Request.URL.String()))
			ctx.AbortWithStatus(401)
			return
		}
		tp, err := j.ValidateToken(tokenStr)
		if err != nil {
			log.Error("Authorization token error", zap.String("url", ctx.Request.URL.String()), zap.String("token", tokenStr), zap.Error(err))
			ctx.AbortWithStatus(401)
			return
		}
		ctx.Set(constant.CtxTokenKey, tp)
		ctx.Next()
	}
}
