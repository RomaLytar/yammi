package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type NotificationRepo struct {
	db *sql.DB
}

// escapeLikePattern экранирует спецсимволы LIKE/ILIKE в пользовательском вводе.
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func NewNotificationRepo(db *sql.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	metadata, err := json.Marshal(n.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query := `INSERT INTO notifications (id, user_id, type, title, message, metadata, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = retryExec(ctx, r.db, query,
		n.ID, n.UserID, string(n.Type), n.Title, n.Message, metadata, n.IsRead, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) BatchCreate(ctx context.Context, notifications []*domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString(`INSERT INTO notifications (id, user_id, type, title, message, metadata, is_read, created_at) VALUES `)

	args := make([]interface{}, 0, len(notifications)*8)
	for i, n := range notifications {
		if i > 0 {
			b.WriteString(", ")
		}
		offset := i * 8
		fmt.Fprintf(&b, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8)

		metadataJSON, err := json.Marshal(n.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata for %s: %w", n.ID, err)
		}
		args = append(args, n.ID, n.UserID, string(n.Type), n.Title, n.Message, metadataJSON, n.IsRead, n.CreatedAt)
	}

	_, err := retryExec(ctx, r.db, b.String(), args...)
	if err != nil {
		return fmt.Errorf("batch insert notifications: %w", err)
	}
	return nil
}

func (r *NotificationRepo) ListByUserID(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
	args := []interface{}{userID}
	conditions := []string{"user_id = $1"}
	argIdx := 2

	// Cursor-based pagination (created_at based)
	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339Nano, cursor)
		if err == nil {
			conditions = append(conditions, fmt.Sprintf("created_at < $%d", argIdx))
			args = append(args, cursorTime)
			argIdx++
		}
	}

	// Type filter — поддержка фильтра по категории (например "card" матчит card_created, card_deleted и т.д.)
	if typeFilter != "" {
		conditions = append(conditions, fmt.Sprintf("type LIKE $%d", argIdx))
		args = append(args, typeFilter+"%")
		argIdx++
	}

	// Search by title
	if search != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+escapeLikePattern(search)+"%")
		argIdx++
	}

	where := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(
		`SELECT id, user_id, type, title, message, metadata, is_read, created_at
		FROM notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d`, where, argIdx)
	args = append(args, limit+1) // Запрашиваем +1 для определения next_cursor

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var ntype string
		var metadataJSON []byte

		if err := rows.Scan(&n.ID, &n.UserID, &ntype, &n.Title, &n.Message, &metadataJSON, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scan notification: %w", err)
		}

		n.Type = domain.NotificationType(ntype)
		n.Metadata = make(map[string]string)
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
				return nil, "", fmt.Errorf("unmarshal metadata: %w", err)
			}
		}

		notifications = append(notifications, n)
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("rows error: %w", err)
	}

	var nextCursor string
	if len(notifications) > limit {
		// Есть ещё записи — курсор из последнего возвращаемого элемента
		notifications = notifications[:limit]
		nextCursor = notifications[limit-1].CreatedAt.Format(time.RFC3339Nano)
	}

	return notifications, nextCursor, nil
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, userID string, ids []string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, userID, pgtype.FlatArray[string](ids))
	if err != nil {
		return fmt.Errorf("mark as read: %w", err)
	}
	return nil
}

func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("mark all as read: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`
	err := retryQueryRow(ctx, r.db, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}
	return count, nil
}
