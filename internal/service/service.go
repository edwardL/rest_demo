package service

import (
	"rest_demo/pkg/cache"
	"rest_demo/pkg/jwt"
	"rest_demo/pkg/websocket"

	"go.uber.org/zap"
)

type Config struct {
}

type Service struct {
	config   *Config
	logger   *zap.Logger
	jwt      *jwt.JWT
	cache    cache.Cache
	wsClient *websocket.ClientManager
}

func NewService(conf *Config, logger *zap.Logger, jwt *jwt.JWT,
	cache cache.Cache,
	wsClient *websocket.ClientManager,
) *Service {
	return &Service{
		config:   conf,
		logger:   logger,
		jwt:      jwt,
		cache:    cache,
		wsClient: wsClient,
	}
}
