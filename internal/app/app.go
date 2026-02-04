package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"chainfeed-go/internal/config"
	"chainfeed-go/internal/database"
	"chainfeed-go/internal/server"
	"chainfeed-go/pkg/logger"
)

type App struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sqlx.DB
	redis  *redis.Client
	server *server.Server
}

func New(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	// Initialize logger
	zapLogger, err := logger.New(cfg.Log)
	if err != nil {
		return nil, err
	}

	// Connect to PostgreSQL
	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
		return nil, err
	}
	zapLogger.Info("Connected to PostgreSQL")

	// Connect to Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		zapLogger.Fatal("Failed to connect to redis", zap.Error(err))
		return nil, err
	}
	zapLogger.Info("Connected to Redis")

	// Create server
	srv := server.New(cfg, zapLogger, db, rdb)

	return &App{
		cfg:    cfg,
		logger: zapLogger,
		db:     db,
		redis:  rdb,
		server: srv,
	}, nil
}

func (a *App) Run() error {
	// Start server in goroutine
	go func() {
		if err := a.server.Start(); err != nil {
			a.logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Fatal("Server forced to shutdown", zap.Error(err))
		return err
	}

	// Close database connections
	a.db.Close()
	a.redis.Close()
	a.logger.Sync()

	a.logger.Info("Server exited")
	return nil
}
