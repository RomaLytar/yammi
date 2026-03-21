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

// ListByUserID возвращает доски с фильтрацией, поиском и сортировкой (offset pagination)
func (r *BoardRepository) ListByUserID(ctx context.Context, userID string, limit int, cursor string, ownerOnly bool, search string, sortBy string) ([]*domain.Board, string, error) {
	offset := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}

	// Build WHERE clause
	args := []interface{}{userID}
	where := "WHERE bm.user_id = $1"
	argIdx := 2

	if ownerOnly {
		where += fmt.Sprintf(" AND b.owner_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}

	if search != "" {
		where += fmt.Sprintf(" AND b.title ILIKE $%d", argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	// ORDER BY
	orderBy := "ORDER BY b.updated_at DESC, b.id DESC"
	switch sortBy {
	case "created_at":
		orderBy = "ORDER BY b.created_at DESC, b.id DESC"
	case "title":
		orderBy = "ORDER BY b.title ASC, b.id ASC"
	}

	query := fmt.Sprintf(`
		SELECT b.id, b.title, b.description, b.owner_id, b.version, b.created_at, b.updated_at
		FROM boards b
		INNER JOIN board_members bm ON b.id = bm.board_id
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, argIdx, argIdx+1)

	args = append(args, limit+1, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
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

	var nextCursor string
	if len(boards) > limit {
		boards = boards[:limit]
		nextCursor = fmt.Sprintf("%d", offset+limit)
	}

	return boards, nextCursor, rows.Err()
}

// BatchDelete удаляет несколько досок в одной транзакции
func (r *BoardRepository) BatchDelete(ctx context.Context, boardIDs []string) error {
	if len(boardIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Строим IN clause
	placeholders := ""
	args := make([]interface{}, len(boardIDs))
	for i, id := range boardIDs {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// Удаляем cards (партиционированная таблица, нет FK CASCADE)
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM cards WHERE board_id IN (%s)`, placeholders), args...); err != nil {
		return fmt.Errorf("delete cards: %w", err)
	}

	// Удаляем доски (columns и members удалятся по CASCADE)
	result, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM boards WHERE id IN (%s)`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("delete boards: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrBoardNotFound
	}

	return tx.Commit()
}

// TouchUpdatedAt обновляет updated_at доски (вызывается при изменении карточек/колонок)
func (r *BoardRepository) TouchUpdatedAt(ctx context.Context, boardID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE boards SET updated_at = $1 WHERE id = $2`, time.Now(), boardID)
	return err
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

// Delete удаляет доску (каскадно удаляет members, columns; cards удаляются явно т.к. партиционированная таблица без FK)
func (r *BoardRepository) Delete(ctx context.Context, boardID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Удаляем cards явно (партиционированная таблица, нет FK CASCADE)
	if _, err := tx.ExecContext(ctx, `DELETE FROM cards WHERE board_id = $1`, boardID); err != nil {
		return fmt.Errorf("delete cards: %w", err)
	}

	// Удаляем доску (columns и members удалятся по CASCADE)
	result, err := tx.ExecContext(ctx, `DELETE FROM boards WHERE id = $1`, boardID)
	if err != nil {
		return fmt.Errorf("delete board: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrBoardNotFound
	}

	return tx.Commit()
}
