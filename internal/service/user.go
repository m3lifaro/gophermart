package service

import (
	"fmt"
	internal "github.com/m3lifaro/gophermart/internal/errors"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"regexp"
)

var (
	loginRegexp    = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	passwordRegexp = regexp.MustCompile(`^[A-Za-z\d!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]{8,64}$`)
)

type UserService struct {
	storage repository.Storage
	logger  *zap.Logger
}

func NewUserService(storage repository.Storage, logger *zap.Logger) *UserService {
	return &UserService{
		storage: storage,
		logger:  logger,
	}
}

func (s *UserService) CreateUser(login, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	var user = &model.UserDao{
		Password: string(hash),
		User: model.User{
			Login: login,
		},
	}
	err = s.storage.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("got error, while creating user: %w", err)
	}
	return &user.User, nil
}

func (s *UserService) GetUserByLogin(login string) (*model.User, error) {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("got error, while getting user: %w", err)
	}
	return &user.User, nil
}

func (s *UserService) ValidateAndGetUser(login, password string) (*model.User, error) {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("got error, while getting user: %w", err) //todo process missing login
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, internal.ErrWrongLoginOrPassword
	}
	return &user.User, nil
}

func (s *UserService) ValidateCredentials(req *model.CreateUserRequest) bool {
	if req == nil {
		return false
	}
	if !loginRegexp.MatchString(req.Login) {
		return false
	}
	if !passwordRegexp.MatchString(req.Password) {
		return false
	}
	return true
}
