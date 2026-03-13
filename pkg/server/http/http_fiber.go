package http

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type FiberServer struct {
	*fiber.App
	httpSrv *fasthttp.Server
	logger  *zap.Logger
	config  *Config
}

func NewServerFiber(logger *zap.Logger, config *Config) *FiberServer {
	s := &FiberServer{
		App:    fiber.New(fiber.Config{}),
		logger: logger,
		config: config,
	}
	return s
}

func (s *FiberServer) Start(ctx context.Context) error {
	s.httpSrv = &fasthttp.Server{
		Handler: s.Handler(),
		Name:    "fiber app",
	}
	go func() {
		if err := s.httpSrv.ListenAndServe(":4000"); err != nil {
			// if err == fasthttp.ErrServerClosed {
			// 	log.Println("服务器已关闭")
			// } else {
			log.Fatalf("服务器错误: %v", err)
			// }

		}
	}()
	return nil
}

func (s *FiberServer) Stop(ctx context.Context) error {
	s.logger.Sugar().Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.ShutdownWithContext(ctx); err != nil {
		log.Printf("HTTP服务器关闭错误: %v", err)
	}
	return nil
}
