package handler

import (
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/middleware"
	"github.com/bwmspring/chainfeed-go/internal/models"
	"github.com/bwmspring/chainfeed-go/internal/repository"
	"github.com/bwmspring/chainfeed-go/internal/response"
)

type TransactionHandler struct {
	txRepo          *repository.TransactionRepository
	watchedAddrRepo *repository.WatchedAddressRepository
	logger          *zap.Logger
}

func NewTransactionHandler(
	txRepo *repository.TransactionRepository,
	watchedAddrRepo *repository.WatchedAddressRepository,
	logger *zap.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		txRepo:          txRepo,
		watchedAddrRepo: watchedAddrRepo,
		logger:          logger,
	}
}

type TransactionListResponse struct {
	Transactions []TransactionWithAddress `json:"transactions"`
	TotalCount   int                      `json:"total_count"`
	Page         int                      `json:"page"`
	PageSize     int                      `json:"page_size"`
}

type TransactionWithAddress struct {
	ID             int64  `json:"id"`
	TxHash         string `json:"tx_hash"`
	BlockNumber    int64  `json:"block_number"`
	BlockTimestamp string `json:"block_timestamp"`
	FromAddress    string `json:"from_address"`
	ToAddress      string `json:"to_address"`
	Value          string `json:"value"`
	TxType         string `json:"tx_type"`
	TokenAddress   string `json:"token_address,omitempty"`
	TokenID        string `json:"token_id,omitempty"`
	TokenSymbol    string `json:"token_symbol,omitempty"`
	TokenDecimals  int    `json:"token_decimals,omitempty"`
	WatchedAddress struct {
		Address string `json:"address"`
		Label   string `json:"label"`
		ENSName string `json:"ens_name,omitempty"`
	} `json:"watched_address"`
}

// GetByAddress 获取指定地址的交易列表
// @Summary      获取地址交易
// @Description  获取指定监控地址的交易列表
// @Tags         交易
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        address path string true "以太坊地址"
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200 {object} TransactionListResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /addresses/{address}/transactions [get]
func (h *TransactionHandler) GetByAddress(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	address := c.Param("address")
	if !common.IsHexAddress(address) {
		response.BadRequest(c, "invalid address")
		return
	}
	address = common.HexToAddress(address).Hex()

	// 验证用户是否监控了该地址
	watchedAddrs, err := h.watchedAddrRepo.FindByAddress(address)
	if err != nil {
		h.logger.Error("Failed to find watched address", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	var watchedAddr *models.WatchedAddress
	for _, wa := range watchedAddrs {
		if wa.UserID == userID {
			watchedAddr = &wa
			break
		}
	}

	if watchedAddr == nil {
		response.NotFound(c, "address not watched")
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询交易
	txs, err := h.txRepo.GetByAddress(address, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to get transactions", zap.Error(err))
		response.InternalServerError(c, "internal server error")
		return
	}

	// 转换为响应格式
	result := make([]TransactionWithAddress, len(txs))
	for i, tx := range txs {
		result[i] = TransactionWithAddress{
			ID:             tx.ID,
			TxHash:         tx.TxHash,
			BlockNumber:    tx.BlockNumber,
			BlockTimestamp: tx.BlockTimestamp.Format("2006-01-02T15:04:05Z07:00"),
			FromAddress:    tx.FromAddress,
			ToAddress:      tx.ToAddress,
			Value:          tx.Value,
			TxType:         tx.TxType,
			TokenAddress:   tx.TokenAddress,
			TokenID:        tx.TokenID,
			TokenSymbol:    tx.TokenSymbol,
			TokenDecimals:  tx.TokenDecimals,
		}
		result[i].WatchedAddress.Address = watchedAddr.Address
		result[i].WatchedAddress.Label = watchedAddr.Label
		result[i].WatchedAddress.ENSName = watchedAddr.ENSName
	}

	response.Success(c, TransactionListResponse{
		Transactions: result,
		TotalCount:   len(result),
		Page:         page,
		PageSize:     pageSize,
	})
}
