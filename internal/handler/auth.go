package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/auth"
	"github.com/bwmspring/chainfeed-go/internal/models"
	"github.com/bwmspring/chainfeed-go/internal/repository"
)

type AuthHandler struct {
	userRepo    *repository.UserRepository
	web3Svc     *auth.Web3Service
	jwtSvc      *auth.JWTService
	logger      *zap.Logger
	nonceExpiry time.Duration
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	web3Svc *auth.Web3Service,
	jwtSvc *auth.JWTService,
	logger *zap.Logger,
	nonceExpiry time.Duration,
) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		web3Svc:     web3Svc,
		jwtSvc:      jwtSvc,
		logger:      logger,
		nonceExpiry: nonceExpiry,
	}
}

type GetNonceRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type GetNonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

// GetNonce 获取签名用的 nonce
// @Summary      获取签名 Nonce
// @Description  获取用于 MetaMask 签名的 nonce 和消息
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body GetNonceRequest true "钱包地址"
// @Success      200 {object} GetNonceResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/nonce [post]
func (h *AuthHandler) GetNonce(c *gin.Context) {
	var req GetNonceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证地址格式
	if !common.IsHexAddress(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet address"})
		return
	}

	address := common.HexToAddress(req.WalletAddress).Hex()
	ctx := context.Background()

	// 查找或创建用户
	user, err := h.userRepo.GetByWalletAddress(ctx, address)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 生成新 nonce
	nonce, err := h.web3Svc.GenerateNonce()
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		// 创建新用户
		user = &models.User{
			WalletAddress: address,
			Nonce:         nonce,
		}
		if err := h.userRepo.Create(ctx, user); err != nil {
			h.logger.Error("Failed to create user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	} else {
		// 更新 nonce
		if err := h.userRepo.UpdateNonce(ctx, user.ID, nonce); err != nil {
			h.logger.Error("Failed to update nonce", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	message := h.web3Svc.GetSignMessage(address, nonce)

	c.JSON(http.StatusOK, GetNonceResponse{
		Nonce:   nonce,
		Message: message,
	})
}

type VerifySignatureRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature"      binding:"required"`
}

type VerifySignatureResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// VerifySignature 验证签名并返回 JWT token
// @Summary      验证签名
// @Description  验证 MetaMask 签名并返回 JWT token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request body VerifySignatureRequest true "签名信息"
// @Success      200 {object} VerifySignatureResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /auth/verify [post]
func (h *AuthHandler) VerifySignature(c *gin.Context) {
	var req VerifySignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address := common.HexToAddress(req.WalletAddress).Hex()
	ctx := context.Background()

	// 获取用户
	user, err := h.userRepo.GetByWalletAddress(ctx, address)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// 验证签名
	if err := h.web3Svc.VerifySignature(address, req.Signature, user.Nonce); err != nil {
		h.logger.Warn("Signature verification failed", zap.Error(err), zap.String("address", address))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// 生成新 nonce（防止重放攻击）
	newNonce, err := h.web3Svc.GenerateNonce()
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.userRepo.UpdateNonce(ctx, user.ID, newNonce); err != nil {
		h.logger.Error("Failed to update nonce", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// 生成 JWT token
	token, err := h.jwtSvc.GenerateToken(user.ID, user.WalletAddress)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, VerifySignatureResponse{
		Token: token,
		User:  user,
	})
}
