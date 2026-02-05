package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/auth"
	"github.com/bwmspring/chainfeed-go/internal/config"
	"github.com/bwmspring/chainfeed-go/internal/handler"
	"github.com/bwmspring/chainfeed-go/internal/middleware"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/service"
	"github.com/bwmspring/chainfeed-go/internal/websocket"
)

type APIRoutes struct {
	cfg                   *config.Config
	logger                *zap.Logger
	db                    *sqlx.DB
	authHandler           *handler.AuthHandler
	watchedAddressHandler *handler.WatchedAddressHandler
	feedHandler           *handler.FeedHandler
	wsHandler             *handler.WebSocketHandler
	jwtService            *auth.JWTService
}

func NewAPIRoutes(cfg *config.Config, logger *zap.Logger, db *sqlx.DB, hub *websocket.Hub) *APIRoutes {
	// 初始化 repositories
	userRepo := repository.NewUserRepository(db)
	watchedAddrRepo := repository.NewWatchedAddressRepository(db)
	feedRepo := repository.NewFeedRepository(db)

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
	feedHandler := handler.NewFeedHandler(feedRepo)
	wsHandler := handler.NewWebSocketHandler(hub, logger)

	return &APIRoutes{
		cfg:                   cfg,
		logger:                logger,
		db:                    db,
		authHandler:           authHandler,
		watchedAddressHandler: watchedAddressHandler,
		feedHandler:           feedHandler,
		wsHandler:             wsHandler,
		jwtService:            jwtSvc,
	}
}

func (r *APIRoutes) RegisterRoutes(router *gin.RouterGroup) {
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket endpoint (with auth middleware)
	router.GET("/ws", middleware.AuthMiddleware(r.jwtService), r.wsHandler.HandleWebSocket)

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

			// Feed routes
			feed := protected.Group("/feed")
			{
				feed.GET("", r.feedHandler.GetFeed)
			}
		}
	}
}

// Health check endpoint
// @Summary      健康检查
// @Description  检查服务是否正常运行
// @Tags         系统
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /ping [get]
func (r *APIRoutes) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// getUserProfile 获取用户信息
// @Summary      获取用户信息
// @Description  获取当前登录用户的基本信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string
// @Router       /profile [get]
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
