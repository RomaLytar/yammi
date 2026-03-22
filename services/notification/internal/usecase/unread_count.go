package usecase

import (
	"context"
	"log"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type GetUnreadCountUseCase struct {
	boardEventRepo BoardEventRepository
	memberRepo     BoardMemberRepository
	repo           NotificationRepository
	cache          UnreadCounter // Redis — lazy cache, НЕ источник истины
}

func NewGetUnreadCountUseCase(
	boardEventRepo BoardEventRepository,
	memberRepo BoardMemberRepository,
	repo NotificationRepository,
	cache UnreadCounter,
) *GetUnreadCountUseCase {
	return &GetUnreadCountUseCase{
		boardEventRepo: boardEventRepo,
		memberRepo:     memberRepo,
		repo:           repo,
		cache:          cache,
	}
}

// Execute: Redis GET (cache hit) → SQL seq diff (cache miss) → SET в Redis.
// Write path НЕ обновляет Redis. Инвалидация — через MarkRead/MarkAllRead.
func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}

	// 1. Попробовать Redis cache
	if uc.cache != nil {
		if cached, err := uc.cache.Get(ctx, userID); err == nil && cached >= 0 {
			return cached, nil
		}
	}

	// 2. Cache miss → вычисляем из SQL
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list boards for user %s: %v", userID, err)
		boardIDs = nil
	}

	boardUnread := 0
	if len(boardIDs) > 0 {
		if count, err := uc.boardEventRepo.GetUnreadCountBySeq(ctx, userID, boardIDs); err == nil {
			boardUnread = count
		}
	}

	directUnread, _ := uc.repo.GetUnreadCount(ctx, userID)
	total := boardUnread + directUnread

	// 3. Кэшируем в Redis (без TTL — инвалидируется через MarkRead)
	if uc.cache != nil {
		_ = uc.cache.Set(ctx, userID, total)
	}

	return total, nil
}
