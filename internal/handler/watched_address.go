package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/middleware"
	"github.com/bwmspring/chainfeed-go/internal/models"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/service"
)

type WatchedAddressHandler struct {
	repo       *repository.WatchedAddressRepository
	ensService *service.ENSService
	logger     *zap.Logger
}

func NewWatchedAddressHandler(
	repo *repository.WatchedAddressRepository,
	ensService *service.ENSService,
	logger *zap.Logger,
) *WatchedAddressHandler {
	return &WatchedAddressHandler{
		repo:       repo,
		ensService: ensService,
		logger:     logger,
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx := context.Background()
	addresses, err := h.repo.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get watched addresses", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": addresses})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req AddWatchedAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "ENS resolution not available"})
			return
		}

		resolvedAddr, err := h.ensService.Resolve(ctx, req.Address)
		if err != nil {
			h.logger.Warn("Failed to resolve ENS", zap.Error(err), zap.String("ens", req.Address))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address or ENS name"})
			return
		}
		address = resolvedAddr
		ensName = req.Address
	}

	// 检查是否已存在
	exists, err := h.repo.Exists(ctx, userID, address)
	if err != nil {
		h.logger.Error("Failed to check address existence", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "address already watched"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": watchedAddr})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := context.Background()
	if err := h.repo.Delete(ctx, id, userID); err != nil {
		h.logger.Error("Failed to delete watched address", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "address removed"})
}
