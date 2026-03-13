package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		st := time.Now()
		var reqBody []byte
		//上传文件的不记录
		if ctx.Request.Body != nil && !strings.HasPrefix(ctx.Request.Header.Get("Content-Type"), "multipart/form-data") {
			reqBody, _ = ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw

		zapFields := make([]zap.Field, 0, 6)
		zapFields = append(zapFields, zap.String("method", ctx.Request.Method))
		zapFields = append(zapFields, zap.String("url", ctx.Request.URL.String()))
		zapFields = append(zapFields, zap.Any("header", ctx.Request.Header))
		ctx.Next()
		zapFields = append(zapFields, zap.Duration("runTime", time.Now().Sub(st)))
		zapFields = append(zapFields, zap.ByteString("reqBody", reqBody))
		zapFields = append(zapFields, zap.ByteString("resBody", blw.body.Bytes()))
		log.Debug("Request", zapFields...)
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
