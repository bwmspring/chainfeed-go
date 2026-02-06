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

	"github.com/bwmspring/chainfeed-go/internal/config"
	"github.com/bwmspring/chainfeed-go/internal/database"
	"github.com/bwmspring/chainfeed-go/internal/server"
	"github.com/bwmspring/chainfeed-go/internal/service"
	"github.com/bwmspring/chainfeed-go/internal/websocket"
	"github.com/bwmspring/chainfeed-go/pkg/logger"
)

type App struct {
	cfg       *config.Config
	logger    *zap.Logger
	db        *sqlx.DB
	redis     *redis.Client
	server    *server.Server
	hub       *websocket.Hub
	stream    *service.StreamService
	cancelCtx context.CancelFunc
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

	// Create WebSocket hub
	hub := websocket.NewHub(zapLogger)

	// Create Stream service
	streamService := service.NewStreamService(rdb, hub, zapLogger)

	// Create server
	srv := server.New(cfg, zapLogger, db, rdb, hub)

	return &App{
		cfg:    cfg,
		logger: zapLogger,
		db:     db,
		redis:  rdb,
		server: srv,
		hub:    hub,
		stream: streamService,
	}, nil
}

func (a *App) Run() error {
	// Start WebSocket hub
	go a.hub.Run()

	// Start Redis Stream consumer
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelCtx = cancel
	go func() {
		if err := a.stream.Consume(ctx); err != nil && err != context.Canceled {
			a.logger.Error("Stream consumer error", zap.Error(err))
		}
	}()

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

	// Cancel Stream context
	if a.cancelCtx != nil {
		a.cancelCtx()
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
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
