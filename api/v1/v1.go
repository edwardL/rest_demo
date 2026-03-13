package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code    int    `json:"code"`    //0为成功，非0失败
	Message string `json:"message"` //失败消息
	Result  any    `json:"result"`
}

type PageResult struct {
	Total int64 `json:"total"` //总数量
	Items any   `json:"items"` //数据列表
}

func Success(ctx *gin.Context, data any) {
	if data == nil {
		data = gin.H{}
	}
	ctx.AbortWithStatusJSON(http.StatusOK, Resp{Code: 0, Message: "success", Result: data})
}

func Error(ctx *gin.Context, err error, data ...any) {
	code := 5000
	ctx.AbortWithStatusJSON(http.StatusOK, Resp{Code: code, Message: err.Error(), Result: data})
}

func Response(ctx *gin.Context, err error, data any) {
	if err != nil {
		Error(ctx, err, data)
	} else {
		Success(ctx, data)
	}
}

func Failed(ctx *gin.Context, code int, err error) {
	ctx.AbortWithError(code, err)
}
