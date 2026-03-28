package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type BoardTemplateRepository struct {
	db *sql.DB
}

func NewBoardTemplateRepository(db *sql.DB) *BoardTemplateRepository {
	return &BoardTemplateRepository{db: db}
}

// Create создает новый шаблон доски
func (r *BoardTemplateRepository) Create(ctx context.Context, t *domain.BoardTemplate) error {
	columnsJSON, err := json.Marshal(t.ColumnsData)
	if err != nil {
		return fmt.Errorf("marshal columns_data: %w", err)
	}

	labelsJSON, err := json.Marshal(t.LabelsData)
	if err != nil {
		return fmt.Errorf("marshal labels_data: %w", err)
	}

	query := `
		INSERT INTO board_templates (id, user_id, name, description, columns_data, labels_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		t.ID, t.UserID, t.Name, t.Description,
		columnsJSON, labelsJSON,
		t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert board_template: %w", err)
	}

	return nil
}

// GetByID возвращает шаблон доски по ID
func (r *BoardTemplateRepository) GetByID(ctx context.Context, id string) (*domain.BoardTemplate, error) {
	query := `
		SELECT id, user_id, name, description, columns_data, labels_data, created_at, updated_at
		FROM board_templates
		WHERE id = $1
	`

	var t domain.BoardTemplate
	var columnsJSON, labelsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.Name, &t.Description,
		&columnsJSON, &labelsJSON,
		&t.CreatedAt, &t.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select board_template: %w", err)
	}

	if err := json.Unmarshal(columnsJSON, &t.ColumnsData); err != nil {
		return nil, fmt.Errorf("unmarshal columns_data: %w", err)
	}

	if err := json.Unmarshal(labelsJSON, &t.LabelsData); err != nil {
		return nil, fmt.Errorf("unmarshal labels_data: %w", err)
	}

	return &t, nil
}

// ListByUserID возвращает все шаблоны досок пользователя
func (r *BoardTemplateRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.BoardTemplate, error) {
	query := `
		SELECT id, user_id, name, description, columns_data, labels_data, created_at, updated_at
		FROM board_templates
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("select board_templates: %w", err)
	}
	defer rows.Close()

	var templates []*domain.BoardTemplate
	for rows.Next() {
		var t domain.BoardTemplate
		var columnsJSON, labelsJSON []byte

		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Name, &t.Description,
			&columnsJSON, &labelsJSON,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan board_template: %w", err)
		}

		if err := json.Unmarshal(columnsJSON, &t.ColumnsData); err != nil {
			return nil, fmt.Errorf("unmarshal columns_data: %w", err)
		}

		if err := json.Unmarshal(labelsJSON, &t.LabelsData); err != nil {
			return nil, fmt.Errorf("unmarshal labels_data: %w", err)
		}

		templates = append(templates, &t)
	}

	return templates, rows.Err()
}

// Update обновляет шаблон доски
func (r *BoardTemplateRepository) Update(ctx context.Context, t *domain.BoardTemplate) error {
	columnsJSON, err := json.Marshal(t.ColumnsData)
	if err != nil {
		return fmt.Errorf("marshal columns_data: %w", err)
	}

	labelsJSON, err := json.Marshal(t.LabelsData)
	if err != nil {
		return fmt.Errorf("marshal labels_data: %w", err)
	}

	query := `
		UPDATE board_templates
		SET name = $1, description = $2, columns_data = $3, labels_data = $4, updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.ExecContext(ctx, query,
		t.Name, t.Description, columnsJSON, labelsJSON, t.ID,
	)
	if err != nil {
		return fmt.Errorf("update board_template: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

// Delete удаляет шаблон доски по ID
func (r *BoardTemplateRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM board_templates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete board_template: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}
