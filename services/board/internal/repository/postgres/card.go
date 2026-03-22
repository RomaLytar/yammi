package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// Create создает новую карточку (партиции прозрачны)
func (r *CardRepository) Create(ctx context.Context, card *domain.Card) error {
	// Сначала получаем board_id через column_id
	var boardID string
	err := r.db.QueryRowContext(ctx, "SELECT board_id FROM columns WHERE id = $1", card.ColumnID).Scan(&boardID)
	if err == sql.ErrNoRows {
		return domain.ErrColumnNotFound
	}
	if err != nil {
		return fmt.Errorf("select board_id: %w", err)
	}

	query := `
		INSERT INTO cards (id, column_id, board_id, title, description, position, assignee_id, creator_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.ExecContext(ctx, query,
		card.ID, card.ColumnID, boardID, card.Title, card.Description,
		card.Position, card.AssigneeID, card.CreatorID, card.CreatedAt, card.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert card: %w", err)
	}

	return nil
}

// GetByID возвращает карточку по ID с проверкой принадлежности к доске (IDOR protection)
func (r *CardRepository) GetByID(ctx context.Context, cardID, boardID string) (*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, created_at, updated_at
		FROM cards
		WHERE id = $1 AND board_id = $2
	`

	var card domain.Card
	var assigneeID sql.NullString

	err := r.db.QueryRowContext(ctx, query, cardID, boardID).Scan(
		&card.ID, &card.ColumnID, &card.Title, &card.Description,
		&card.Position, &assigneeID, &card.CreatorID, &card.CreatedAt, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select card: %w", err)
	}

	if assigneeID.Valid {
		card.AssigneeID = &assigneeID.String
	}

	return &card, nil
}

// GetLastInColumn возвращает последнюю карточку в колонке (для генерации lexorank)
func (r *CardRepository) GetLastInColumn(ctx context.Context, columnID string) (*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, created_at, updated_at
		FROM cards
		WHERE column_id = $1
		ORDER BY position DESC
		LIMIT 1
	`

	var card domain.Card
	var assigneeID sql.NullString

	err := r.db.QueryRowContext(ctx, query, columnID).Scan(
		&card.ID, &card.ColumnID, &card.Title, &card.Description,
		&card.Position, &assigneeID, &card.CreatorID, &card.CreatedAt, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get last card in column: %w", err)
	}

	if assigneeID.Valid {
		card.AssigneeID = &assigneeID.String
	}

	return &card, nil
}

// ListByColumnID возвращает все карточки колонки в порядке lexorank position
func (r *CardRepository) ListByColumnID(ctx context.Context, columnID string) ([]*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, created_at, updated_at
		FROM cards
		WHERE column_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, columnID)
	if err != nil {
		return nil, fmt.Errorf("select cards: %w", err)
	}
	defer rows.Close()

	var cards []*domain.Card
	for rows.Next() {
		var c domain.Card
		var assigneeID sql.NullString

		if err := rows.Scan(&c.ID, &c.ColumnID, &c.Title, &c.Description, &c.Position, &assigneeID, &c.CreatorID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}

		if assigneeID.Valid {
			c.AssigneeID = &assigneeID.String
		}

		cards = append(cards, &c)
	}

	return cards, rows.Err()
}

// Update обновляет карточку (title, description, assignee, position, column_id)
func (r *CardRepository) Update(ctx context.Context, card *domain.Card) error {
	// Получаем board_id для обновления (нужен для партиционирования)
	var boardID string
	err := r.db.QueryRowContext(ctx, "SELECT board_id FROM columns WHERE id = $1", card.ColumnID).Scan(&boardID)
	if err == sql.ErrNoRows {
		return domain.ErrColumnNotFound
	}
	if err != nil {
		return fmt.Errorf("select board_id: %w", err)
	}

	query := `
		UPDATE cards
		SET column_id = $1, title = $2, description = $3, position = $4, assignee_id = $5, updated_at = $6
		WHERE id = $7 AND board_id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		card.ColumnID, card.Title, card.Description, card.Position, card.AssigneeID, card.UpdatedAt,
		card.ID, boardID,
	)
	if err != nil {
		return fmt.Errorf("update card: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}

// Delete удаляет карточку
func (r *CardRepository) Delete(ctx context.Context, cardID string) error {
	query := `DELETE FROM cards WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, cardID)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}

// BatchDelete удаляет несколько карточек по ID в рамках одной доски (partition key)
func (r *CardRepository) BatchDelete(ctx context.Context, boardID string, cardIDs []string) error {
	if len(cardIDs) == 0 {
		return nil
	}

	// Строим запрос с плейсхолдерами: DELETE FROM cards WHERE board_id = $1 AND id IN ($2, $3, ...)
	query := `DELETE FROM cards WHERE board_id = $1 AND id IN (`
	args := make([]interface{}, 0, len(cardIDs)+1)
	args = append(args, boardID)

	for i, id := range cardIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}
	query += ")"

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch delete cards: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}
