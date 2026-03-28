package cached

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// BoardSettingsRepository — lazy Redis кеш для настроек доски.
type BoardSettingsRepository struct {
	pg    usecase.BoardSettingsRepository
	cache *cache.DataCache
}

func NewBoardSettingsRepository(pg usecase.BoardSettingsRepository, c *cache.DataCache) *BoardSettingsRepository {
	return &BoardSettingsRepository{pg: pg, cache: c}
}

// GetByBoardID — hot path для Available Labels. Кешируется.
func (r *BoardSettingsRepository) GetByBoardID(ctx context.Context, boardID string) (*domain.BoardSettings, error) {
	if r.cache != nil {
		data, found, err := r.cache.GetBoardSettings(ctx, boardID)
		if err != nil {
			slog.Warn("cache: GetBoardSettings redis error", "error", err, "board_id", boardID)
		} else if found {
			var settings domain.BoardSettings
			if err := json.Unmarshal(data, &settings); err == nil {
				return &settings, nil
			}
		}
	}

	settings, err := r.pg.GetByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.SetBoardSettings(ctx, boardID, settings)
	}

	return settings, nil
}

// Upsert — write pass-through + invalidate cache.
func (r *BoardSettingsRepository) Upsert(ctx context.Context, settings *domain.BoardSettings) error {
	if err := r.pg.Upsert(ctx, settings); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateBoardSettings(ctx, settings.BoardID)
	}
	return nil
}
