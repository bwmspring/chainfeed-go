package repository

import (
	"fmt"

	"github.com/bwmspring/chainfeed-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

type FeedItemDetail struct {
	models.FeedItem
	Transaction    models.Transaction    `db:"transaction"`
	WatchedAddress models.WatchedAddress `db:"watched_address"`
}

func (r *FeedRepository) GetUserFeed(userID int64, limit, offset int) ([]FeedItemDetail, error) {
	query := `
		SELECT 
			fi.id, fi.user_id, fi.transaction_id, fi.watched_address_id, fi.created_at,
			t.id as "transaction.id", t.tx_hash as "transaction.tx_hash", 
			t.block_number as "transaction.block_number", t.block_timestamp as "transaction.block_timestamp",
			t.from_address as "transaction.from_address", t.to_address as "transaction.to_address",
			t.value as "transaction.value", t.tx_type as "transaction.tx_type",
			t.token_address as "transaction.token_address", t.token_id as "transaction.token_id",
			t.token_symbol as "transaction.token_symbol", t.token_decimals as "transaction.token_decimals",
			wa.id as "watched_address.id", wa.address as "watched_address.address",
			wa.label as "watched_address.label", wa.ens_name as "watched_address.ens_name"
		FROM feed_items fi
		JOIN transactions t ON fi.transaction_id = t.id
		JOIN watched_addresses wa ON fi.watched_address_id = wa.id
		WHERE fi.user_id = $1
		ORDER BY fi.created_at DESC
		LIMIT $2 OFFSET $3`

	var items []FeedItemDetail
	err := r.db.Select(&items, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user feed: %w", err)
	}

	return items, nil
}

func (r *FeedRepository) Create(item *models.FeedItem) error {
	query := `
		INSERT INTO feed_items (user_id, transaction_id, watched_address_id)
		VALUES (:user_id, :transaction_id, :watched_address_id)
		ON CONFLICT (user_id, transaction_id) DO NOTHING
		RETURNING id, created_at`

	rows, err := r.db.NamedQuery(query, item)
	if err != nil {
		return fmt.Errorf("failed to create feed item: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&item.ID, &item.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan feed item: %w", err)
		}
	}

	return nil
}
