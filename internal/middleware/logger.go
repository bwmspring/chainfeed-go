package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 不需要记录日志的路径
var skipLogPaths = map[string]bool{
	"/health":      true,
	"/favicon.ico": true,
}

// RequestLogger 请求日志中间件 - 生产级别实现
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 跳过健康检查等路径
		if skipLogPaths[path] {
			c.Next()
			return
		}

		// WebSocket 特殊处理
		isWebSocket := strings.HasPrefix(path, "/ws")

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		// WebSocket 连接建立成功
		if isWebSocket && status == 101 {
			logger.Info("websocket connected",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("client_ip", clientIP),
				zap.Int("status", status),
				zap.Duration("latency", duration),
			)
			return
		}

		// 根据状态码决定日志级别
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Int("status", status),
			zap.Duration("latency", duration),
		}

		if status >= 500 {
			logger.Error("request completed", fields...)
		} else if status >= 400 {
			logger.Warn("request completed", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}
