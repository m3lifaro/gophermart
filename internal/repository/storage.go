package repository

import (
	"errors"
	"github.com/m3lifaro/gophermart/internal/model"
)

var (
	ErrUserExists                   = errors.New("user already exists")
	ErrUserNotFound                 = errors.New("user not found")
	ErrOrderAlreadyProcessed        = errors.New("current order already processed")
	ErrOrderAlreadyProcessedByOther = errors.New("order processed by other user")
)

type Storage interface {
	GetUserByLogin(login string) (*model.UserDao, error)
	CreateUser(user *model.UserDao) error
	AddOrder(userID int32, orderID string) error
	GetOrders(userID int32) ([]model.OrderItem, error)
	UpdateOrder(orderID, status string, amount float64) error
}

//type MemoryStorage struct {
//	mu    sync.RWMutex
//	users map[string]*model.UserDao
//}
//
//func NewMemoryStorage() *MemoryStorage {
//	return &MemoryStorage{
//		users: make(map[string]*model.UserDao),
//		mu:    sync.RWMutex{},
//	}
//}
//
//func (s *MemoryStorage) GetUserByLogin(login string) (*model.UserDao, error) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	user, ok := s.users[login]
//	if !ok {
//		return nil, ErrUserNotFound
//	}
//	return user, nil
//}
//
//func (s *MemoryStorage) CreateUser(user *model.UserDao) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	_, ok := s.users[user.Login]
//	if ok {
//		return ErrUserExists
//	}
//	s.users[user.Login] = user
//	return nil
//}
