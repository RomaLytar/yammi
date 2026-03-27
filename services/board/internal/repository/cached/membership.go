package cached

import (
	"context"
	"log/slog"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// MembershipRepository — декоратор над PostgreSQL репозиторием с Redis кешем.
// Реализует usecase.MembershipRepository — usecase слой не знает про кеш.
//
// Read path:  Redis → cache hit? return : PostgreSQL fallback
// Write path: PostgreSQL only (NATS событие обновит Redis через cache consumer)
type MembershipRepository struct {
	pg    usecase.MembershipRepository
	cache *cache.MembershipCache
}

func NewMembershipRepository(pg usecase.MembershipRepository, c *cache.MembershipCache) *MembershipRepository {
	return &MembershipRepository{pg: pg, cache: c}
}

// IsMember проверяет членство — горячий путь (80% всех запросов).
// 1. Redis HGET → O(1), микросекунды
// 2. Cache miss/error → PostgreSQL fallback
func (r *MembershipRepository) IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error) {
	// Fast path: Redis
	if r.cache != nil {
		role, found, err := r.cache.GetRole(ctx, boardID, userID)
		if err != nil {
			slog.Warn("cache: IsMember redis error, falling back to PostgreSQL", "error", err, "board_id", boardID)
		} else if found {
			return true, domain.Role(role), nil
		}
		// miss → fall through to PostgreSQL
	}

	// Slow path: PostgreSQL
	return r.pg.IsMember(ctx, boardID, userID)
}

// AddMember — write pass-through. Пишет ТОЛЬКО в PostgreSQL.
// Redis обновляется асинхронно через NATS событие member.added → cache consumer.
func (r *MembershipRepository) AddMember(ctx context.Context, boardID, userID string, role domain.Role) error {
	return r.pg.AddMember(ctx, boardID, userID, role)
}

// RemoveMember — write pass-through. Пишет ТОЛЬКО в PostgreSQL.
// Redis обновляется через NATS событие member.removed → cache consumer.
func (r *MembershipRepository) RemoveMember(ctx context.Context, boardID, userID string) error {
	return r.pg.RemoveMember(ctx, boardID, userID)
}

// ListMembers — pass-through к PostgreSQL (пагинация плохо ложится на Redis HGETALL).
func (r *MembershipRepository) ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error) {
	return r.pg.ListMembers(ctx, boardID, limit, offset)
}
