package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type BoardRepository struct {
	db *sql.DB
}

func NewBoardRepository(db *sql.DB) *BoardRepository {
	return &BoardRepository{db: db}
}

// Create создает доску + автоматически добавляет owner в board_members
func (r *BoardRepository) Create(ctx context.Context, board *domain.Board) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. INSERT board
	query := `
		INSERT INTO boards (id, title, description, owner_id, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query,
		board.ID, board.Title, board.Description, board.OwnerID, board.Version, board.CreatedAt, board.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert board: %w", err)
	}

	// 2. INSERT owner в board_members
	memberQuery := `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, 'owner', $3)
	`
	_, err = tx.ExecContext(ctx, memberQuery, board.ID, board.OwnerID, board.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert owner member: %w", err)
	}

	return tx.Commit()
}

// GetByID загружает board (БЕЗ members, БЕЗ columns)
func (r *BoardRepository) GetByID(ctx context.Context, boardID string) (*domain.Board, error) {
	query := `
		SELECT id, title, description, owner_id, version, created_at, updated_at
		FROM boards
		WHERE id = $1
	`

	var board domain.Board
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(
		&board.ID, &board.Title, &board.Description, &board.OwnerID,
		&board.Version, &board.CreatedAt, &board.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrBoardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select board: %w", err)
	}

	return &board, nil
}

// ListByUserID возвращает доски где user is member (cursor pagination)
func (r *BoardRepository) ListByUserID(ctx context.Context, userID string, limit int, cursor string) ([]*domain.Board, string, error) {
	// Cursor format: "created_at_RFC3339Nano|id"
	var createdAt time.Time
	var cursorID string

	if cursor != "" {
		// Parse cursor: "2024-03-19T10:00:00.123456789Z|uuid"
		var err error
		_, err = fmt.Sscanf(cursor, "%s|%s", &createdAt, &cursorID)
		if err != nil {
			createdAt = time.Time{} // Invalid cursor — start from beginning
			cursorID = ""
		}
	}

	// Если cursor пустой, используем максимальные значения для начала
	if cursor == "" {
		createdAt = time.Now().Add(24 * time.Hour) // будущее время
		cursorID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	}

	query := `
		SELECT b.id, b.title, b.description, b.owner_id, b.version, b.created_at, b.updated_at
		FROM boards b
		INNER JOIN board_members bm ON b.id = bm.board_id
		WHERE bm.user_id = $1
		  AND (b.created_at, b.id) < ($2, $3)
		ORDER BY b.created_at DESC, b.id DESC
		LIMIT $4
	`

	rows, err := r.db.QueryContext(ctx, query, userID, createdAt, cursorID, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("select boards: %w", err)
	}
	defer rows.Close()

	var boards []*domain.Board
	for rows.Next() {
		var b domain.Board
		if err := rows.Scan(&b.ID, &b.Title, &b.Description, &b.OwnerID, &b.Version, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, &b)
	}

	// Generate next cursor
	var nextCursor string
	if len(boards) > limit {
		last := boards[limit-1]
		nextCursor = fmt.Sprintf("%s|%s", last.CreatedAt.Format(time.RFC3339Nano), last.ID)
		boards = boards[:limit]
	}

	return boards, nextCursor, rows.Err()
}

// Update с optimistic locking
func (r *BoardRepository) Update(ctx context.Context, board *domain.Board) error {
	query := `
		UPDATE boards
		SET title = $1, description = $2, version = $3, updated_at = $4
		WHERE id = $5 AND version = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		board.Title, board.Description, board.Version, board.UpdatedAt,
		board.ID, board.Version-1, // проверяем старую версию
	)
	if err != nil {
		return fmt.Errorf("update board: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrInvalidVersion // optimistic lock conflict
	}

	return nil
}

// Delete удаляет доску (каскадно удаляет members, columns, cards)
func (r *BoardRepository) Delete(ctx context.Context, boardID string) error {
	query := `DELETE FROM boards WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, boardID)
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrBoardNotFound
	}

	return nil
}
