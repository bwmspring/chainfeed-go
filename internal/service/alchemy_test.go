package service

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestAlchemyService_GetAddressTransfers(t *testing.T) {
	// 从环境变量或直接填写 API key
	apiKey := "xxx_xxxx_" // 替换为真实的 API key

	logger, _ := zap.NewDevelopment()
	service := NewAlchemyService(apiKey, logger)

	// Vitalik.eth
	address := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"

	ctx := context.Background()
	transactions, err := service.GetAddressTransfers(ctx, address)

	if err != nil {
		t.Fatalf("Failed to get transfers: %v", err)
	}

	t.Logf("Found %d transactions", len(transactions))

	for i, tx := range transactions {
		if i >= 1 { // 只打印前 1 条
			break
		}
		t.Logf("TX %d: %s | From: %s | To: %s | Value: %s",
			i+1, tx.TxHash, tx.FromAddress[:10], tx.ToAddress[:10], tx.Value)
	}
}
