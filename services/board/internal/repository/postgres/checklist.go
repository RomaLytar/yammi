package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ChecklistRepository struct {
	db *sql.DB
}

func NewChecklistRepository(db *sql.DB) *ChecklistRepository {
	return &ChecklistRepository{db: db}
}

// CreateChecklist создает новый чеклист
func (r *ChecklistRepository) CreateChecklist(ctx context.Context, checklist *domain.Checklist) error {
	query := `
		INSERT INTO checklists (id, card_id, board_id, title, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		checklist.ID, checklist.CardID, checklist.BoardID,
		checklist.Title, checklist.Position,
		checklist.CreatedAt, checklist.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert checklist: %w", err)
	}

	return nil
}

// GetChecklistByID возвращает чеклист по ID (с элементами)
func (r *ChecklistRepository) GetChecklistByID(ctx context.Context, checklistID, boardID string) (*domain.Checklist, error) {
	query := `
		SELECT id, card_id, board_id, title, position, created_at, updated_at
		FROM checklists
		WHERE id = $1 AND board_id = $2
	`

	var checklist domain.Checklist
	err := r.db.QueryRowContext(ctx, query, checklistID, boardID).Scan(
		&checklist.ID, &checklist.CardID, &checklist.BoardID,
		&checklist.Title, &checklist.Position,
		&checklist.CreatedAt, &checklist.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrChecklistNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select checklist: %w", err)
	}

	// Загружаем элементы чеклиста
	items, err := r.ListItemsByChecklistID(ctx, checklistID, boardID)
	if err != nil {
		return nil, err
	}
	checklist.Items = items

	return &checklist, nil
}

// ListByCardID возвращает все чеклисты карточки (с элементами)
func (r *ChecklistRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Checklist, error) {
	query := `
		SELECT id, card_id, board_id, title, position, created_at, updated_at
		FROM checklists
		WHERE card_id = $1 AND board_id = $2
		ORDER BY position ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select checklists: %w", err)
	}
	defer rows.Close()

	var checklists []*domain.Checklist
	for rows.Next() {
		var cl domain.Checklist
		if err := rows.Scan(
			&cl.ID, &cl.CardID, &cl.BoardID,
			&cl.Title, &cl.Position,
			&cl.CreatedAt, &cl.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan checklist: %w", err)
		}
		checklists = append(checklists, &cl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Загружаем элементы для каждого чеклиста
	for _, cl := range checklists {
		items, err := r.ListItemsByChecklistID(ctx, cl.ID, boardID)
		if err != nil {
			return nil, err
		}
		cl.Items = items
	}

	return checklists, nil
}

// UpdateChecklist обновляет чеклист (title)
func (r *ChecklistRepository) UpdateChecklist(ctx context.Context, checklist *domain.Checklist) error {
	query := `
		UPDATE checklists
		SET title = $1, updated_at = $2
		WHERE id = $3 AND board_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		checklist.Title, checklist.UpdatedAt,
		checklist.ID, checklist.BoardID,
	)
	if err != nil {
		return fmt.Errorf("update checklist: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChecklistNotFound
	}

	return nil
}

// DeleteChecklist удаляет чеклист по ID (ручной CASCADE для items — партиции не поддерживают FK)
func (r *ChecklistRepository) DeleteChecklist(ctx context.Context, checklistID, boardID string) error {
	// Сначала удаляем элементы чеклиста
	deleteItemsQuery := `DELETE FROM checklist_items WHERE checklist_id = $1 AND board_id = $2`
	_, err := r.db.ExecContext(ctx, deleteItemsQuery, checklistID, boardID)
	if err != nil {
		return fmt.Errorf("delete checklist items: %w", err)
	}

	// Затем удаляем сам чеклист
	query := `DELETE FROM checklists WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, checklistID, boardID)
	if err != nil {
		return fmt.Errorf("delete checklist: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChecklistNotFound
	}

	return nil
}

// CreateItem создает новый элемент чеклиста
func (r *ChecklistRepository) CreateItem(ctx context.Context, item *domain.ChecklistItem) error {
	query := `
		INSERT INTO checklist_items (id, checklist_id, board_id, title, is_checked, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.ChecklistID, item.BoardID,
		item.Title, item.IsChecked, item.Position,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert checklist item: %w", err)
	}

	return nil
}

// GetItemByID возвращает элемент чеклиста по ID
func (r *ChecklistRepository) GetItemByID(ctx context.Context, itemID, boardID string) (*domain.ChecklistItem, error) {
	query := `
		SELECT id, checklist_id, board_id, title, is_checked, position, created_at, updated_at
		FROM checklist_items
		WHERE id = $1 AND board_id = $2
	`

	var item domain.ChecklistItem
	err := r.db.QueryRowContext(ctx, query, itemID, boardID).Scan(
		&item.ID, &item.ChecklistID, &item.BoardID,
		&item.Title, &item.IsChecked, &item.Position,
		&item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrChecklistItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select checklist item: %w", err)
	}

	return &item, nil
}

// ListItemsByChecklistID возвращает все элементы чеклиста
func (r *ChecklistRepository) ListItemsByChecklistID(ctx context.Context, checklistID, boardID string) ([]domain.ChecklistItem, error) {
	query := `
		SELECT id, checklist_id, board_id, title, is_checked, position, created_at, updated_at
		FROM checklist_items
		WHERE checklist_id = $1 AND board_id = $2
		ORDER BY position ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, checklistID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select checklist items: %w", err)
	}
	defer rows.Close()

	var items []domain.ChecklistItem
	for rows.Next() {
		var item domain.ChecklistItem
		if err := rows.Scan(
			&item.ID, &item.ChecklistID, &item.BoardID,
			&item.Title, &item.IsChecked, &item.Position,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan checklist item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateItem обновляет элемент чеклиста (title)
func (r *ChecklistRepository) UpdateItem(ctx context.Context, item *domain.ChecklistItem) error {
	query := `
		UPDATE checklist_items
		SET title = $1, updated_at = $2
		WHERE id = $3 AND board_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		item.Title, item.UpdatedAt,
		item.ID, item.BoardID,
	)
	if err != nil {
		return fmt.Errorf("update checklist item: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChecklistItemNotFound
	}

	return nil
}

// DeleteItem удаляет элемент чеклиста по ID
func (r *ChecklistRepository) DeleteItem(ctx context.Context, itemID, boardID string) error {
	query := `DELETE FROM checklist_items WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, itemID, boardID)
	if err != nil {
		return fmt.Errorf("delete checklist item: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChecklistItemNotFound
	}

	return nil
}

// ToggleItem переключает состояние is_checked элемента
func (r *ChecklistRepository) ToggleItem(ctx context.Context, itemID, boardID string, isChecked bool) error {
	query := `
		UPDATE checklist_items
		SET is_checked = $1, updated_at = NOW()
		WHERE id = $2 AND board_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, isChecked, itemID, boardID)
	if err != nil {
		return fmt.Errorf("toggle checklist item: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChecklistItemNotFound
	}

	return nil
}

// ToggleItemAtomic атомарно инвертирует is_checked и возвращает новое значение (один запрос вместо SELECT + UPDATE)
func (r *ChecklistRepository) ToggleItemAtomic(ctx context.Context, itemID, boardID string) (bool, error) {
	query := `
		UPDATE checklist_items
		SET is_checked = NOT is_checked, updated_at = NOW()
		WHERE id = $1 AND board_id = $2
		RETURNING is_checked
	`

	var newChecked bool
	err := r.db.QueryRowContext(ctx, query, itemID, boardID).Scan(&newChecked)
	if err == sql.ErrNoRows {
		return false, domain.ErrChecklistItemNotFound
	}
	if err != nil {
		return false, fmt.Errorf("toggle checklist item atomic: %w", err)
	}

	return newChecked, nil
}
