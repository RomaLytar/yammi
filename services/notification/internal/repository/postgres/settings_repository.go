package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type SettingsRepo struct {
	db *sql.DB
}

func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	query := `SELECT user_id, enabled, realtime_enabled, created_at, updated_at
		FROM notification_settings WHERE user_id = $1`

	s := &domain.NotificationSettings{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.UserID, &s.Enabled, &s.RealtimeEnabled, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DefaultSettings(userID), nil
		}
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return s, nil
}

func (r *SettingsRepo) BatchGet(ctx context.Context, userIDs []string) (map[string]*domain.NotificationSettings, error) {
	result := make(map[string]*domain.NotificationSettings, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	query := `SELECT user_id, enabled, realtime_enabled, created_at, updated_at
		FROM notification_settings WHERE user_id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, pgtype.FlatArray[string](userIDs))
	if err != nil {
		return nil, fmt.Errorf("batch get settings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		s := &domain.NotificationSettings{}
		if err := rows.Scan(&s.UserID, &s.Enabled, &s.RealtimeEnabled, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan settings: %w", err)
		}
		result[s.UserID] = s
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Заполняем пропущенных юзеров дефолтными настройками
	for _, uid := range userIDs {
		if _, ok := result[uid]; !ok {
			result[uid] = domain.DefaultSettings(uid)
		}
	}

	return result, nil
}

func (r *SettingsRepo) Upsert(ctx context.Context, settings *domain.NotificationSettings) error {
	query := `INSERT INTO notification_settings (user_id, enabled, realtime_enabled, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET enabled = $2, realtime_enabled = $3, updated_at = $4`

	_, err := r.db.ExecContext(ctx, query,
		settings.UserID, settings.Enabled, settings.RealtimeEnabled, settings.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
