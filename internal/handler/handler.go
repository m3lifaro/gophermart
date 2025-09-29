package handler

import (
	"github.com/m3lifaro/gophermart/internal/concurrent"
	"github.com/m3lifaro/gophermart/internal/service"
	"net/http"

	"go.uber.org/zap"
)

const defaultTimeoutSec = 30

type Handlers struct {
	Register        http.HandlerFunc
	Login           http.HandlerFunc
	CreateOrder     http.HandlerFunc
	OrderList       http.HandlerFunc
	Withdraw        http.HandlerFunc
	UserBalance     http.HandlerFunc
	WithdrawalsList http.HandlerFunc
}

func NewHandlers(authService service.Auth, userService *service.UserService, orderService *service.OrderService, logger *zap.Logger, wp *concurrent.WorkerPool) *Handlers {
	return &Handlers{
		Register:        NewAuthHandler(authService, userService, logger).ServeCreateHTTP,
		Login:           NewAuthHandler(authService, userService, logger).ServeLoginHTTP,
		CreateOrder:     NewOrderHandler(authService, orderService, logger, wp).ServeCreateOrderHTTP,
		OrderList:       NewOrderHandler(authService, orderService, logger, wp).ServeListOrdersHTTP,
		Withdraw:        NewBalanceHandler(authService, orderService, logger).ServeWithdrawHTTP,
		UserBalance:     NewBalanceHandler(authService, orderService, logger).ServeGetBalanceHTTP,
		WithdrawalsList: NewBalanceHandler(authService, orderService, logger).ServeGetWithdrawalsHTTP,
	}
}
