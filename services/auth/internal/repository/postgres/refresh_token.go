package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

type RefreshTokenRepo struct {
	db *sql.DB
}

func NewRefreshTokenRepo(db *sql.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.Revoked, token.CreatedAt)
	return err
}

func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, revoked, created_at FROM refresh_tokens WHERE token = $1`

	rt := &domain.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.Revoked, &rt.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepo) RevokeByToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

func (r *RefreshTokenRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
