package handler

import (
	"github.com/m3lifaro/gophermart/internal/service"
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Register http.HandlerFunc
	Login    http.HandlerFunc
	Orders   http.HandlerFunc
}

func NewHandlers(authService service.Auth, userService *service.UserService, orderService *service.OrderService, logger *zap.Logger) *Handlers {
	return &Handlers{
		Register: NewAuthHandler(authService, userService, logger).ServeCreateHTTP,
		Login:    NewAuthHandler(authService, userService, logger).ServeLoginHTTP,
		Orders:   NewOrderHandler(authService, orderService, logger).ServeCreateOrderHTTP,
	}
}
