package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/middleware"
	"github.com/bwmspring/chainfeed-go/internal/models"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/response"
	"github.com/bwmspring/chainfeed-go/internal/service"
	"github.com/bwmspring/chainfeed-go/internal/websocket"
)

type WatchedAddressHandler struct {
	repo           *repository.WatchedAddressRepository
	ensService     *service.ENSService
	alchemyService *service.AlchemyService
	txRepo         *repository.TransactionRepository
	feedRepo       *repository.FeedRepository
	redis          *redis.Client
	logger         *zap.Logger
}

func NewWatchedAddressHandler(
	repo *repository.WatchedAddressRepository,
	ensService *service.ENSService,
	alchemyService *service.AlchemyService,
	txRepo *repository.TransactionRepository,
	feedRepo *repository.FeedRepository,
	redis *redis.Client,
	logger *zap.Logger,
) *WatchedAddressHandler {
	return &WatchedAddressHandler{
		repo:           repo,
		ensService:     ensService,
		alchemyService: alchemyService,
		txRepo:         txRepo,
		feedRepo:       feedRepo,
		redis:          redis,
		logger:         logger,
	}
}

// List 获取用户的监控地址列表
// @Summary      获取监控地址列表
// @Description  获取当前用户的所有监控地址
// @Tags         监控地址
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} map[string][]models.WatchedAddress
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /addresses [get]
func (h *WatchedAddressHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	ctx := context.Background()
	addresses, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get watched addresses", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	response.Success(c, addresses)
}

type AddWatchedAddressRequest struct {
	Address string `json:"address" binding:"required"`
	Label   string `json:"label"`
}

// Add 添加监控地址
// @Summary      添加监控地址
// @Description  添加新的监控地址（支持以太坊地址或 ENS 域名）
// @Tags         监控地址
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body AddWatchedAddressRequest true "地址信息"
// @Success      201 {object} map[string]models.WatchedAddress
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /addresses [post]
func (h *WatchedAddressHandler) Add(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req AddWatchedAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := context.Background()

	// 处理 ENS 或地址
	var address, ensName string
	if common.IsHexAddress(req.Address) {
		// 是以太坊地址
		address = common.HexToAddress(req.Address).Hex()

		// 尝试反向解析 ENS
		if h.ensService != nil {
			if name, err := h.ensService.ReverseResolve(ctx, address); err == nil && name != "" {
				ensName = name
			}
		}
	} else {
		// 可能是 ENS 域名
		if h.ensService == nil {
			response.BadRequest(c, "ENS resolution not available")
			return
		}

		resolvedAddr, err := h.ensService.Resolve(ctx, req.Address)
		if err != nil {
			h.logger.Warn("Failed to resolve ENS", zap.Error(err), zap.String("ens", req.Address))
			response.BadRequest(c, "invalid address or ENS name")
			return
		}
		address = resolvedAddr
		ensName = req.Address
	}

	// 检查是否已存在
	exists, err := h.repo.Exists(ctx, userID, address)
	if err != nil {
		h.logger.Error("Failed to check address existence", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	if exists {
		response.Error(c, http.StatusConflict, 409, "address already watched")
		return
	}

	// 创建监控地址
	watchedAddr := &models.WatchedAddress{
		UserID:  userID,
		Address: address,
		Label:   req.Label,
		ENSName: ensName,
	}

	if err := h.repo.Create(ctx, watchedAddr); err != nil {
		h.logger.Error("Failed to create watched address", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	// 立即查询该地址的交易记录
	go h.fetchAndStoreTransactions(context.Background(), userID, watchedAddr)

	response.Success(c, watchedAddr)
}

// fetchAndStoreTransactions 查询并存储交易记录
func (h *WatchedAddressHandler) fetchAndStoreTransactions(ctx context.Context, userID int64, watchedAddr *models.WatchedAddress) {
	if h.alchemyService == nil {
		return
	}

	// 查询最近 20 笔交易
	transactions, err := h.alchemyService.GetAddressTransfersWithLimit(ctx, watchedAddr.Address, 20)
	if err != nil {
		h.logger.Error("Failed to fetch transactions from Alchemy",
			zap.String("address", watchedAddr.Address),
			zap.Error(err))
		return
	}

	h.logger.Info("Fetched transactions from Alchemy",
		zap.String("address", watchedAddr.Address),
		zap.Int("count", len(transactions)))

	// 存储交易并创建 feed
	for _, tx := range transactions {
		// 存储交易（数据库唯一索引会自动去重）
		if err := h.txRepo.Create(tx); err != nil {
			// 如果是重复交易，跳过
			h.logger.Debug("Transaction already exists or failed to store",
				zap.String("tx_hash", tx.TxHash),
				zap.Error(err))
			continue
		}

		// 创建 feed item
		feedItem := &models.FeedItem{
			UserID:           userID,
			TransactionID:    tx.ID,
			WatchedAddressID: watchedAddr.ID,
		}

		if err := h.feedRepo.Create(feedItem); err != nil {
			h.logger.Error("Failed to create feed item",
				zap.String("tx_hash", tx.TxHash),
				zap.Error(err))
			continue
		}

		// 推送到 Redis Stream（推送完整的 FeedItem）
		h.publishFeedUpdate(ctx, feedItem, tx, watchedAddr)
	}
}

func (h *WatchedAddressHandler) publishFeedUpdate(ctx context.Context, feedItem *models.FeedItem, tx *models.Transaction, wa *models.WatchedAddress) {
	// 构造前端期望的数据格式
	payload := map[string]interface{}{
		"id":              feedItem.ID,
		"created_at":      feedItem.CreatedAt,
		"transaction":     tx,
		"watched_address": wa,
	}

	msg := &websocket.Message{
		UserID:  feedItem.UserID,
		Type:    "new_transaction",
		Payload: payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	values := map[string]any{
		"user_id": feedItem.UserID,
		"type":    "new_transaction",
		"payload": string(data),
	}

	if err := h.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: service.FeedStream,
		Values: values,
	}).Err(); err != nil {
		h.logger.Error("Failed to publish to stream", zap.Error(err))
	}
}

// Remove 删除监控地址
// @Summary      删除监控地址
// @Description  删除指定的监控地址
// @Tags         监控地址
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "地址 ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /addresses/{id} [delete]
func (h *WatchedAddressHandler) Remove(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	ctx := context.Background()
	if err := h.repo.Delete(ctx, id, userID); err != nil {
		h.logger.Error("Failed to delete watched address", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	response.SuccessWithMessage(c, "address removed", nil)
}
