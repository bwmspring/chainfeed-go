package models

import (
	"time"
)

type User struct {
	ID            int64     `db:"id" json:"id"`
	WalletAddress string    `db:"wallet_address" json:"wallet_address"`
	Nonce         string    `db:"nonce" json:"nonce"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type WatchedAddress struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Address   string    `db:"address" json:"address"`
	Label     string    `db:"label" json:"label"`
	ENSName   string    `db:"ens_name" json:"ens_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Transaction struct {
	ID             int64     `db:"id" json:"id"`
	TxHash         string    `db:"tx_hash" json:"tx_hash"`
	BlockNumber    int64     `db:"block_number" json:"block_number"`
	BlockTimestamp time.Time `db:"block_timestamp" json:"block_timestamp"`
	FromAddress    string    `db:"from_address" json:"from_address"`
	ToAddress      string    `db:"to_address" json:"to_address"`
	Value          string    `db:"value" json:"value"`
	TxType         string    `db:"tx_type" json:"tx_type"`
	TokenAddress   string    `db:"token_address" json:"token_address"`
	TokenID        string    `db:"token_id" json:"token_id"`
	TokenSymbol    string    `db:"token_symbol" json:"token_symbol"`
	TokenDecimals  int       `db:"token_decimals" json:"token_decimals"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type FeedItem struct {
	ID                int64     `db:"id" json:"id"`
	UserID            int64     `db:"user_id" json:"user_id"`
	TransactionID     int64     `db:"transaction_id" json:"transaction_id"`
	WatchedAddressID  int64     `db:"watched_address_id" json:"watched_address_id"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}
