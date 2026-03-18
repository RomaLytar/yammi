package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/romanlovesweed/yammi/services/user/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO profiles (id, email, name, avatar_url, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.AvatarURL, user.Bio, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrEmailExists
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, name, avatar_url, bio, created_at, updated_at FROM profiles WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE profiles SET name = $1, avatar_url = $2, bio = $3, updated_at = $4 WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query,
		user.Name, user.AvatarURL, user.Bio, user.UpdatedAt, user.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
