package service

import (
	"errors"
	"fmt"
	"github.com/m3lifaro/gophermart/internal/repository"
	"go.uber.org/zap"
	"strconv"
)

var (
	ErrOrderIDWrongFormat = errors.New("order id should be a number")
	ErrOrderIDLuhnCheck   = errors.New("order id should be a valid number")
)

type OrderService struct {
	storage repository.Storage
	logger  *zap.Logger
}

func NewOrderService(storage repository.Storage, logger *zap.Logger) *OrderService {
	return &OrderService{
		storage: storage,
		logger:  logger,
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
