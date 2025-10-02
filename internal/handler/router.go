package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
)

func NewRouter(h *Handlers, auth service.Auth, logger *zap.Logger) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.AuthProvider()))
				r.Use(jwtauth.Authenticator(auth.AuthProvider()))
				r.Get("/withdrawals", h.WithdrawalsList)
				r.Route("/orders", func(r chi.Router) {
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
