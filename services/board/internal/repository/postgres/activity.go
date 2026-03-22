package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ActivityRepository struct {
	db *sql.DB
}

func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Create создает новую запись активности
func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	changesJSON, err := json.Marshal(activity.Changes)
	if err != nil {
		return fmt.Errorf("marshal changes: %w", err)
	}

	query := `
		INSERT INTO card_activities (id, card_id, board_id, actor_id, activity_type, description, changes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		activity.ID, activity.CardID, activity.BoardID, activity.ActorID,
		string(activity.Type), activity.Description, changesJSON, activity.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert activity: %w", err)
	}

	return nil
}

// ListByCardID возвращает записи активности по карточке с cursor-пагинацией
func (r *ActivityRepository) ListByCardID(ctx context.Context, cardID, boardID string, limit int, cursor string) ([]*domain.Activity, string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if cursor != "" {
		// Cursor = created_at предыдущей последней записи (ISO 8601)
		cursorTime, parseErr := time.Parse(time.RFC3339Nano, cursor)
		if parseErr != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", parseErr)
		}

		query := `
			SELECT id, card_id, board_id, actor_id, activity_type, description, changes, created_at
			FROM card_activities
			WHERE card_id = $1 AND board_id = $2 AND created_at < $3
			ORDER BY created_at DESC
			LIMIT $4
		`
		rows, err = r.db.QueryContext(ctx, query, cardID, boardID, cursorTime, limit+1)
	} else {
		query := `
			SELECT id, card_id, board_id, actor_id, activity_type, description, changes, created_at
			FROM card_activities
			WHERE card_id = $1 AND board_id = $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		rows, err = r.db.QueryContext(ctx, query, cardID, boardID, limit+1)
	}
	if err != nil {
		return nil, "", fmt.Errorf("select activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.Activity
	for rows.Next() {
		var a domain.Activity
		var activityType string
		var changesJSON []byte

		if err := rows.Scan(&a.ID, &a.CardID, &a.BoardID, &a.ActorID, &activityType, &a.Description, &changesJSON, &a.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scan activity: %w", err)
		}

		a.Type = domain.ActivityType(activityType)
		a.Changes = map[string]string{}
		if len(changesJSON) > 0 {
			if err := json.Unmarshal(changesJSON, &a.Changes); err != nil {
				return nil, "", fmt.Errorf("unmarshal changes: %w", err)
			}
		}

		activities = append(activities, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("rows error: %w", err)
	}

	// Определяем next_cursor
	var nextCursor string
	if len(activities) > limit {
		activities = activities[:limit]
		lastActivity := activities[len(activities)-1]
		nextCursor = lastActivity.CreatedAt.Format(time.RFC3339Nano)
	}

	return activities, nextCursor, nil
}
