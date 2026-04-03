package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type BoardSettingsRepository struct {
	db *sql.DB
}

func NewBoardSettingsRepository(db *sql.DB) *BoardSettingsRepository {
	return &BoardSettingsRepository{db: db}
}

// GetByBoardID возвращает настройки доски (дефолтные, если нет записи)
func (r *BoardSettingsRepository) GetByBoardID(ctx context.Context, boardID string) (*domain.BoardSettings, error) {
	query := `
		SELECT board_id, use_board_labels_only, done_column_id, sprint_duration_days, releases_enabled, created_at, updated_at
		FROM board_settings
		WHERE board_id = $1
	`

	var settings domain.BoardSettings
	var doneColumnID sql.NullString
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(
		&settings.BoardID, &settings.UseBoardLabelsOnly, &doneColumnID, &settings.SprintDurationDays, &settings.ReleasesEnabled, &settings.CreatedAt, &settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Возвращаем дефолтные настройки (UseBoardLabelsOnly = false)
		return domain.NewBoardSettings(boardID), nil
	}
	if err != nil {
		return nil, fmt.Errorf("select board_settings: %w", err)
	}

	if doneColumnID.Valid {
		settings.DoneColumnID = &doneColumnID.String
	}

	return &settings, nil
}

// Upsert создает или обновляет настройки доски (INSERT ON CONFLICT DO UPDATE)
func (r *BoardSettingsRepository) Upsert(ctx context.Context, settings *domain.BoardSettings) error {
	query := `
		INSERT INTO board_settings (board_id, use_board_labels_only, done_column_id, sprint_duration_days, releases_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (board_id) DO UPDATE SET
			use_board_labels_only = EXCLUDED.use_board_labels_only,
			done_column_id = EXCLUDED.done_column_id,
			sprint_duration_days = EXCLUDED.sprint_duration_days,
			releases_enabled = EXCLUDED.releases_enabled,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		settings.BoardID, settings.UseBoardLabelsOnly, settings.DoneColumnID, settings.SprintDurationDays, settings.ReleasesEnabled, settings.CreatedAt, settings.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert board_settings: %w", err)
	}

	return nil
}
