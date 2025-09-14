package main

import (
	"github.com/m3lifaro/gophermart/cmd/config"
	"github.com/m3lifaro/gophermart/internal/logger"
	"log"
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
}
