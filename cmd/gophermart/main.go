package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/m3lifaro/gophermart/cmd/config"
	"github.com/m3lifaro/gophermart/internal/handler"
	"github.com/m3lifaro/gophermart/internal/logger"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"github.com/pressly/goose/v3"
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
	var storage repository.Storage
	if cfg.DBDsn != "" {
		pool, err := pgxpool.New(context.Background(), cfg.DBDsn)
		if err != nil {
			log.Fatalf("Failed to initialize connection pool: %v", err)
		}

		db := stdlib.OpenDBFromPool(pool)

		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatal("goose up failed:", err)
		}

		log.Println("Migrations applied successfully")

		if err := db.Close(); err != nil {
			log.Fatal("DB wasn't closed:", err)
		}
		storage = repository.NewPGStorage(pool, zl)
	} else {
		//storage = repository.NewMemoryStorage()
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	zl.Info("Hello world!")
	authService := service.NewAuth("secret-key")

	userService := service.NewUserService(storage, zl)
	orderService := service.NewOrderService(storage, zl, cfg.AccrualSystem)
	handlers := handler.NewHandlers(authService, userService, orderService, zl)
	r := handler.NewRouter(handlers, authService, zl)
	log.Printf("Server started on %s", cfg.ServeAddress)
	log.Fatal(http.ListenAndServe(cfg.ServeAddress, r))
}
