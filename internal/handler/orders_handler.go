package handler

import (
	"errors"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"mime"
	"net/http"
)

type OrderHandler struct {
	authService  service.Auth
	logger       *zap.Logger
	orderService *service.OrderService
}

func NewOrderHandler(authService service.Auth, orderService *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		authService:  authService,
		orderService: orderService,
		logger:       logger,
	}
}

func (h *OrderHandler) ServeCreateOrderHTTP(w http.ResponseWriter, r *http.Request) {
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

	err = h.orderService.ProcessOrder(user.ID, orderNum)
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
	w.WriteHeader(http.StatusAccepted)
}
