package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
)

func NewRouter(h *Handlers, auth service.Auth, logger *zap.Logger) chi.Router {
	//func NewRouter(h *Handlers, logger *zap.Logger, auth *auth.AuthImpl) chi.Router {
	r := chi.NewRouter()
	//r.Use(gzipMiddleware(logger))
	//r.Use(LoggingMiddleware(logger))

	r.Group(func(r chi.Router) {
		//r.Use(authMiddleware(logger, auth))
		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.AuthProvider()))
				r.Use(jwtauth.Authenticator(auth.AuthProvider()))
				r.Route("/order", func(r chi.Router) {
					r.Post("/", h.CreateOrder)
					r.Get("/", h.OrderList)
				})
				r.Route("/balance", func(r chi.Router) {
					r.Get("/", h.UserBalance)
					r.Post("/withdraw", h.Withdraw)
				})
			})
		})
	})

	return r
}
