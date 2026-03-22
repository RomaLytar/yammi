package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type BoardEventRepo struct {
	db *sql.DB
}

func NewBoardEventRepo(db *sql.DB) *BoardEventRepo {
	return &BoardEventRepo{db: db}
}

// Create сохраняет 1 board event (заменяет N INSERT в notifications).
func (r *BoardEventRepo) Create(ctx context.Context, event *domain.BoardEvent) error {
	query := `
		INSERT INTO board_events (id, board_id, actor_id, event_type, title, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.ID, event.BoardID, event.ActorID, event.EventType,
		event.Title, event.Message, event.MetadataJSON(), event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert board_event: %w", err)
	}
	return nil
}

// ListForUser возвращает board events для пользователя (по его доскам) с cursor-based pagination.
// LEFT JOIN user_board_cursors для определения is_read.
func (r *BoardEventRepo) ListForUser(ctx context.Context, userID string, boardIDs []string, limit int, cursor, typeFilter, search string) ([]*domain.Notification, string, error) {
	if len(boardIDs) == 0 {
		return nil, "", nil
	}

	// Строим WHERE board_id IN ($1, $2, ...)
	placeholders := make([]string, len(boardIDs))
	args := make([]interface{}, 0, len(boardIDs)+5)
	args = append(args, userID)
	for i, bid := range boardIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, bid)
	}

	query := fmt.Sprintf(`
		SELECT be.id, be.board_id, be.actor_id, be.event_type, be.title, be.message, be.metadata, be.created_at,
			   COALESCE(be.created_at <= ubc.read_at, FALSE) as is_read
		FROM board_events be
		LEFT JOIN user_board_cursors ubc ON ubc.board_id = be.board_id AND ubc.user_id = $1
		WHERE be.board_id IN (%s)
		  AND be.actor_id != $1
	`, strings.Join(placeholders, ","))

	// Фильтр по типу (prefix: "card" → card_created, card_moved, ...)
	if typeFilter != "" {
		idx := len(args) + 1
		query += fmt.Sprintf(" AND be.event_type LIKE $%d", idx)
		args = append(args, typeFilter+"%")
	}

	// Поиск по title
	if search != "" {
		idx := len(args) + 1
		query += fmt.Sprintf(" AND be.title ILIKE $%d", idx)
		args = append(args, "%"+search+"%")
	}

	cursorIdx := len(args) + 1
	if cursor != "" {
		query += fmt.Sprintf(" AND be.created_at < $%d", cursorIdx)
		args = append(args, cursor)
	}

	limitIdx := len(args) + 1
	query += fmt.Sprintf(" ORDER BY be.created_at DESC LIMIT $%d", limitIdx)
	args = append(args, limit+1) // +1 для next cursor

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list board events: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var boardID, actorID, metadataJSON string
		var isRead bool

		if err := rows.Scan(&n.ID, &boardID, &actorID, &n.Type, &n.Title, &n.Message, &metadataJSON, &n.CreatedAt, &isRead); err != nil {
			return nil, "", fmt.Errorf("scan board event: %w", err)
		}

		n.UserID = userID
		n.IsRead = isRead
		n.Metadata = domain.ParseMetadataJSON(metadataJSON)
		n.Metadata["board_id"] = boardID
		n.Metadata["actor_id"] = actorID
		notifications = append(notifications, &n)
	}

	var nextCursor string
	if len(notifications) > limit {
		nextCursor = notifications[limit].CreatedAt.Format(time.RFC3339Nano)
		notifications = notifications[:limit]
	}

	return notifications, nextCursor, rows.Err()
}

// MarkBoardRead обновляет cursor для конкретной доски.
func (r *BoardEventRepo) MarkBoardRead(ctx context.Context, userID, boardID string) error {
	query := `
		INSERT INTO user_board_cursors (user_id, board_id, read_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, board_id) DO UPDATE SET read_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, userID, boardID)
	return err
}

// MarkAllBoardsRead обновляет cursors для всех досок пользователя.
func (r *BoardEventRepo) MarkAllBoardsRead(ctx context.Context, userID string, boardIDs []string) error {
	if len(boardIDs) == 0 {
		return nil
	}

	// Batch UPSERT
	values := make([]string, len(boardIDs))
	args := make([]interface{}, 0, len(boardIDs)+1)
	args = append(args, userID)
	for i, bid := range boardIDs {
		values[i] = fmt.Sprintf("($1, $%d, NOW())", i+2)
		args = append(args, bid)
	}

	query := fmt.Sprintf(`
		INSERT INTO user_board_cursors (user_id, board_id, read_at)
		VALUES %s
		ON CONFLICT (user_id, board_id) DO UPDATE SET read_at = NOW()
	`, strings.Join(values, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetBoardIDByEventID находит board_id по event ID (для mark specific as read).
func (r *BoardEventRepo) GetBoardIDByEventID(ctx context.Context, eventID string) (string, error) {
	var boardID string
	err := r.db.QueryRowContext(ctx, "SELECT board_id FROM board_events WHERE id = $1", eventID).Scan(&boardID)
	if err == sql.ErrNoRows {
		return "", nil // not a board event
	}
	return boardID, err
}
