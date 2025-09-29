package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/m3lifaro/gophermart/internal/concurrent"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"mime"
	"net/http"
	"time"
)

type OrderHandler struct {
	authService  service.Auth
	logger       *zap.Logger
	orderService *service.OrderService
	wp           *concurrent.WorkerPool
}

func NewOrderHandler(authService service.Auth, orderService *service.OrderService, logger *zap.Logger, wp *concurrent.WorkerPool) *OrderHandler {
	return &OrderHandler{
		authService:  authService,
		orderService: orderService,
		logger:       logger,
		wp:           wp,
	}
}

func (h *OrderHandler) ServeCreateOrderHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, err := h.authService.ReadToken(r.Context())
	if err != nil {
		h.logger.Error(
			"got error while reading token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Content-Type header"))
		return
	}

	if mediaType != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'text/plain', 'plain/text' or 'application/x-gzip', got: " + mediaType))
		return
	}

	orderNum := string(body)

	err = h.orderService.ProcessOrder(ctx, user.ID, orderNum)
	if err != nil {
		if errors.Is(err, service.ErrOrderIDWrongFormat) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrOrderIDLuhnCheck) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyProcessed) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, repository.ErrOrderAlreadyProcessedByOther) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		h.logger.Error(
			"got error while processing order",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.wp.AddJob(orderNum, user.ID)
	h.logger.Debug("Order created",
		zap.String("orderNum", orderNum),
		zap.Int32("userID", user.ID),
	)
	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) ServeListOrdersHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	user, err := h.authService.ReadToken(r.Context())
	if err != nil {
		h.logger.Error(
			"got error while reading token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", jsonContentType)
	orders, err := h.orderService.ListOrders(ctx, user.ID)
	if err != nil {
		h.logger.Error(
			"got error while processing order",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.Debug("Orders served",
		zap.Int("orders_count", len(orders)),
		zap.Int32("userID", user.ID),
		zap.String("login", user.Login),
		zap.Any("orders", orders),
	)
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("Failed to encode user orders response", zap.Error(err))
	}
}
