package server

import (
	"context"
	"net/url"
)

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type Endpointer interface {
	Endpointer() (*url.URL, error)
}
