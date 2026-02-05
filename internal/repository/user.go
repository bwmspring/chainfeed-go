package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bwmspring/chainfeed-go/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	var user models.User
	query := `SELECT id, wallet_address, nonce, created_at, updated_at FROM users WHERE wallet_address = $1`
	err := r.db.GetContext(ctx, &user, query, walletAddress)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (wallet_address, nonce, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.WalletAddress, user.Nonce).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) UpdateNonce(ctx context.Context, userID int64, nonce string) error {
	query := `UPDATE users SET nonce = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, nonce, userID)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User
	query := `SELECT id, wallet_address, nonce, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
