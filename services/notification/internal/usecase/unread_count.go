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

// Execute: Redis GET (cache hit) → Redis MGET + SQL cursors (cache miss) → SET.
// Быстрый path: Redis cache hit = O(1). Медленный: MGET + 1 SQL query.
// При ошибках возвращает 0 (eventual consistency), не таймаутит.
func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}

	// 1. Redis cache (hit = instant, miss = SQL fallback)
	if uc.cache != nil {
		if cached, err := uc.cache.Get(ctx, userID); err == nil && cached >= 0 {
			return cached, nil
		}
	}


	// 2. Cache miss → вычисляем: Redis MGET (board seqs) + SQL (user cursors)
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list boards for user %s: %v", userID, err)
		boardIDs = nil
	}

	boardUnread := 0
	if len(boardIDs) > 0 && uc.cache != nil {
		// Redis MGET: max_seq per board (1 round-trip, O(1))
		boardSeqs, _ := uc.cache.GetBoardSeqs(ctx, boardIDs)
		if len(boardSeqs) > 0 {
			// SQL: user cursors only (1 lightweight query, no board_events scan)
			cursors, _ := uc.boardEventRepo.GetUserCursors(ctx, userID, boardIDs)
			for boardID, maxSeq := range boardSeqs {
				lastSeen := cursors[boardID]
				if diff := maxSeq - lastSeen; diff > 0 {
					boardUnread += int(diff)
				}
			}
		}
		// Redis miss для доски = 0 unread (eventually consistent)
		// Следующий event на доске SET'ит board_seq в Redis
	}

	directUnread, _ := uc.repo.GetUnreadCount(ctx, userID)
	total := boardUnread + directUnread

	// 3. Кэшируем в Redis (без TTL — инвалидируется через MarkRead)
	if uc.cache != nil {
		_ = uc.cache.Set(ctx, userID, total)
	}

	return total, nil
}
