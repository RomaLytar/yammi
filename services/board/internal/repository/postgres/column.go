package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ColumnRepository struct {
	db *sql.DB
}

func NewColumnRepository(db *sql.DB) *ColumnRepository {
	return &ColumnRepository{db: db}
}

// Create создает новую колонку
func (r *ColumnRepository) Create(ctx context.Context, column *domain.Column) error {
	query := `
		INSERT INTO columns (id, board_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		column.ID, column.BoardID, column.Title, column.Position, column.CreatedAt, column.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert column: %w", err)
	}

	return nil
}

// GetByID возвращает колонку по ID
func (r *ColumnRepository) GetByID(ctx context.Context, columnID string) (*domain.Column, error) {
	query := `
		SELECT id, board_id, title, position, created_at, updated_at
		FROM columns
		WHERE id = $1
	`

	var column domain.Column
	err := r.db.QueryRowContext(ctx, query, columnID).Scan(
		&column.ID, &column.BoardID, &column.Title, &column.Position, &column.CreatedAt, &column.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrColumnNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select column: %w", err)
	}

	return &column, nil
}

// ListByBoardID возвращает все колонки доски в порядке position
func (r *ColumnRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Column, error) {
	query := `
		SELECT id, board_id, title, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("select columns: %w", err)
	}
	defer rows.Close()

	var columns []*domain.Column
	for rows.Next() {
		var c domain.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Title, &c.Position, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		columns = append(columns, &c)
	}

	return columns, rows.Err()
}

// Update обновляет колонку (title и/или position)
func (r *ColumnRepository) Update(ctx context.Context, column *domain.Column) error {
	query := `
		UPDATE columns
		SET title = $1, position = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query, column.Title, column.Position, column.UpdatedAt, column.ID)
	if err != nil {
		return fmt.Errorf("update column: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrColumnNotFound
	}

	return nil
}

// Delete удаляет колонку и все её карточки
func (r *ColumnRepository) Delete(ctx context.Context, columnID, boardID string) error {
	// Сначала удаляем карточки колонки (с board_id для partition pruning)
	_, err := r.db.ExecContext(ctx, `DELETE FROM cards WHERE column_id = $1 AND board_id = $2`, columnID, boardID)
	if err != nil {
		return fmt.Errorf("delete column cards: %w", err)
	}

	// Потом удаляем саму колонку
	result, err := r.db.ExecContext(ctx, `DELETE FROM columns WHERE id = $1`, columnID)
	if err != nil {
		return fmt.Errorf("delete column: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrColumnNotFound
	}

	return nil
}
