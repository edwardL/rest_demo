//go:build wireinject
// +build wireinject

package wire

import (
	"context"
	"rest_demo/internal/handler"
	"rest_demo/internal/repository"
	"rest_demo/internal/server"
	"rest_demo/internal/service"
	"rest_demo/pkg/app"
	"rest_demo/pkg/cache"
	"rest_demo/pkg/db"
	"rest_demo/pkg/jwt"
	"rest_demo/pkg/llm"
	"rest_demo/pkg/payment/wechat"
	"rest_demo/pkg/redis"
	"rest_demo/pkg/websocket"

	"rest_demo/pkg/server/http"

	"github.com/google/wire"
	"go.uber.org/zap"
)

// var pkgSet = wire.NewSet()

var repositorySet = wire.NewSet(
	db.NewMsDB,
	jwt.NewJWT,
	redis.NewRedis,
	cache.NewCache,
	wechat.NewJSAPIClient,
	websocket.NewWsManager,
	repository.NewRepository,
	repository.NewSysUserRepository,
)

var serviceSet = wire.NewSet(
	service.NewService,
	service.NewLoginService,
	service.NewLLMService,
)

var handlerSet = wire.NewSet(
	handler.NewLoginHandler,
	handler.NewLLMHandler,
)

var serverSet = wire.NewSet(
	server.NewHTTPServer,
	// server.NewHTTPServerFiber,
)

// build App
func newApp(httpServer *http.Server) *app.App {
	return app.NewApp(
		app.WithServer(httpServer),
		// app.WithServer(httpServer, job),
		app.WithName("hobyadm-server"),
	)
}

func NewApp(context.Context, *zap.Logger,
	*db.MsConfig,
	*http.Config,
	*service.Config,
	*jwt.Config,
	*redis.Config,
	*wechat.WeChatPayConfig,
	*cache.Config,
	*llm.Config,
) (*app.App, func(), error) {
	panic(wire.Build(repositorySet, serviceSet, handlerSet, serverSet, newApp))
}
