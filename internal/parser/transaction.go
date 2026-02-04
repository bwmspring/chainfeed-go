package parser

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"chainfeed-go/internal/models"
)

type AlchemyWebhook struct {
	WebhookID string                 `json:"webhookId"`
	ID        string                 `json:"id"`
	CreatedAt time.Time              `json:"createdAt"`
	Type      string                 `json:"type"`
	Event     AlchemyWebhookEvent    `json:"event"`
}

type AlchemyWebhookEvent struct {
	Network     string                    `json:"network"`
	Activity    []AlchemyActivityEvent    `json:"activity"`
}

type AlchemyActivityEvent struct {
	BlockNum         string                 `json:"blockNum"`
	Hash             string                 `json:"hash"`
	FromAddress      string                 `json:"fromAddress"`
	ToAddress        string                 `json:"toAddress"`
	Value            float64                `json:"value"`
	Asset            string                 `json:"asset"`
	Category         string                 `json:"category"`
	RawContract      AlchemyRawContract     `json:"rawContract"`
	TypeTraceAddress string                 `json:"typeTraceAddress"`
	Log              AlchemyLog             `json:"log"`
}

type AlchemyRawContract struct {
	Value    string `json:"value"`
	Address  string `json:"address"`
	Decimals int    `json:"decimals"`
}

type AlchemyLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
	BlockHash        string   `json:"blockHash"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
}

type TransactionParser struct{}

func NewTransactionParser() *TransactionParser {
	return &TransactionParser{}
}

func (p *TransactionParser) ParseAlchemyWebhook(webhook *AlchemyWebhook) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	
	for _, activity := range webhook.Event.Activity {
		tx, err := p.parseActivity(&activity)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activity: %w", err)
		}
		if tx != nil {
			transactions = append(transactions, tx)
		}
	}
	
	return transactions, nil
}

func (p *TransactionParser) parseActivity(activity *AlchemyActivityEvent) (*models.Transaction, error) {
	blockNum, err := strconv.ParseInt(strings.TrimPrefix(activity.BlockNum, "0x"), 16, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block number: %w", err)
	}
	
	tx := &models.Transaction{
		TxHash:         activity.Hash,
		BlockNumber:    blockNum,
		BlockTimestamp: time.Now(), // TODO: Get actual block timestamp
		FromAddress:    strings.ToLower(activity.FromAddress),
		ToAddress:      strings.ToLower(activity.ToAddress),
		Value:          p.formatValue(activity.Value),
		TxType:         p.getTxType(activity.Category),
	}
	
	// Handle token transfers
	if activity.Category == "erc20" || activity.Category == "erc721" {
		tx.TokenAddress = strings.ToLower(activity.RawContract.Address)
		tx.TokenDecimals = activity.RawContract.Decimals
		tx.TokenSymbol = activity.Asset
		
		if activity.Category == "erc721" {
			tx.TokenID = activity.RawContract.Value
		}
	}
	
	return tx, nil
}

func (p *TransactionParser) formatValue(value float64) string {
	// Convert to wei (18 decimals)
	wei := big.NewFloat(value)
	wei.Mul(wei, big.NewFloat(1e18))
	
	result, _ := wei.Int(nil)
	return result.String()
}

func (p *TransactionParser) getTxType(category string) string {
	switch category {
	case "external":
		return "ETH"
	case "erc20":
		return "ERC20"
	case "erc721":
		return "ERC721"
	default:
		return "UNKNOWN"
	}
}
