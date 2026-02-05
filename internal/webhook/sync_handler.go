package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"chainfeed-go/internal/config"
	"chainfeed-go/internal/parser"
	"chainfeed-go/internal/repository"
)

// SyncHandler 用于测试的同步处理器
type SyncHandler struct {
	cfg    *config.Config
	logger *zap.Logger
	parser *parser.TransactionParser
	txRepo *repository.TransactionRepository
}

func NewSyncHandler(cfg *config.Config, logger *zap.Logger, db *sqlx.DB) *SyncHandler {
	return &SyncHandler{
		cfg:    cfg,
		logger: logger,
		parser: parser.NewTransactionParser(),
		txRepo: repository.NewTransactionRepository(db),
	}
}

func (h *SyncHandler) HandleAlchemy(c *gin.Context) {
	// Verify webhook signature
	if !h.verifySignature(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Parse request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var webhook parser.AlchemyWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		h.logger.Error("Failed to unmarshal webhook", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	// Parse transactions
	transactions, err := h.parser.ParseAlchemyWebhook(&webhook)
	if err != nil {
		h.logger.Error("Failed to parse webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse failed"})
		return
	}

	// Store transactions synchronously
	for _, tx := range transactions {
		if err := h.txRepo.Create(tx); err != nil {
			h.logger.Error("Failed to store transaction",
				zap.String("tx_hash", tx.TxHash),
				zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"processed": len(transactions),
	})
}

func (h *SyncHandler) verifySignature(c *gin.Context) bool {
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
