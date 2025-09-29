package repository

import (
	"context"
	"errors"
	"github.com/m3lifaro/gophermart/internal/model"
)

var (
	ErrUserExists                   = errors.New("user already exists")
	ErrUserNotFound                 = errors.New("user not found")
	ErrOrderAlreadyProcessed        = errors.New("current order already processed")
	ErrOrderAlreadyProcessedByOther = errors.New("order processed by other user")
	ErrInsufficientFunds            = errors.New("insufficient funds")
)

type Storage interface {
	GetUserByLogin(ctx context.Context, login string) (*model.UserDao, error)
	CreateUser(ctx context.Context, user *model.UserDao) error
	AddOrder(ctx context.Context, userID int32, orderID string) error
	GetOrders(ctx context.Context, userID int32) ([]model.OrderItem, error)
	UpdateOrder(ctx context.Context, orderID, status string, amount float64, userID int32) error
	WithdrawBonuses(ctx context.Context, userID int32, orderID string, amount float64) error
	GetBalance(ctx context.Context, userID int32) (*model.UserBalance, error)
	GetWithdrawals(ctx context.Context, userID int32) ([]model.WithdrawItem, error)
}
