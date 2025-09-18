package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"mime"
	"net/http"
)

const jsonContentType = "application/json"

var users = make(map[string]string) // login: hashedPassword
var usersMap = make(map[string]model.User)

type AuthHandler struct {
	authService service.Auth
	logger      *zap.Logger
}

func NewAuthHandler(authService service.Auth, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}
func validateCredentials(req *model.CreateUserRequest) bool {
	//todo
	return req.Login != "" && req.Password != ""
}
func (h *AuthHandler) ServeCreateHTTP(w http.ResponseWriter, r *http.Request) {
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

	if !validateCredentials(&req) {
		h.logger.Error("invalid credentials", zap.Any("credentials", req))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	users[req.Login] = string(hash)
	h.logger.Debug("got request", zap.String("username", req.Login))
	user := &model.User{
		Login: req.Login,
		UUID:  uuid.New().String(),
	}
	usersMap[req.Login] = *user
	tokenString, err := h.authService.GenerateToken(user)
	if err != nil {
		h.logger.Error(
			"got error while generating token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) ServeLoginHTTP(w http.ResponseWriter, r *http.Request) {
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

	hash, ok := users[req.Login]
	if !ok || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	user, ok := usersMap[req.Login]
	tokenString, err := h.authService.GenerateToken(&user)
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

func (h *AuthHandler) ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
	user, err := h.authService.ReadToken(r.Context())
	if err != nil {
		h.logger.Error(
			"got error while reading token",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Welcome, " + user.Login + " ID: " + user.UUID))
}
