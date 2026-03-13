package server

import (
	"rest_demo/pkg/server/http"

	"go.uber.org/zap"
)

func NewHTTPServerFiber(
	log *zap.Logger,
	conf *http.Config,
) *http.FiberServer {
	s := http.NewServerFiber(log, conf)

	return s
}
