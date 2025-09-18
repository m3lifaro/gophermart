package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

func NewRouter(h *Handlers, logger *zap.Logger) chi.Router {
	//func NewRouter(h *Handlers, logger *zap.Logger, auth *auth.AuthImpl) chi.Router {
	r := chi.NewRouter()
	//r.Use(gzipMiddleware(logger))
	//r.Use(LoggingMiddleware(logger))

	r.Group(func(r chi.Router) {
		//r.Get("/ping", h.Ping)
	})

	r.Group(func(r chi.Router) {
		//r.Use(authMiddleware(logger, auth))
		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", h.Register)
			r.Post("/login", h.Login)
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(tokenAuth))
				r.Use(jwtauth.Authenticator(tokenAuth))
				r.Get("/protected", h.Protected)
			})
			//r.Post("/api/shorten", h.ShortenJSON)
			//r.Post("/api/shorten/batch", h.BatchShorten)
			//r.Get("/{id}", h.Redirect)
		})
	})

	return r
}
