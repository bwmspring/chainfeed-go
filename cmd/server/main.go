package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chainfeed-go/internal/config"
	"chainfeed-go/internal/database"
	"chainfeed-go/internal/server"
	"chainfeed-go/pkg/logger"
	"go.uber.org/zap"
)

var configPath = flag.String("config", "config/config.yaml", "path to config file")

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(cfg.Log)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Connect to PostgreSQL
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	zapLogger.Info("Connected to PostgreSQL")

	// Connect to Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		zapLogger.Fatal("Failed to connect to redis", zap.Error(err))
	}
	defer rdb.Close()
	zapLogger.Info("Connected to Redis")

	// Create server
	srv := server.New(cfg, zapLogger, db, rdb)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			zapLogger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}
