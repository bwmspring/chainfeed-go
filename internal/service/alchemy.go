package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/bwmspring/chainfeed-go/internal/models"
)

type AlchemyService struct {
	apiKey string
	apiURL string
	logger *zap.Logger
	client *http.Client
}

func NewAlchemyService(apiKey string, logger *zap.Logger) *AlchemyService {
	return &AlchemyService{
		apiKey: apiKey,
		apiURL: fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", apiKey),
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type AlchemyTransfer struct {
	BlockNum string           `json:"blockNum"`
	Hash     string           `json:"hash"`
	From     string           `json:"from"`
	To       string           `json:"to"`
	Value    float64          `json:"value"`
	Asset    string           `json:"asset"`
	Category string           `json:"category"`
	Metadata AlchemyMetadata  `json:"metadata"`
}

type AlchemyMetadata struct {
	BlockTimestamp string `json:"blockTimestamp"`
}

type AlchemyResponse struct {
	Result struct {
		Transfers []AlchemyTransfer `json:"transfers"`
	} `json:"result"`
}

// GetAddressTransfers 获取地址的 ETH 转账记录（发送+接收）
func (s *AlchemyService) GetAddressTransfers(ctx context.Context, address string) ([]*models.Transaction, error) {
	return s.GetAddressTransfersWithLimit(ctx, address, 0)
}

func (s *AlchemyService) GetAddressTransfersWithLimit(ctx context.Context, address string, limit int) ([]*models.Transaction, error) {
	// 查询从创世区块开始
	fromBlock := "0x0"
	
	// 获取发送的交易
	outgoing, err := s.getTransfers(ctx, address, "", fromBlock)
	if err != nil {
		return nil, err
	}
	
	// 获取接收的交易
	incoming, err := s.getTransfers(ctx, "", address, fromBlock)
	if err != nil {
		return nil, err
	}
	
	// 合并并去重
	txMap := make(map[string]*models.Transaction)
	for _, tx := range append(outgoing, incoming...) {
		txMap[tx.TxHash] = tx
	}
	
	result := make([]*models.Transaction, 0, len(txMap))
	for _, tx := range txMap {
		result = append(result, tx)
	}
	
	// 按时间戳降序排序（最新的在前）
	sort.Slice(result, func(i, j int) bool {
		return result[i].BlockTimestamp.After(result[j].BlockTimestamp)
	})
	
	// 限制返回数量（取最新的 N 条）
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	
	s.logger.Info("Merged transactions", zap.Int("total", len(result)), zap.Int("limit", limit))
	return result, nil
}

func (s *AlchemyService) getTransfers(ctx context.Context, fromAddr, toAddr, fromBlock string) ([]*models.Transaction, error) {
	params := map[string]interface{}{
		"fromBlock":        fromBlock,
		"toBlock":          "latest",
		"category":         []string{"external"},
		"withMetadata":     true,
		"excludeZeroValue": true,
		"maxCount":         "0x64", // 100 条
	}
	
	if fromAddr != "" {
		params["fromAddress"] = fromAddr
	}
	if toAddr != "" {
		params["toAddress"] = toAddr
	}
	
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "alchemy_getAssetTransfers",
		"params":  []interface{}{params},
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Alchemy API error", 
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)))
		return nil, fmt.Errorf("alchemy API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	s.logger.Debug("Alchemy API response", zap.String("body", string(body)))
	
	var alchemyResp AlchemyResponse
	if err := json.Unmarshal(body, &alchemyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	s.logger.Info("Alchemy transfers fetched",
		zap.String("from", fromAddr),
		zap.String("to", toAddr),
		zap.Int("count", len(alchemyResp.Result.Transfers)))
	
	transactions := make([]*models.Transaction, 0, len(alchemyResp.Result.Transfers))
	for _, transfer := range alchemyResp.Result.Transfers {
		blockNum, err := strconv.ParseInt(strings.TrimPrefix(transfer.BlockNum, "0x"), 16, 64)
		if err != nil {
			s.logger.Warn("Failed to parse block number", zap.String("blockNum", transfer.BlockNum))
			continue
		}
		
		// 解析区块时间戳
		blockTime, err := time.Parse(time.RFC3339, transfer.Metadata.BlockTimestamp)
		if err != nil {
			s.logger.Warn("Failed to parse block timestamp", 
				zap.String("timestamp", transfer.Metadata.BlockTimestamp),
				zap.Error(err))
			blockTime = time.Now()
		}
		
		valueWei := fmt.Sprintf("%.0f", transfer.Value*1e18)
		
		tx := &models.Transaction{
			TxHash:         transfer.Hash,
			BlockNumber:    blockNum,
			BlockTimestamp: blockTime,
			FromAddress:    transfer.From,
			ToAddress:      transfer.To,
			Value:          valueWei,
			TxType:         "ETH",
		}
		
		transactions = append(transactions, tx)
	}
	
	return transactions, nil
}
