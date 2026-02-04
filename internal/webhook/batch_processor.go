package webhook

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"chainfeed-go/internal/models"
	"chainfeed-go/internal/repository"
)

// BatchProcessor 批量处理交易以提高吞吐量
type BatchProcessor struct {
	txRepo     *repository.TransactionRepository
	redis      *redis.Client
	logger     *zap.Logger
	batchSize  int
	flushTime  time.Duration
	buffer     []*models.Transaction
	mutex      sync.Mutex
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

func NewBatchProcessor(txRepo *repository.TransactionRepository, redis *redis.Client, logger *zap.Logger) *BatchProcessor {
	bp := &BatchProcessor{
		txRepo:    txRepo,
		redis:     redis,
		logger:    logger,
		batchSize: 100,           // 批量大小
		flushTime: 5 * time.Second, // 最大等待时间
		buffer:    make([]*models.Transaction, 0, 100),
		stopCh:    make(chan struct{}),
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
	
	// 批量插入数据库
	for _, tx := range bp.buffer {
		if err := bp.txRepo.Create(tx); err != nil {
			bp.logger.Error("Failed to store transaction", 
				zap.String("tx_hash", tx.TxHash), 
				zap.Error(err))
		}
	}
	
	// 发布到 Redis 用于实时推送
	ctx := context.Background()
	for _, tx := range bp.buffer {
		bp.redis.Publish(ctx, "transactions", tx.TxHash)
	}
	
	bp.logger.Info("Batch processed", 
		zap.Int("count", len(bp.buffer)),
		zap.Duration("duration", time.Since(start)))
	
	// 清空缓冲区
	bp.buffer = bp.buffer[:0]
}
