package main

import (
	"github.com/m3lifaro/gophermart/cmd/config"
	"github.com/m3lifaro/gophermart/internal/handler"
	"github.com/m3lifaro/gophermart/internal/logger"
	"github.com/m3lifaro/gophermart/internal/service"
	"log"
	"net/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Config: %v", cfg)
	zl, err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	zl.Info("Hello world!")
	authService := service.NewAuth("secret-key")
	handlers := handler.NewHandlers(authService, zl)
	r := handler.NewRouter(handlers, authService, zl)
	log.Printf("Server started on %s", cfg.ServeAddress)
	log.Fatal(http.ListenAndServe(cfg.ServeAddress, r))
}
