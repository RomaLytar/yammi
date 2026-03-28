package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UserLabelRepository struct {
	db *sql.DB
}

func NewUserLabelRepository(db *sql.DB) *UserLabelRepository {
	return &UserLabelRepository{db: db}
}

// Create создает новую пользовательскую метку
func (r *UserLabelRepository) Create(ctx context.Context, label *domain.UserLabel) error {
	query := `
		INSERT INTO user_labels (id, user_id, name, color, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		label.ID, label.UserID, label.Name, label.Color, label.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrUserLabelExists
		}
		return fmt.Errorf("insert user_label: %w", err)
	}

	return nil
}

// GetByID возвращает пользовательскую метку по ID
func (r *UserLabelRepository) GetByID(ctx context.Context, labelID string) (*domain.UserLabel, error) {
	query := `
		SELECT id, user_id, name, color, created_at
		FROM user_labels
		WHERE id = $1
	`

	var label domain.UserLabel
	err := r.db.QueryRowContext(ctx, query, labelID).Scan(
		&label.ID, &label.UserID, &label.Name, &label.Color, &label.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserLabelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user_label: %w", err)
	}

	return &label, nil
}

// ListByUserID возвращает все метки пользователя в порядке создания
func (r *UserLabelRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.UserLabel, error) {
	query := `
		SELECT id, user_id, name, color, created_at
		FROM user_labels
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("select user_labels: %w", err)
	}
	defer rows.Close()

	var labels []*domain.UserLabel
	for rows.Next() {
		var l domain.UserLabel
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user_label: %w", err)
		}
		labels = append(labels, &l)
	}

	return labels, rows.Err()
}

// Update обновляет пользовательскую метку (name, color)
func (r *UserLabelRepository) Update(ctx context.Context, label *domain.UserLabel) error {
	query := `
		UPDATE user_labels
		SET name = $1, color = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, label.Name, label.Color, label.ID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrUserLabelExists
		}
		return fmt.Errorf("update user_label: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserLabelNotFound
	}

	return nil
}

// Delete удаляет пользовательскую метку по ID
func (r *UserLabelRepository) Delete(ctx context.Context, labelID string) error {
	query := `DELETE FROM user_labels WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, labelID)
	if err != nil {
		return fmt.Errorf("delete user_label: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserLabelNotFound
	}

	return nil
}

// CountByUserID возвращает количество меток пользователя
func (r *UserLabelRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM user_labels WHERE user_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count user_labels: %w", err)
	}

	return count, nil
}
