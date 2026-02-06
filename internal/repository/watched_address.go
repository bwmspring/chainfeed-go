package repository

import (
	"context"

	"github.com/bwmspring/chainfeed-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type WatchedAddressRepository struct {
	db *sqlx.DB
}

func NewWatchedAddressRepository(db *sqlx.DB) *WatchedAddressRepository {
	return &WatchedAddressRepository{db: db}
}

func (r *WatchedAddressRepository) GetByUserID(ctx context.Context, userID int64) ([]models.WatchedAddress, error) {
	var addresses []models.WatchedAddress
	query := `
		SELECT id, user_id, address, label, ens_name, created_at
		FROM watched_addresses
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &addresses, query, userID)
	return addresses, err
}

func (r *WatchedAddressRepository) Create(ctx context.Context, addr *models.WatchedAddress) error {
	query := `
		INSERT INTO watched_addresses (user_id, address, label, ens_name, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query, addr.UserID, addr.Address, addr.Label, addr.ENSName).
		Scan(&addr.ID, &addr.CreatedAt)
}

func (r *WatchedAddressRepository) Delete(ctx context.Context, id, userID int64) error {
	query := `DELETE FROM watched_addresses WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return nil // 或返回自定义错误
	}

	return nil
}

func (r *WatchedAddressRepository) Exists(ctx context.Context, userID int64, address string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM watched_addresses WHERE user_id = $1 AND address = $2)`
	err := r.db.GetContext(ctx, &exists, query, userID, address)
	return exists, err
}

func (r *WatchedAddressRepository) UpdateENS(ctx context.Context, id int64, ensName string) error {
	query := `UPDATE watched_addresses SET ens_name = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, ensName, id)
	return err
}

func (r *WatchedAddressRepository) FindByAddress(address string) ([]models.WatchedAddress, error) {
	var addresses []models.WatchedAddress
	query := `
		SELECT id, user_id, address, label, ens_name, created_at
		FROM watched_addresses
		WHERE LOWER(address) = LOWER($1)
	`
	err := r.db.Select(&addresses, query, address)
	return addresses, err
}

func (r *WatchedAddressRepository) GetByUserAndAddress(ctx context.Context, userID int64, address string) (*models.WatchedAddress, error) {
	var addr models.WatchedAddress
	query := `SELECT * FROM watched_addresses WHERE user_id = $1 AND LOWER(address) = LOWER($2)`
	err := r.db.GetContext(ctx, &addr, query, userID, address)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}
