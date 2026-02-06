package handler

import (
	"net/http"

	"github.com/bwmspring/chainfeed-go/internal/response"
	"github.com/bwmspring/chainfeed-go/internal/websocket"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub    *websocket.Hub
	logger *zap.Logger
}

func NewWebSocketHandler(hub *websocket.Hub, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger,
	}
}

// HandleWebSocket godoc
// @Summary WebSocket connection
// @Description Establish WebSocket connection for real-time feed updates
// @Tags websocket
// @Param token query string true "JWT token"
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade websocket connection",
			zap.Error(err),
			zap.Int64("user_id", userID.(int64)),
		)
		return
	}

	client := &websocket.Client{
		UserID: userID.(int64),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	h.hub.Register <- client

	h.logger.Info("websocket client registered",
		zap.Int64("user_id", userID.(int64)),
		zap.String("remote_addr", conn.RemoteAddr().String()),
	)

	go client.WritePump()
	go client.ReadPump()
}
