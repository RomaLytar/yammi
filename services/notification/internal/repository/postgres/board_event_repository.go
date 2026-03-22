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
func (r *BoardEventRepo) Create(ctx context.Context, event *domain.BoardEvent) (int64, error) {
	query := `
		INSERT INTO board_events (id, board_id, actor_id, event_type, title, message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING event_seq
	`
	var seq int64
	err := r.db.QueryRowContext(ctx, query,
		event.ID, event.BoardID, event.ActorID, event.EventType,
		event.Title, event.Message, event.MetadataJSON(), event.CreatedAt,
	).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("insert board_event: %w", err)
	}
	return seq, nil
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

// MarkBoardRead обновляет cursor + last_seen_seq для конкретной доски.
func (r *BoardEventRepo) MarkBoardRead(ctx context.Context, userID, boardID string) error {
	query := `
		INSERT INTO user_board_cursors (user_id, board_id, read_at, last_seen_seq)
		VALUES ($1, $2, NOW(), COALESCE((SELECT MAX(event_seq) FROM board_events WHERE board_id = $2), 0))
		ON CONFLICT (user_id, board_id) DO UPDATE SET
			read_at = NOW(),
			last_seen_seq = COALESCE((SELECT MAX(event_seq) FROM board_events WHERE board_id = $2), 0)
	`
	_, err := r.db.ExecContext(ctx, query, userID, boardID)
	return err
}

// MarkAllBoardsRead обновляет cursors + last_seen_seq для всех досок пользователя.
func (r *BoardEventRepo) MarkAllBoardsRead(ctx context.Context, userID string, boardIDs []string) error {
	if len(boardIDs) == 0 {
		return nil
	}

	// Обновляем каждую доску с её max(event_seq)
	for _, boardID := range boardIDs {
		if err := r.MarkBoardRead(ctx, userID, boardID); err != nil {
			return err
		}
	}
	return nil
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

// GetUserCursors возвращает last_seen_seq для каждой доски пользователя (1 SQL query).
func (r *BoardEventRepo) GetUserCursors(ctx context.Context, userID string, boardIDs []string) (map[string]int64, error) {
	if len(boardIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(boardIDs))
	args := make([]interface{}, 0, len(boardIDs)+1)
	args = append(args, userID)
	for i, bid := range boardIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, bid)
	}

	query := fmt.Sprintf(`
		SELECT board_id, last_seen_seq FROM user_board_cursors
		WHERE user_id = $1 AND board_id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get user cursors: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int64)
	for rows.Next() {
		var boardID string
		var seq int64
		if err := rows.Scan(&boardID, &seq); err != nil {
			return nil, err
		}
		result[boardID] = seq
	}
	return result, rows.Err()
}

// GetUnreadCountBySeq возвращает unread count через event_seq diff.
// O(1) per board: max(event_seq) - COALESCE(last_seen_seq, 0).
func (r *BoardEventRepo) GetUnreadCountBySeq(ctx context.Context, userID string, boardIDs []string) (int, error) {
	if len(boardIDs) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(boardIDs))
	args := make([]interface{}, 0, len(boardIDs)+1)
	args = append(args, userID)
	for i, bid := range boardIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, bid)
	}

	// O(1) per board: MAX(event_seq) - last_seen_seq. Без COUNT, без scan.
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(GREATEST(
			COALESCE(max_seq, 0) - COALESCE(last_seen, 0), 0
		)), 0)
		FROM (
			SELECT
				(SELECT MAX(event_seq) FROM board_events WHERE board_id = b.id) AS max_seq,
				(SELECT last_seen_seq FROM user_board_cursors WHERE board_id = b.id AND user_id = $1) AS last_seen
			FROM unnest(ARRAY[%s]::uuid[]) AS b(id)
		) diffs
	`, strings.Join(placeholders, ","))

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count by seq: %w", err)
	}
	return count, nil
}
