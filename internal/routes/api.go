package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"chainfeed-go/internal/auth"
	"chainfeed-go/internal/config"
	"chainfeed-go/internal/handler"
	"chainfeed-go/internal/middleware"
	"chainfeed-go/internal/repository"
	"chainfeed-go/internal/service"
)

type APIRoutes struct {
	cfg                   *config.Config
	logger                *zap.Logger
	db                    *sqlx.DB
	authHandler           *handler.AuthHandler
	watchedAddressHandler *handler.WatchedAddressHandler
	jwtService            *auth.JWTService
}

func NewAPIRoutes(cfg *config.Config, logger *zap.Logger, db *sqlx.DB) *APIRoutes {
	// 初始化 repositories
	userRepo := repository.NewUserRepository(db)
	watchedAddrRepo := repository.NewWatchedAddressRepository(db)

	// 初始化 services
	web3Svc := auth.NewWeb3Service(cfg.Auth.SignMessage)
	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)
	
	// 初始化 ENS service（可选）
	var ensService *service.ENSService
	if cfg.Ethereum.RPCURL != "" {
		var err error
		ensService, err = service.NewENSService(cfg.Ethereum.RPCURL)
		if err != nil {
			logger.Warn("Failed to initialize ENS service", zap.Error(err))
		}
	}

	// 初始化 handlers
	authHandler := handler.NewAuthHandler(userRepo, web3Svc, jwtSvc, logger, cfg.Auth.NonceExpiry)
	watchedAddressHandler := handler.NewWatchedAddressHandler(watchedAddrRepo, ensService, logger)

	return &APIRoutes{
		cfg:                   cfg,
		logger:                logger,
		db:                    db,
		authHandler:           authHandler,
		watchedAddressHandler: watchedAddressHandler,
		jwtService:            jwtSvc,
	}
}

func (r *APIRoutes) RegisterRoutes(router *gin.RouterGroup) {
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/ping", r.ping)

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/nonce", r.authHandler.GetNonce)
			auth.POST("/verify", r.authHandler.VerifySignature)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(r.jwtService))
		{
			// User profile
			protected.GET("/profile", r.getUserProfile)

			// Watched addresses
			addresses := protected.Group("/addresses")
			{
				addresses.GET("", r.watchedAddressHandler.List)
				addresses.POST("", r.watchedAddressHandler.Add)
				addresses.DELETE("/:id", r.watchedAddressHandler.Remove)
			}

			// Feed routes (placeholder)
			feed := protected.Group("/feed")
			{
				feed.GET("", r.getFeed)
				feed.GET("/transactions/:hash", r.getTransaction)
			}
		}
	}
}

// Health check endpoint
func (r *APIRoutes) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// getUserProfile 获取用户信息
func (r *APIRoutes) getUserProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	walletAddress, _ := middleware.GetWalletAddress(c)

	c.JSON(http.StatusOK, gin.H{
		"user_id":        userID,
		"wallet_address": walletAddress,
	})
}

// Placeholder handlers - to be implemented later
func (r *APIRoutes) getFeed(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) getTransaction(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}
