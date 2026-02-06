package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmspring/chainfeed-go/internal/websocket"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	FeedStream      = "feed:stream"
	FeedConsumerGrp = "feed:consumers"
	ConsumerName    = "consumer-1"
	MaxRetries      = 3
	ClaimMinIdle    = 30 * time.Second
)

type StreamService struct {
	redis  *redis.Client
	hub    *websocket.Hub
	logger *zap.Logger
}

func NewStreamService(redis *redis.Client, hub *websocket.Hub, logger *zap.Logger) *StreamService {
	return &StreamService{
		redis:  redis,
		hub:    hub,
		logger: logger,
	}
}

// InitConsumerGroup 初始化消费者组
func (s *StreamService) InitConsumerGroup(ctx context.Context) error {
	err := s.redis.XGroupCreateMkStream(ctx, FeedStream, FeedConsumerGrp, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	s.logger.Info("consumer group initialized", zap.String("stream", FeedStream))
	return nil
}

// Publish 发布消息到 Stream
func (s *StreamService) Publish(ctx context.Context, msg *websocket.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	values := map[string]interface{}{
		"user_id": msg.UserID,
		"type":    msg.Type,
		"payload": string(data),
	}

	if err := s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: FeedStream,
		Values: values,
	}).Err(); err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	return nil
}

// Consume 消费消息
func (s *StreamService) Consume(ctx context.Context) error {
	if err := s.InitConsumerGroup(ctx); err != nil {
		return err
	}

	s.logger.Info("started consuming from stream", zap.String("stream", FeedStream))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 读取新消息
			streams, err := s.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    FeedConsumerGrp,
				Consumer: ConsumerName,
				Streams:  []string{FeedStream, ">"},
				Count:    10,
				Block:    time.Second,
			}).Result()

			if err != nil && err != redis.Nil {
				s.logger.Error("failed to read from stream", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := s.processMessage(ctx, message); err != nil {
						s.logger.Error("failed to process message",
							zap.String("message_id", message.ID),
							zap.Error(err))
					}
				}
			}

			// 处理待处理消息（超时未确认的）
			s.claimPendingMessages(ctx)
		}
	}
}

func (s *StreamService) processMessage(ctx context.Context, msg redis.XMessage) error {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		return fmt.Errorf("invalid payload format")
	}

	var message websocket.Message
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		// 无法解析的消息直接 ACK 丢弃
		s.redis.XAck(ctx, FeedStream, FeedConsumerGrp, msg.ID)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// 推送到 WebSocket
	s.hub.Broadcast(&message)

	// 确认消息
	if err := s.redis.XAck(ctx, FeedStream, FeedConsumerGrp, msg.ID).Err(); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	return nil
}

// claimPendingMessages 认领超时的待处理消息
func (s *StreamService) claimPendingMessages(ctx context.Context) {
	pending, err := s.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: FeedStream,
		Group:  FeedConsumerGrp,
		Start:  "-",
		End:    "+",
		Count:  10,
	}).Result()

	if err != nil {
		return
	}

	for _, msg := range pending {
		// 超过最大重试次数，直接 ACK 丢弃
		if msg.RetryCount >= MaxRetries {
			s.logger.Warn("message exceeded max retries, discarding",
				zap.String("message_id", msg.ID),
				zap.Int64("retry_count", msg.RetryCount))
			s.redis.XAck(ctx, FeedStream, FeedConsumerGrp, msg.ID)
			continue
		}

		// 认领超时消息
		if msg.Idle >= ClaimMinIdle {
			claimed, err := s.redis.XClaim(ctx, &redis.XClaimArgs{
				Stream:   FeedStream,
				Group:    FeedConsumerGrp,
				Consumer: ConsumerName,
				MinIdle:  ClaimMinIdle,
				Messages: []string{msg.ID},
			}).Result()

			if err != nil {
				continue
			}

			for _, claimedMsg := range claimed {
				if err := s.processMessage(ctx, claimedMsg); err != nil {
					s.logger.Error("failed to process claimed message",
						zap.String("message_id", claimedMsg.ID),
						zap.Error(err))
				}
			}
		}
	}
}
