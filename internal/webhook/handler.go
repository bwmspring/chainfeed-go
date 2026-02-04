package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	
	"chainfeed-go/internal/config"
	"chainfeed-go/internal/parser"
	"chainfeed-go/internal/repository"
)

type Handler struct {
	cfg            *config.Config
	logger         *zap.Logger
	parser         *parser.TransactionParser
	batchProcessor *BatchProcessor
	
	// 性能监控
	requestCount   int64
	errorCount     int64
	processingTime time.Duration
	mutex          sync.RWMutex
}

func NewHandler(cfg *config.Config, logger *zap.Logger, db *sqlx.DB, redis *redis.Client) *Handler {
	txRepo := repository.NewTransactionRepository(db)
	batchProcessor := NewBatchProcessor(txRepo, redis, logger)
	
	return &Handler{
		cfg:            cfg,
		logger:         logger,
		parser:         parser.NewTransactionParser(),
		batchProcessor: batchProcessor,
	}
}

func (h *Handler) HandleAlchemy(c *gin.Context) {
	start := time.Now()
	
	// 增加请求计数
	h.mutex.Lock()
	h.requestCount++
	h.mutex.Unlock()
	
	// 快速验证签名
	if !h.verifySignature(c) {
		h.incrementErrorCount()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}
	
	// 异步处理请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.incrementErrorCount()
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	
	// 立即返回响应，提高 Webhook 稳定性
	c.JSON(http.StatusOK, gin.H{"status": "accepted"})
	
	// 异步处理数据
	go h.processWebhookAsync(body, start)
}

func (h *Handler) processWebhookAsync(body []byte, start time.Time) {
	var webhook parser.AlchemyWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		h.incrementErrorCount()
		h.logger.Error("Failed to unmarshal webhook", zap.Error(err))
		return
	}
	
	// 解析交易
	transactions, err := h.parser.ParseAlchemyWebhook(&webhook)
	if err != nil {
		h.incrementErrorCount()
		h.logger.Error("Failed to parse webhook", zap.Error(err))
		return
	}
	
	// 添加到批量处理器
	for _, tx := range transactions {
		h.batchProcessor.AddTransaction(tx)
	}
	
	// 更新性能指标
	duration := time.Since(start)
	h.mutex.Lock()
	h.processingTime += duration
	h.mutex.Unlock()
	
	h.logger.Info("Webhook processed", 
		zap.String("webhook_id", webhook.WebhookID),
		zap.Int("transactions", len(transactions)),
		zap.Duration("duration", duration))
}

func (h *Handler) incrementErrorCount() {
	h.mutex.Lock()
	h.errorCount++
	h.mutex.Unlock()
}

func (h *Handler) GetStats() map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	avgProcessingTime := time.Duration(0)
	if h.requestCount > 0 {
		avgProcessingTime = h.processingTime / time.Duration(h.requestCount)
	}
	
	return map[string]interface{}{
		"requests_total":        h.requestCount,
		"errors_total":          h.errorCount,
		"avg_processing_time":   avgProcessingTime.String(),
		"error_rate":           float64(h.errorCount) / float64(h.requestCount),
	}
}

func (h *Handler) verifySignature(c *gin.Context) bool {
	signature := c.GetHeader("X-Alchemy-Signature")
	if signature == "" {
		return false
	}
	
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return false
	}
	
	// Reset body for further reading
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	
	mac := hmac.New(sha256.New, []byte(h.cfg.Webhook.Secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
