package cached

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// UserLabelRepository — lazy Redis кеш для пользовательских меток.
type UserLabelRepository struct {
	pg    usecase.UserLabelRepository
	cache *cache.DataCache
}

func NewUserLabelRepository(pg usecase.UserLabelRepository, c *cache.DataCache) *UserLabelRepository {
	return &UserLabelRepository{pg: pg, cache: c}
}

// ListByUserID — hot path для Available Labels. Кешируется.
func (r *UserLabelRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.UserLabel, error) {
	if r.cache != nil {
		data, found, err := r.cache.GetUserLabels(ctx, userID)
		if err != nil {
			slog.Warn("cache: ListByUserID redis error", "error", err, "user_id", userID)
		} else if found {
			var labels []*domain.UserLabel
			if err := json.Unmarshal(data, &labels); err == nil {
				return labels, nil
			}
		}
	}

	labels, err := r.pg.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		_ = r.cache.SetUserLabels(ctx, userID, labels)
	}

	return labels, nil
}

// Create — write pass-through + invalidate.
func (r *UserLabelRepository) Create(ctx context.Context, label *domain.UserLabel) error {
	if err := r.pg.Create(ctx, label); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateUserLabels(ctx, label.UserID)
	}
	return nil
}

func (r *UserLabelRepository) CreateWithLimit(ctx context.Context, label *domain.UserLabel, maxCount int) error {
	if err := r.pg.CreateWithLimit(ctx, label, maxCount); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateUserLabels(ctx, label.UserID)
	}
	return nil
}

func (r *UserLabelRepository) Update(ctx context.Context, label *domain.UserLabel) error {
	if err := r.pg.Update(ctx, label); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateUserLabels(ctx, label.UserID)
	}
	return nil
}

func (r *UserLabelRepository) Delete(ctx context.Context, labelID string) error {
	label, err := r.pg.GetByID(ctx, labelID)
	if err != nil {
		return r.pg.Delete(ctx, labelID)
	}
	if err := r.pg.Delete(ctx, labelID); err != nil {
		return err
	}
	if r.cache != nil {
		_ = r.cache.InvalidateUserLabels(ctx, label.UserID)
	}
	return nil
}

// Pass-through
func (r *UserLabelRepository) GetByID(ctx context.Context, labelID string) (*domain.UserLabel, error) {
	return r.pg.GetByID(ctx, labelID)
}

func (r *UserLabelRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	return r.pg.CountByUserID(ctx, userID)
}
