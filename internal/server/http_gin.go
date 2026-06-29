package server

import (
	"rest_demo/internal/handler"
	"rest_demo/internal/middleware"
	"rest_demo/pkg/cfgstruct"
	"rest_demo/pkg/jwt"
	"rest_demo/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewHTTPServer(
	log *zap.Logger,
	conf *http.Config,
	jwt *jwt.JWT,
	loginHandler *handler.LoginHandler,
	llmHandler *handler.LLMHandler,
	redisClient *redis.Client,
) *http.Server {
	switch cfgstruct.DefaultsType() {
	case cfgstruct.DefaultsRelease:
		gin.SetMode(gin.ReleaseMode)
	case cfgstruct.DefaultsTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	s := http.NewServer(gin.Default(), log, conf)

	// 跨域
	s.Use(middleware.CORSMiddleware())
	s.Use(middleware.LoginRateLimit(redisClient))
	s.Use(middleware.RequestLogMiddleware(log))
	// s.Use(middleware.Auth(jwt, log))

	s.POST("login", loginHandler.Login)

	s.POST("create-order", loginHandler.CreateOrder)

	s.GET("/api", loginHandler.GetApiList(s.Engine))

	s.POST("wsplay", loginHandler.Wsplay)

	s.GET("/stream", loginHandler.Stream)

	s.POST("llm/chat", llmHandler.Chat)
	s.POST("llm/workflow", llmHandler.Workflow)
	s.GET("llm/topics", llmHandler.GetTopics)

	return s
}
