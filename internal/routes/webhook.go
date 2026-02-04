package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"chainfeed-go/internal/config"
	"chainfeed-go/internal/webhook"
)

type WebhookRoutes struct {
	handler *webhook.Handler
}

func NewWebhookRoutes(cfg *config.Config, logger *zap.Logger, db *sqlx.DB, redis *redis.Client) *WebhookRoutes {
	return &WebhookRoutes{
		handler: webhook.NewHandler(cfg, logger, db, redis),
	}
}

func (r *WebhookRoutes) RegisterRoutes(router *gin.RouterGroup) {
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/alchemy", r.handler.HandleAlchemy)
	}
}

func (r *WebhookRoutes) GetHandler() *webhook.Handler {
	return r.handler
}

type MonitoringRoutes struct {
	webhookHandler *webhook.Handler
}

func NewMonitoringRoutes(webhookHandler *webhook.Handler) *MonitoringRoutes {
	return &MonitoringRoutes{
		webhookHandler: webhookHandler,
	}
}

func (r *MonitoringRoutes) RegisterRoutes(router *gin.RouterGroup) {
	monitoring := router.Group("/monitoring")
	{
		monitoring.GET("/stats", func(c *gin.Context) {
			c.JSON(http.StatusOK, r.webhookHandler.GetStats())
		})
	}
}
