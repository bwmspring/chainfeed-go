package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

type Hub struct {
	clients    map[int64]map[*Client]bool
	broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	logger     *zap.Logger
}

type Message struct {
	UserID  int64       `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[int64]map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			clientCount := len(h.clients[client.UserID])
			totalClients := 0
			for _, clients := range h.clients {
				totalClients += len(clients)
			}
			h.mu.Unlock()
			
			h.logger.Info("websocket client connected",
				zap.Int64("user_id", client.UserID),
				zap.Int("user_connections", clientCount),
				zap.Int("total_connections", totalClients),
			)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			totalClients := 0
			for _, clients := range h.clients {
				totalClients += len(clients)
			}
			h.mu.Unlock()
			
			h.logger.Info("websocket client disconnected",
				zap.Int64("user_id", client.UserID),
				zap.Int("total_connections", totalClients),
			)

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[message.UserID]
			h.mu.RUnlock()

			data, err := json.Marshal(message)
			if err != nil {
				h.logger.Error("failed to marshal message", zap.Error(err))
				continue
			}

			for client := range clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Warn("websocket unexpected close",
					zap.Int64("user_id", c.UserID),
					zap.Error(err),
				)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
