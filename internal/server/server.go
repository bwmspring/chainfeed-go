package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"chainfeed-go/internal/config"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sqlx.DB
	redis  *redis.Client
	router *gin.Engine
	http   *http.Server
}

func New(cfg *config.Config, logger *zap.Logger, db *sqlx.DB, rdb *redis.Client) *Server {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	s := &Server{
		cfg:    cfg,
		logger: logger,
		db:     db,
		redis:  rdb,
		router: router,
	}

	s.setupRoutes()

	s.http = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthCheck)

	api := s.router.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	ctx := context.Background()

	// Check database
	if err := s.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	// Check redis
	if err := s.redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "redis connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.Int("port", s.cfg.Server.Port))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")
	return s.http.Shutdown(ctx)
}
