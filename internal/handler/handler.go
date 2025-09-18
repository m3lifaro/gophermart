package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Register  http.HandlerFunc
	Login     http.HandlerFunc
	Protected http.HandlerFunc
}

func NewHandlers(logger *zap.Logger) *Handlers {
	return &Handlers{
		Register:  NewAuthHandler(logger).ServeCreateHTTP,
		Login:     NewAuthHandler(logger).ServeLoginHTTP,
		Protected: NewAuthHandler(logger).ProtectedEndpoint,
	}
}
