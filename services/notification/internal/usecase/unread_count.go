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
}

func NewGetUnreadCountUseCase(boardEventRepo BoardEventRepository, memberRepo BoardMemberRepository, repo NotificationRepository) *GetUnreadCountUseCase {
	return &GetUnreadCountUseCase{
		boardEventRepo: boardEventRepo,
		memberRepo:     memberRepo,
		repo:           repo,
	}
}

// Execute возвращает unread count через event_seq diff: O(1) на каждую доску.
// unread = sum(max_seq - last_seen_seq) по всем доскам пользователя + direct notifications.
func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}

	// 1. Board events: seq diff per board
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list boards for user %s: %v", userID, err)
		boardIDs = nil
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

	// 2. Direct notifications (welcome, member_added) — simple SQL COUNT
	directUnread, err := uc.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		log.Printf("failed to get direct unread count for user %s: %v", userID, err)
	}

	return boardUnread + directUnread, nil
}
