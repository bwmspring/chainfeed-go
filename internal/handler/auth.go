package handler

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/auth"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/response"
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
	Address string `json:"address" binding:"required"`
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
		response.BadRequest(c, err.Error())
		return
	}

	// 验证地址格式
	if !common.IsHexAddress(req.Address) {
		response.BadRequest(c, "invalid wallet address")
		return
	}

	address := common.HexToAddress(req.Address).Hex()
	ctx := c.Request.Context()

	// 生成新 nonce
	nonce, err := h.web3Svc.GenerateNonce()
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	// 原子操作：创建用户或更新 nonce
	_, err = h.userRepo.UpsertNonce(ctx, address, nonce)
	if err != nil {
		h.logger.Error("Failed to upsert user nonce", zap.Error(err), zap.String("address", address))
		response.InternalServerError(c, "internal server error")
		return
	}

	message := h.web3Svc.GetSignMessage(address, nonce)

	response.Success(c, GetNonceResponse{
		Nonce:   nonce,
		Message: message,
	})
}

type VerifySignatureRequest struct {
	Address   string `json:"address" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

type VerifySignatureResponse struct {
	Token string `json:"token"`
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
		response.BadRequest(c, err.Error())
		return
	}

	// 验证地址格式
	if !common.IsHexAddress(req.Address) {
		response.BadRequest(c, "invalid wallet address")
		return
	}

	address := common.HexToAddress(req.Address).Hex()
	ctx := c.Request.Context()

	// 获取用户
	user, err := h.userRepo.GetByWalletAddress(ctx, address)
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	if user == nil {
		response.Unauthorized(c, "user not found")
		return
	}

	// 验证签名
	if err := h.web3Svc.VerifySignature(address, req.Signature, user.Nonce); err != nil {
		h.logger.Warn("Signature verification failed",
			zap.Error(err),
			zap.String("address", address),
			zap.Int64("user_id", user.ID),
		)
		response.Unauthorized(c, "invalid signature")
		return
	}

	// 生成新 nonce（防止重放攻击）
	newNonce, err := h.web3Svc.GenerateNonce()
	if err != nil {
		h.logger.Error("Failed to generate nonce", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	if err := h.userRepo.UpdateNonce(ctx, user.ID, newNonce); err != nil {
		h.logger.Error("Failed to update nonce", zap.Error(err), zap.Int64("user_id", user.ID))
		response.InternalServerError(c, "internal server error")
		return
	}

	// 生成 JWT token
	token, err := h.jwtSvc.GenerateToken(user.ID, user.WalletAddress)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err), zap.Int64("user_id", user.ID))
		response.InternalServerError(c, "internal server error")
		return
	}

	h.logger.Info("User authenticated successfully",
		zap.String("address", address),
		zap.Int64("user_id", user.ID),
	)

	response.Success(c, VerifySignatureResponse{
		Token: token,
	})
}
