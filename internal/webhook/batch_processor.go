package webhook

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/models"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/service"
	"github.com/bwmspring/chainfeed-go/internal/websocket"
)

// BatchProcessor 批量处理交易以提高吞吐量
type BatchProcessor struct {
	txRepo          *repository.TransactionRepository
	feedRepo        *repository.FeedRepository
	watchedAddrRepo *repository.WatchedAddressRepository
	redis           *redis.Client
	logger          *zap.Logger
	batchSize       int
	flushTime       time.Duration
	buffer          []*models.Transaction
	mutex           sync.Mutex
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

func NewBatchProcessor(
	txRepo *repository.TransactionRepository,
	feedRepo *repository.FeedRepository,
	watchedAddrRepo *repository.WatchedAddressRepository,
	redis *redis.Client,
	logger *zap.Logger,
) *BatchProcessor {
	bp := &BatchProcessor{
		txRepo:          txRepo,
		feedRepo:        feedRepo,
		watchedAddrRepo: watchedAddrRepo,
		redis:           redis,
		logger:          logger,
		batchSize:       100,             // 批量大小
		flushTime:       5 * time.Second, // 最大等待时间
		buffer:          make([]*models.Transaction, 0, 100),
		stopCh:          make(chan struct{}),
	}

	bp.start()
	return bp
}

func (bp *BatchProcessor) start() {
	bp.wg.Add(1)
	go bp.flushLoop()
}

func (bp *BatchProcessor) Stop() {
	close(bp.stopCh)
	bp.wg.Wait()
	bp.flushBuffer() // 处理剩余数据
}

func (bp *BatchProcessor) AddTransaction(tx *models.Transaction) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.buffer = append(bp.buffer, tx)

	if len(bp.buffer) >= bp.batchSize {
		bp.flushBuffer()
	}
}

func (bp *BatchProcessor) flushLoop() {
	defer bp.wg.Done()
	ticker := time.NewTicker(bp.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bp.mutex.Lock()
			if len(bp.buffer) > 0 {
				bp.flushBuffer()
			}
			bp.mutex.Unlock()
		case <-bp.stopCh:
			return
		}
	}
}

func (bp *BatchProcessor) flushBuffer() {
	if len(bp.buffer) == 0 {
		return
	}

	start := time.Now()
	ctx := context.Background()

	// 批量插入交易到数据库
	for _, tx := range bp.buffer {
		if err := bp.txRepo.Create(tx); err != nil {
			bp.logger.Error("Failed to store transaction",
				zap.String("tx_hash", tx.TxHash),
				zap.Error(err))
			continue
		}

		// 查找监控该地址的用户
		bp.createFeedItems(ctx, tx)
	}

	bp.logger.Info("Batch processed",
		zap.Int("count", len(bp.buffer)),
		zap.Duration("duration", time.Since(start)))

	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
}

func (bp *BatchProcessor) createFeedItems(ctx context.Context, tx *models.Transaction) {
	// 查找监控 from_address 或 to_address 的用户
	addresses := []string{tx.FromAddress}
	if tx.ToAddress != "" {
		addresses = append(addresses, tx.ToAddress)
	}

	for _, addr := range addresses {
		// 查询监控该地址的所有记录
		watchedAddrs, err := bp.watchedAddrRepo.FindByAddress(addr)
		if err != nil {
			bp.logger.Error("Failed to find watched addresses",
				zap.String("address", addr),
				zap.Error(err))
			continue
		}

		// 为每个监控该地址的用户创建 feed_item
		for _, wa := range watchedAddrs {
			feedItem := &models.FeedItem{
				UserID:           wa.UserID,
				TransactionID:    tx.ID,
				WatchedAddressID: wa.ID,
			}

			if err := bp.feedRepo.Create(feedItem); err != nil {
				bp.logger.Error("Failed to create feed item",
					zap.Int64("user_id", wa.UserID),
					zap.String("tx_hash", tx.TxHash),
					zap.Error(err))
				continue
			}

			// 通过 Redis Pub/Sub 推送消息
			bp.publishFeedUpdate(ctx, wa.UserID, tx, &wa)
		}
	}
}

func (bp *BatchProcessor) publishFeedUpdate(ctx context.Context, userID int64, tx *models.Transaction, wa *models.WatchedAddress) {
	msg := &websocket.Message{
		UserID: userID,
		Type:   "new_transaction",
		Payload: map[string]interface{}{
			"transaction":     tx,
			"watched_address": wa,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		bp.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	if err := bp.redis.Publish(ctx, service.FeedChannel, data).Err(); err != nil {
		bp.logger.Error("Failed to publish message", zap.Error(err))
	}
}
