package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type LabelRepository struct {
	db *sql.DB
}

func NewLabelRepository(db *sql.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

// Create создает новую метку
func (r *LabelRepository) Create(ctx context.Context, label *domain.Label) error {
	query := `
		INSERT INTO labels (id, board_id, name, color, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		label.ID, label.BoardID, label.Name, label.Color, label.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrLabelExists
		}
		return fmt.Errorf("insert label: %w", err)
	}

	return nil
}

// GetByID возвращает метку по ID
func (r *LabelRepository) GetByID(ctx context.Context, labelID string) (*domain.Label, error) {
	query := `
		SELECT id, board_id, name, color, created_at
		FROM labels
		WHERE id = $1
	`

	var label domain.Label
	err := r.db.QueryRowContext(ctx, query, labelID).Scan(
		&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrLabelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select label: %w", err)
	}

	return &label, nil
}

// ListByBoardID возвращает все метки доски в порядке создания
func (r *LabelRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Label, error) {
	query := `
		SELECT id, board_id, name, color, created_at
		FROM labels
		WHERE board_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("select labels: %w", err)
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan label: %w", err)
		}
		labels = append(labels, &l)
	}

	return labels, rows.Err()
}

// Update обновляет метку (name, color)
func (r *LabelRepository) Update(ctx context.Context, label *domain.Label) error {
	query := `
		UPDATE labels
		SET name = $1, color = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, label.Name, label.Color, label.ID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrLabelExists
		}
		return fmt.Errorf("update label: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrLabelNotFound
	}

	return nil
}

// Delete удаляет метку по ID (CASCADE удалит card_labels)
func (r *LabelRepository) Delete(ctx context.Context, labelID string) error {
	query := `DELETE FROM labels WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, labelID)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrLabelNotFound
	}

	return nil
}

// AddToCard назначает метку на карточку (board_id для partition pruning)
func (r *LabelRepository) AddToCard(ctx context.Context, cardID, boardID, labelID string) error {
	query := `
		INSERT INTO card_labels (card_id, board_id, label_id, created_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.db.ExecContext(ctx, query, cardID, boardID, labelID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrLabelAlreadyOnCard
		}
		return fmt.Errorf("add label to card: %w", err)
	}

	return nil
}

// RemoveFromCard снимает метку с карточки
func (r *LabelRepository) RemoveFromCard(ctx context.Context, cardID, boardID, labelID string) error {
	query := `DELETE FROM card_labels WHERE card_id = $1 AND board_id = $2 AND label_id = $3`
	result, err := r.db.ExecContext(ctx, query, cardID, boardID, labelID)
	if err != nil {
		return fmt.Errorf("remove label from card: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrLabelNotFound
	}

	return nil
}

// ListByCardID возвращает все метки карточки (JOIN через card_labels)
func (r *LabelRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Label, error) {
	query := `
		SELECT l.id, l.board_id, l.name, l.color, l.created_at
		FROM labels l
		INNER JOIN card_labels cl ON cl.label_id = l.id
		WHERE cl.card_id = $1 AND cl.board_id = $2
		ORDER BY l.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select card labels: %w", err)
	}
	defer rows.Close()

	var labels []*domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan card label: %w", err)
		}
		labels = append(labels, &l)
	}

	return labels, rows.Err()
}

// CountByBoardID возвращает количество меток доски
func (r *LabelRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM labels WHERE board_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count labels: %w", err)
	}

	return count, nil
}
