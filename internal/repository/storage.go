package repository

import (
	"fmt"
	err "github.com/m3lifaro/gophermart/internal/errors"
	"github.com/m3lifaro/gophermart/internal/model"
	"sync"
)

type Storage interface {
	GetUserByLogin(login string) (*model.UserDao, error)
	CreateUser(user *model.UserDao) error
}

type MemoryStorage struct {
	mu    sync.RWMutex
	users map[string]*model.UserDao
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users: make(map[string]*model.UserDao),
		mu:    sync.RWMutex{},
	}
}

func (s *MemoryStorage) GetUserByLogin(login string) (*model.UserDao, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[login]
	if !ok {
		return nil, fmt.Errorf("user with login: '%s' not found", login)
	}
	return user, nil
}

func (s *MemoryStorage) CreateUser(user *model.UserDao) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.users[user.Login]
	if ok {
		return err.ErrUserExists
	}
	s.users[user.Login] = user
	return nil
}
