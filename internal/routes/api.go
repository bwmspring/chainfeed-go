package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"chainfeed-go/internal/config"
)

type APIRoutes struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *sqlx.DB
}

func NewAPIRoutes(cfg *config.Config, logger *zap.Logger, db *sqlx.DB) *APIRoutes {
	return &APIRoutes{
		cfg:    cfg,
		logger: logger,
		db:     db,
	}
}

func (r *APIRoutes) RegisterRoutes(router *gin.RouterGroup) {
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/ping", r.ping)
		
		// User routes (placeholder for future implementation)
		users := api.Group("/users")
		{
			users.POST("/auth", r.authUser)           // Web3 wallet auth
			users.GET("/profile", r.getUserProfile)   // Get user profile
		}
		
		// Watched addresses routes
		addresses := api.Group("/addresses")
		{
			addresses.GET("", r.getWatchedAddresses)     // Get user's watched addresses
			addresses.POST("", r.addWatchedAddress)      // Add watched address
			addresses.DELETE("/:id", r.removeWatchedAddress) // Remove watched address
		}
		
		// Feed routes
		feed := api.Group("/feed")
		{
			feed.GET("", r.getFeed)                      // Get user's transaction feed
			feed.GET("/transactions/:hash", r.getTransaction) // Get transaction details
		}
	}
}

// Health check endpoint
func (r *APIRoutes) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// Placeholder handlers - to be implemented in Phase 1.3
func (r *APIRoutes) authUser(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) getUserProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) getWatchedAddresses(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) addWatchedAddress(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) removeWatchedAddress(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) getFeed(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}

func (r *APIRoutes) getTransaction(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented yet"})
}
