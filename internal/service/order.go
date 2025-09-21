package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/repository"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrOrderIDWrongFormat = errors.New("order id should be a number")
	ErrOrderIDLuhnCheck   = errors.New("order id should be a valid number")
	finalStatuses         = map[string]bool{
		"INVALID":   true,
		"PROCESSED": true,
	}
)

type OrderService struct {
	storage       repository.Storage
	logger        *zap.Logger
	accrualSystem string
}

func NewOrderService(storage repository.Storage, logger *zap.Logger, accrualSystemAddress string) *OrderService {
	return &OrderService{
		storage:       storage,
		logger:        logger,
		accrualSystem: accrualSystemAddress,
	}
}

func (s *OrderService) ProcessOrder(userID int32, orderID string) error {
	_, err := strconv.Atoi(orderID)
	if err != nil {
		return ErrOrderIDWrongFormat
	}
	isValid := isValidLuhn(orderID)
	if !isValid {
		return ErrOrderIDLuhnCheck
	}
	err = s.storage.AddOrder(userID, orderID)
	if err != nil {
		return fmt.Errorf("error adding order: %w", err)
	}
	return nil
}
func (s *OrderService) ListOrders(userID int32) ([]model.OrderItem, error) {
	orders, err := s.storage.GetOrders(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}
	return orders, nil
}

func (s *OrderService) ProcessAccrual(orderID string) {
	err := s.storage.UpdateOrder(orderID, "PROCESSING", 0)
	if err != nil {
		s.logger.Error("error updating order", zap.Error(err))
		return
	}
	for {
		resp, err := http.Get(s.accrualSystem + "/api/orders/" + orderID)
		if err != nil {
			s.logger.Error("error getting order accrual status",
				zap.String("orderId", orderID),
				zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()
		status := resp.StatusCode
		if status == http.StatusNoContent {
			err := s.storage.UpdateOrder(orderID, "INVALID", 0)
			if err != nil {
				s.logger.Error("error updating order", zap.Error(err))
				break
			}
		}
		body, _ := io.ReadAll(resp.Body)
		var orderResp model.ExternalOrderResponse
		_ = json.Unmarshal(body, &orderResp)

		s.logger.Debug("orderResp", zap.Any("orderResp", orderResp))

		if finalStatuses[orderResp.Status] {
			err := s.storage.UpdateOrder(orderID, orderResp.Status, orderResp.Accrual)
			if err != nil {
				s.logger.Error("error updating order", zap.Error(err))
			}
			break
		}
		time.Sleep(10 * time.Second)
	}
}
func isValidLuhn(number string) bool {
	var sum int
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}
