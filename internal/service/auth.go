package service

import (
	"context"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"github.com/m3lifaro/gophermart/internal/model"
)

type Auth interface {
	GenerateToken(user *model.User) (string, error)
	ReadToken(ctx context.Context) (*model.User, error)
	AuthProvider() *jwtauth.JWTAuth
}

type AuthImpl struct {
	auth *jwtauth.JWTAuth
}

func NewAuth(secret string) *AuthImpl {
	var tokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	return &AuthImpl{
		auth: tokenAuth,
	}
}

func (h *AuthImpl) GenerateToken(user *model.User) (string, error) {
	_, token, err := h.auth.Encode(map[string]interface{}{model.LoginKey: user.Login, model.UserIDKey: user.UUID, model.EmailKey: user.Email, model.RoleKey: user.Role})
	if err != nil {
		return "", fmt.Errorf("error generating jwt token for user(%s): %w", user, err)
	}
	return token, nil
}

func (h *AuthImpl) ReadToken(ctx context.Context) (*model.User, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read jwt token from context (%s): %w", ctx, err)
	}
	login := claims[model.LoginKey].(string)
	uuid := claims[model.UserIDKey].(string)
	email := claims[model.EmailKey].(string)
	role := claims[model.RoleKey].(string)
	user := &model.User{
		Login: login,
		UUID:  uuid,
		Email: email,
		Role:  role,
	}
	return user, nil
}

func (h *AuthImpl) AuthProvider() *jwtauth.JWTAuth {
	return h.auth
}
