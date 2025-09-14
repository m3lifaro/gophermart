package handler

import (
	"encoding/json"
	"github.com/m3lifaro/gophermart/internal/model"
	"go.uber.org/zap"
	"mime"
	"net/http"
)

const jsonContentType = "application/json"

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
	h.logger.Debug("got request", zap.String("username", req.Login))
	w.Header().Set("Content-Type", jsonContentType)
	w.WriteHeader(http.StatusOK)
}
