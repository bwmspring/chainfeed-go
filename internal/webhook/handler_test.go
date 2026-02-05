package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"chainfeed-go/internal/config"
	"chainfeed-go/internal/parser"

	_ "github.com/mattn/go-sqlite3"
)

func TestHandleAlchemy(t *testing.T) {
	// Setup test database (in-memory SQLite for testing)
	db, err := sqlx.Connect("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE transactions (
			id INTEGER PRIMARY KEY,
			tx_hash TEXT UNIQUE,
			block_number INTEGER,
			block_timestamp DATETIME,
			from_address TEXT,
			to_address TEXT,
			value TEXT,
			tx_type TEXT,
			token_address TEXT,
			token_id TEXT,
			token_symbol TEXT,
			token_decimals INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	cfg := &config.Config{
		Webhook: config.WebhookConfig{
			Secret: "test-secret",
		},
	}

	logger := zap.NewNop()

	// Use sync handler for testing
	handler := NewSyncHandler(cfg, logger, db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/webhooks/alchemy", handler.HandleAlchemy)

	t.Run("Valid Alchemy Address Activity Webhook", func(t *testing.T) {
		// Real Alchemy Address Activity payload structure
		payload := parser.AlchemyWebhook{
			WebhookID: "wh_test123",
			ID:        "whevt_test456",
			CreatedAt: time.Now(),
			Type:      "ADDRESS_ACTIVITY",
			Event: parser.AlchemyWebhookEvent{
				Network: "ETH_MAINNET",
				Activity: []parser.AlchemyActivityEvent{
					{
						BlockNum:    "0x1234567",
						Hash:        "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
						FromAddress: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
						ToAddress:   "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
						Value:       1.5,
						Asset:       "ETH",
						Category:    "external",
						RawContract: parser.AlchemyRawContract{
							Value:    "1500000000000000000",
							Address:  "",
							Decimals: 18,
						},
					},
				},
			},
		}

		jsonData, err := json.Marshal(payload)
		require.NoError(t, err)

		// Create HMAC signature
		mac := hmac.New(sha256.New, []byte("test-secret"))
		mac.Write(jsonData)
		signature := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/webhooks/alchemy", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Alchemy-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Equal(t, float64(1), response["processed"])

		// Verify transaction was stored
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM transactions WHERE tx_hash = ?",
			"0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		payload := map[string]string{"test": "data"}
		jsonData, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/webhooks/alchemy", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Alchemy-Signature", "invalid-signature")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ERC20 Token Transfer", func(t *testing.T) {
		payload := parser.AlchemyWebhook{
			WebhookID: "wh_test123",
			ID:        "whevt_test789",
			CreatedAt: time.Now(),
			Type:      "ADDRESS_ACTIVITY",
			Event: parser.AlchemyWebhookEvent{
				Network: "ETH_MAINNET",
				Activity: []parser.AlchemyActivityEvent{
					{
						BlockNum:    "0x1234568",
						Hash:        "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
						FromAddress: "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
						ToAddress:   "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
						Value:       1000.0,
						Asset:       "USDC",
						Category:    "erc20",
						RawContract: parser.AlchemyRawContract{
							Value:    "1000000000", // 1000 USDC (6 decimals)
							Address:  "0xA0b86a33E6441b8C4505B8C4505B8C4505B8C450",
							Decimals: 6,
						},
					},
				},
			},
		}

		jsonData, err := json.Marshal(payload)
		require.NoError(t, err)

		mac := hmac.New(sha256.New, []byte("test-secret"))
		mac.Write(jsonData)
		signature := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/webhooks/alchemy", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Alchemy-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify ERC20 transaction was stored with token info
		var tokenSymbol string
		err = db.Get(&tokenSymbol, "SELECT token_symbol FROM transactions WHERE tx_hash = ?",
			"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
		require.NoError(t, err)
		assert.Equal(t, "USDC", tokenSymbol)
	})
}
