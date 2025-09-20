package handler

import (
	"github.com/m3lifaro/gophermart/internal/service"
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Register  http.HandlerFunc
	Login     http.HandlerFunc
	Protected http.HandlerFunc
}

func NewHandlers(authService service.Auth, userService *service.UserService, logger *zap.Logger) *Handlers {
	return &Handlers{
		Register:  NewAuthHandler(authService, userService, logger).ServeCreateHTTP,
		Login:     NewAuthHandler(authService, userService, logger).ServeLoginHTTP,
		Protected: NewAuthHandler(authService, userService, logger).ProtectedEndpoint,
	}
}
