package handler

import (
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Auth http.HandlerFunc
}

func NewHandlers(logger *zap.Logger) *Handlers {
	return &Handlers{
		Auth: NewAuthHandler(logger).ServeCreateHTTP,
	}
}
