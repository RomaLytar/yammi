package usecase

import (
	"context"
	"log"
	"sync"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type GetUnreadCountUseCase struct {
	boardEventRepo BoardEventRepository
	memberRepo     BoardMemberRepository
	repo           NotificationRepository
	cache          UnreadCounter

	// Singleflight: предотвращает thundering herd при cache miss.
	// 1000 concurrent miss для одного userID → 1 SQL query, остальные ждут.
	mu      sync.Mutex
	inflight map[string]chan struct{}
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
		inflight:       make(map[string]chan struct{}),
	}
}

func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}

	// 1. Redis cache hit = O(1)
	if uc.cache != nil {
		if cached, err := uc.cache.Get(ctx, userID); err == nil && cached >= 0 {
			return cached, nil
		}
	}

	// 2. Singleflight: если уже вычисляем для этого user — ждём
	uc.mu.Lock()
	if ch, ok := uc.inflight[userID]; ok {
		uc.mu.Unlock()
		<-ch // ждём завершения другого запроса
		// Теперь результат в Redis cache
		if uc.cache != nil {
			if cached, err := uc.cache.Get(ctx, userID); err == nil && cached >= 0 {
				return cached, nil
			}
		}
		return 0, nil
	}
	ch := make(chan struct{})
	uc.inflight[userID] = ch
	uc.mu.Unlock()

	defer func() {
		uc.mu.Lock()
		delete(uc.inflight, userID)
		close(ch) // разблокируем ожидающих
		uc.mu.Unlock()
	}()

	// 3. Вычисляем: Redis MGET (board seqs) + SQL (user cursors)
	total := uc.computeUnread(ctx, userID)

	// 4. Кэшируем в Redis
	if uc.cache != nil {
		_ = uc.cache.Set(ctx, userID, total)
	}

	return total, nil
}

func (uc *GetUnreadCountUseCase) computeUnread(ctx context.Context, userID string) int {
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list boards for user %s: %v", userID, err)
		return 0
	}

	boardUnread := 0
	if len(boardIDs) > 0 {
		count, err := uc.boardEventRepo.GetUnreadCountBySeq(ctx, userID, boardIDs)
		if err != nil {
			log.Printf("failed to get board unread count for user %s: %v", userID, err)
		} else {
			boardUnread = count
		}
	}

	directUnread, _ := uc.repo.GetUnreadCount(ctx, userID)
	return boardUnread + directUnread
}
