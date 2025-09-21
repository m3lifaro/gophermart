package handler

import (
	"encoding/json"
	"errors"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"mime"
	"net/http"
)

type BalanceHandler struct {
	authService  service.Auth
	logger       *zap.Logger
	orderService *service.OrderService
}

func NewBalanceHandler(authService service.Auth, orderService *service.OrderService, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		authService:  authService,
		orderService: orderService,
		logger:       logger,
	}
}

func (h *BalanceHandler) ServeWithdrawHTTP(w http.ResponseWriter, r *http.Request) {
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

	var req model.WithdrawalRequest

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&req); err != nil {
		h.logger.Error(
			"got error, while decoding HTTP request",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentHeader := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentHeader)
	if err != nil || mediaType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported Content-Type. Expected 'application/json', got: " + mediaType))
		return
	}
	defer r.Body.Close()

	err = h.orderService.ProcessWithdrawal(user.ID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, service.ErrOrderIDLuhnCheck) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, repository.ErrInsufficientFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		h.logger.Error(
			"got error while processing order",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Debug("Withdrawal for order successfully processed",
		zap.String("orderNum", req.Order),
		zap.Int32("userID", user.ID),
		zap.Float64("amount", req.Sum),
	)
	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) ServeGetBalanceHTTP(w http.ResponseWriter, r *http.Request) {
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
	balance, err := h.orderService.GetUserBalance(user.ID)
	if err != nil {
		h.logger.Error(
			"got error while getting user balance",
			zap.Int32("userID", user.ID),
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("Failed to encode user balance response", zap.Error(err))
	}
}

func (h *BalanceHandler) ServeGetWithdrawalsHTTP(w http.ResponseWriter, r *http.Request) {
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
	withdrawals, err := h.orderService.GetWithdrawals(user.ID)
	if err != nil {
		h.logger.Error(
			"got error while getting user withdrawals",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.logger.Debug("Withdrawals served",
		zap.Int("withdrawals_count", len(withdrawals)),
		zap.Int32("userID", user.ID),
		zap.String("login", user.Login),
		zap.Any("withdrawals", withdrawals),
	)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("Failed to encode user withdrawals response", zap.Error(err))
	}
}
