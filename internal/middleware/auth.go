package middleware

import (
	"net/http"
	"strings"

	"github.com/bwmspring/chainfeed-go/internal/auth"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 优先从 Header 获取（REST API）
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}

		// 如果 Header 没有，从 Query 获取（WebSocket）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("wallet_address", claims.WalletAddress)
		c.Next()
	}
}

// GetUserID 从上下文获取用户 ID
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	return id, ok
}

// GetWalletAddress 从上下文获取钱包地址
func GetWalletAddress(c *gin.Context) (string, bool) {
	address, exists := c.Get("wallet_address")
	if !exists {
		return "", false
	}
	addr, ok := address.(string)
	return addr, ok
}
