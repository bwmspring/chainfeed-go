package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bwmspring/chainfeed-go/internal/websocket"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const FeedChannel = "feed:updates"

type PubSubService struct {
	redis  *redis.Client
	hub    *websocket.Hub
	logger *zap.Logger
}

func NewPubSubService(redis *redis.Client, hub *websocket.Hub, logger *zap.Logger) *PubSubService {
	return &PubSubService{
		redis:  redis,
		hub:    hub,
		logger: logger,
	}
}

func (s *PubSubService) Subscribe(ctx context.Context) error {
	pubsub := s.redis.Subscribe(ctx, FeedChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	s.logger.Info("subscribed to Redis channel", zap.String("channel", FeedChannel))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			var message websocket.Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				s.logger.Error("failed to unmarshal message", zap.Error(err))
				continue
			}

			s.hub.Broadcast(&message)
		}
	}
}

func (s *PubSubService) Publish(ctx context.Context, msg *websocket.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := s.redis.Publish(ctx, FeedChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
