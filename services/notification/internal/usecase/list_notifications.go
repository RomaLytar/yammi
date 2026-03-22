package usecase

import (
	"context"
	"log"
	"sort"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type ListNotificationsUseCase struct {
	repo           NotificationRepository
	boardEventRepo BoardEventRepository
	memberRepo     BoardMemberRepository
	unreadCounter  UnreadCounter
}

func NewListNotificationsUseCase(
	repo NotificationRepository,
	boardEventRepo BoardEventRepository,
	memberRepo BoardMemberRepository,
	unreadCounter UnreadCounter,
) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{
		repo:           repo,
		boardEventRepo: boardEventRepo,
		memberRepo:     memberRepo,
		unreadCounter:  unreadCounter,
	}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, int, error) {
	if userID == "" {
		return nil, "", 0, domain.ErrEmptyUserID
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 1. Получаем доски пользователя
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list board IDs for user %s: %v", userID, err)
		boardIDs = nil
	}

	// 2. Запрашиваем board events (если есть доски)
	var boardEvents []*domain.Notification
	var boardCursor string
	if len(boardIDs) > 0 {
		boardEvents, boardCursor, err = uc.boardEventRepo.ListForUser(ctx, userID, boardIDs, limit, cursor, typeFilter, search)
		if err != nil {
			return nil, "", 0, err
		}
	}

	// 3. Запрашиваем direct notifications (welcome, member_added/removed)
	directNotifications, directCursor, err := uc.repo.ListByUserID(ctx, userID, limit, cursor, typeFilter, search)
	if err != nil {
		return nil, "", 0, err
	}

	// 4. Мержим оба списка по created_at DESC
	merged := make([]*domain.Notification, 0, len(boardEvents)+len(directNotifications))
	merged = append(merged, boardEvents...)
	merged = append(merged, directNotifications...)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].CreatedAt.After(merged[j].CreatedAt)
	})

	// Обрезаем до limit
	var nextCursor string
	if len(merged) > limit {
		merged = merged[:limit]
	}

	// Определяем nextCursor — берём более ранний из двух курсоров
	if boardCursor != "" && directCursor != "" {
		if boardCursor < directCursor {
			nextCursor = boardCursor
		} else {
			nextCursor = directCursor
		}
	} else if boardCursor != "" {
		nextCursor = boardCursor
	} else {
		nextCursor = directCursor
	}

	// 5. Получаем unread count из Redis
	unreadCount, err := uc.unreadCounter.Get(ctx, userID)
	if err != nil {
		log.Printf("failed to get unread count from Redis for user %s: %v", userID, err)
		unreadCount = 0
	}

	return merged, nextCursor, unreadCount, nil
}
