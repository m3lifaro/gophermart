package handler

import (
	"encoding/json"
	"github.com/go-chi/jwtauth/v5"
	"github.com/m3lifaro/gophermart/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"mime"
	"net/http"
)

const jsonContentType = "application/json"

var users = make(map[string]string) // login: hashedPassword
var tokenAuth = jwtauth.New("HS256", []byte("secret-key"), nil)

type AuthHandler struct {
	logger *zap.Logger
}

func NewAuthHandler(logger *zap.Logger) *AuthHandler {
	return &AuthHandler{logger: logger}
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
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"login": req.Login, "ID": "bimbo"})
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
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"login": req.Login, "ID": "bimbo"})
	response := struct {
		Token string `json:"token"`
	}{Token: tokenString}
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	login := claims["login"]
	ID := claims["ID"]
	w.Write([]byte("Welcome, " + login.(string) + " ID: " + ID.(string)))
}
