package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"mime"
	"net/http"
	"time"
)

const jsonContentType = "application/json"

type AuthHandler struct {
	authService service.Auth
	logger      *zap.Logger
	userService *service.UserService
}

func NewAuthHandler(authService service.Auth, userService *service.UserService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, userService: userService, logger: logger}
}

func (h *AuthHandler) ServeCreateHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.CreateUserRequest

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

	if !h.userService.ValidateCredentials(&req) {
		h.logger.Debug("invalid credentials", zap.Any("credentials", req))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.userService.CreateUser(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("Login occupied"))
			return
		}
		h.logger.Error("error creating user", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tokenString, err := h.authService.GenerateToken(user)
	if err != nil {
		h.logger.Error(
			"got error while generating token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Debug("User created",
		zap.Int32("userID", user.ID),
		zap.String("login", user.Login),
	)
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) ServeLoginHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeoutSec*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest

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

	user, err := h.userService.ValidateAndGetUser(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrWrongLoginOrPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Wrong login or password"))
			return
		}
	}
	tokenString, err := h.authService.GenerateToken(user)
	if err != nil {
		h.logger.Error(
			"got error while generating token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := struct {
		Token string `json:"token"`
	}{Token: tokenString}
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
