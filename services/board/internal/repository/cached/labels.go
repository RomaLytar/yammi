package cached

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// LabelRepository — lazy Redis кеш для меток доски.
// Read: Redis → miss → PostgreSQL → populate cache.
// Write: PostgreSQL → invalidate Redis key.
type LabelRepository struct {
	pg    usecase.LabelRepository
	cache *cache.DataCache
}

func NewLabelRepository(pg usecase.LabelRepository, c *cache.DataCache) *LabelRepository {
	return &LabelRepository{pg: pg, cache: c}
}

// ListByBoardID — hot path для Available Labels. Кешируется в Redis.
func (r *LabelRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Label, error) {
	if r.cache != nil {
		data, found, err := r.cache.GetBoardLabels(ctx, boardID)
		if err != nil {
			slog.Warn("cache: ListByBoardID redis error", "error", err, "board_id", boardID)
		} else if found {
			var labels []*domain.Label
			if err := json.Unmarshal(data, &labels); err == nil {
				return labels, nil
			}
		}
	}

	labels, err := r.pg.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.SetBoardLabels(ctx, boardID, labels)
	}

	return labels, nil
}

// Create — write pass-through + invalidate cache.
func (r *LabelRepository) Create(ctx context.Context, label *domain.Label) error {
	if err := r.pg.Create(ctx, label); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateBoardLabels(ctx, label.BoardID)
	}
	return nil
}

func (r *LabelRepository) BatchCreate(ctx context.Context, labels []*domain.Label) error {
	if err := r.pg.BatchCreate(ctx, labels); err != nil {
		return err
	}
	if r.cache != nil && len(labels) > 0 {
		_ = r.cache.InvalidateBoardLabels(ctx, labels[0].BoardID)
	}
	return nil
}

func (r *LabelRepository) CreateWithLimit(ctx context.Context, label *domain.Label, maxCount int) error {
	if err := r.pg.CreateWithLimit(ctx, label, maxCount); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateBoardLabels(ctx, label.BoardID)
	}
	return nil
}

func (r *LabelRepository) Update(ctx context.Context, label *domain.Label) error {
	if err := r.pg.Update(ctx, label); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateBoardLabels(ctx, label.BoardID)
	}
	return nil
}

func (r *LabelRepository) Delete(ctx context.Context, labelID string) error {
	// Нужен boardID для инвалидации — получаем метку перед удалением
	label, err := r.pg.GetByID(ctx, labelID)
	if err != nil {
		return r.pg.Delete(ctx, labelID) // fallback: delete without invalidation
	}
	if err := r.pg.Delete(ctx, labelID); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateBoardLabels(ctx, label.BoardID)
	}
	return nil
}

// Pass-through methods (не кешируются)
func (r *LabelRepository) GetByID(ctx context.Context, labelID string) (*domain.Label, error) {
	return r.pg.GetByID(ctx, labelID)
}

func (r *LabelRepository) AddToCard(ctx context.Context, cardID, boardID, labelID string) error {
	return r.pg.AddToCard(ctx, cardID, boardID, labelID)
}

func (r *LabelRepository) RemoveFromCard(ctx context.Context, cardID, boardID, labelID string) error {
	return r.pg.RemoveFromCard(ctx, cardID, boardID, labelID)
}

func (r *LabelRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Label, error) {
	return r.pg.ListByCardID(ctx, cardID, boardID)
}

func (r *LabelRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	return r.pg.CountByBoardID(ctx, boardID)
}
