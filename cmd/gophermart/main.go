package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/m3lifaro/gophermart/cmd/config"
	"github.com/m3lifaro/gophermart/internal/concurrent"
	"github.com/m3lifaro/gophermart/internal/handler"
	"github.com/m3lifaro/gophermart/internal/logger"
	"github.com/m3lifaro/gophermart/internal/repository"
	"github.com/m3lifaro/gophermart/internal/service"
	"github.com/pressly/goose/v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	authService := service.NewAuth("secret-key")

	workerPool := concurrent.NewWorkerPool(cfg.MaxParallel)
	userService := service.NewUserService(storage, zl)
	orderService := service.NewOrderService(storage, zl, cfg.AccrualSystem)

	workerPool.Start(orderService, zl)

	handlers := handler.NewHandlers(authService, userService, orderService, zl, workerPool)
	r := handler.NewRouter(handlers, authService, zl)
	server := &http.Server{
		Addr:    cfg.ServeAddress,
		Handler: r,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server started on %s", cfg.ServeAddress)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Printf("HTTP server error: %v", err)
	case sig := <-stop:
		log.Printf("Received shutdown signal: %v", sig)
	}

	log.Println("Shutting down server and worker pool...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	workerPool.Shutdown()
	log.Println("Graceful shutdown completed.")
}
