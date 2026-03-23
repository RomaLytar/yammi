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

// scanCard сканирует карточку из строки результата (DRY helper)
func scanCard(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Card, error) {
	var card domain.Card
	var assigneeID sql.NullString
	var dueDate sql.NullTime
	var priority, taskType string

	err := scanner.Scan(
		&card.ID, &card.ColumnID, &card.Title, &card.Description,
		&card.Position, &assigneeID, &card.CreatorID,
		&dueDate, &priority, &taskType,
		&card.CreatedAt, &card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		card.AssigneeID = &assigneeID.String
	}
	if dueDate.Valid {
		card.DueDate = &dueDate.Time
	}
	card.Priority = domain.Priority(priority)
	card.TaskType = domain.TaskType(taskType)

	return &card, nil
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
		INSERT INTO cards (id, column_id, board_id, title, description, position, assignee_id, creator_id, due_date, priority, task_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err = r.db.ExecContext(ctx, query,
		card.ID, card.ColumnID, boardID, card.Title, card.Description,
		card.Position, card.AssigneeID, card.CreatorID,
		card.DueDate, string(card.Priority), string(card.TaskType),
		card.CreatedAt, card.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert card: %w", err)
	}

	return nil
}

// GetByID возвращает карточку по ID с проверкой принадлежности к доске (IDOR protection)
func (r *CardRepository) GetByID(ctx context.Context, cardID, boardID string) (*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, due_date, priority, task_type, created_at, updated_at
		FROM cards
		WHERE id = $1 AND board_id = $2
	`

	card, err := scanCard(r.db.QueryRowContext(ctx, query, cardID, boardID))
	if err == sql.ErrNoRows {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select card: %w", err)
	}

	return card, nil
}

// GetLastInColumn возвращает последнюю карточку в колонке (для генерации lexorank)
func (r *CardRepository) GetLastInColumn(ctx context.Context, columnID string) (*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, due_date, priority, task_type, created_at, updated_at
		FROM cards
		WHERE column_id = $1
		ORDER BY position DESC
		LIMIT 1
	`

	card, err := scanCard(r.db.QueryRowContext(ctx, query, columnID))
	if err == sql.ErrNoRows {
		return nil, domain.ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get last card in column: %w", err)
	}

	return card, nil
}

// ListByColumnID возвращает все карточки колонки в порядке lexorank position
func (r *CardRepository) ListByColumnID(ctx context.Context, columnID string) ([]*domain.Card, error) {
	query := `
		SELECT id, column_id, title, description, position, assignee_id, creator_id, due_date, priority, task_type, created_at, updated_at
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
		card, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// Update обновляет карточку (title, description, assignee, position, column_id, due_date, priority, task_type)
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
		SET column_id = $1, title = $2, description = $3, position = $4, assignee_id = $5,
		    due_date = $6, priority = $7, task_type = $8, updated_at = $9
		WHERE id = $10 AND board_id = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		card.ColumnID, card.Title, card.Description, card.Position, card.AssigneeID,
		card.DueDate, string(card.Priority), string(card.TaskType), card.UpdatedAt,
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

// Delete удаляет карточку (boardID для partition pruning)
func (r *CardRepository) Delete(ctx context.Context, cardID, boardID string) error {
	query := `DELETE FROM cards WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, cardID, boardID)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}

// UnassignByUser снимает assignee со всех карточек удалённого участника
func (r *CardRepository) UnassignByUser(ctx context.Context, boardID, userID string) (int, error) {
	query := `UPDATE cards SET assignee_id = NULL, updated_at = NOW() WHERE board_id = $1 AND assignee_id = $2`
	result, err := r.db.ExecContext(ctx, query, boardID, userID)
	if err != nil {
		return 0, fmt.Errorf("unassign by user: %w", err)
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// CountByBoard возвращает количество карточек по колонкам доски (один запрос)
func (r *CardRepository) CountByBoard(ctx context.Context, boardID string) (map[string]int, error) {
	query := `SELECT column_id, COUNT(*) FROM cards WHERE board_id = $1 GROUP BY column_id`

	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("count cards by board: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var columnID string
		var count int
		if err := rows.Scan(&columnID, &count); err != nil {
			return nil, fmt.Errorf("scan card count: %w", err)
		}
		counts[columnID] = count
	}
	return counts, rows.Err()
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
