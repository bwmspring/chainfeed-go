package repository

import (
	"database/sql"
	"fmt"

	"github.com/bwmspring/chainfeed-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type TransactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (tx_hash, block_number, block_timestamp, from_address, to_address, 
			value, tx_type, token_address, token_id, token_symbol, token_decimals)
		VALUES (:tx_hash, :block_number, :block_timestamp, :from_address, :to_address, 
			:value, :tx_type, :token_address, :token_id, :token_symbol, :token_decimals)
		ON CONFLICT (tx_hash) DO UPDATE SET tx_hash = EXCLUDED.tx_hash
		RETURNING id, created_at`

	rows, err := r.db.NamedQuery(query, tx)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&tx.ID, &tx.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan transaction: %w", err)
		}
	}

	return nil
}

func (r *TransactionRepository) GetByHash(hash string) (*models.Transaction, error) {
	var tx models.Transaction
	query := `SELECT * FROM transactions WHERE tx_hash = $1`

	err := r.db.Get(&tx, query, hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &tx, nil
}

func (r *TransactionRepository) GetByAddress(address string, limit, offset int) ([]models.Transaction, error) {
	var txs []models.Transaction
	query := `
		SELECT * FROM transactions 
		WHERE from_address = $1 OR to_address = $1
		ORDER BY block_timestamp DESC
		LIMIT $2 OFFSET $3`

	err := r.db.Select(&txs, query, address, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions by address: %w", err)
	}

	return txs, nil
}
