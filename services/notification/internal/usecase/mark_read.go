package usecase

import (
	"context"
	"log"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type MarkReadUseCase struct {
	repo           NotificationRepository
	boardEventRepo BoardEventRepository
	cache          UnreadCounter
}

func NewMarkReadUseCase(repo NotificationRepository, boardEventRepo BoardEventRepository, cache UnreadCounter) *MarkReadUseCase {
	return &MarkReadUseCase{repo: repo, boardEventRepo: boardEventRepo, cache: cache}
}

func (uc *MarkReadUseCase) Execute(ctx context.Context, userID string, ids []string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		boardID, err := uc.boardEventRepo.GetBoardIDByEventID(ctx, id)
		if err != nil {
			log.Printf("failed to check board event %s: %v", id, err)
			continue
		}

		if boardID != "" {
			if err := uc.boardEventRepo.MarkBoardRead(ctx, userID, boardID); err != nil {
				log.Printf("failed to mark board %s as read: %v", boardID, err)
			}
		} else {
			if err := uc.repo.MarkAsRead(ctx, userID, []string{id}); err != nil {
				log.Printf("failed to mark notification %s as read: %v", id, err)
			}
		}
	}

	// Инвалидируем Redis cache — следующий GetUnreadCount пересчитает из SQL
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, userID)
	}

	return nil
}

type MarkAllReadUseCase struct {
	repo           NotificationRepository
	boardEventRepo BoardEventRepository
	memberRepo     BoardMemberRepository
	cache          UnreadCounter
}

func NewMarkAllReadUseCase(repo NotificationRepository, boardEventRepo BoardEventRepository, memberRepo BoardMemberRepository, cache UnreadCounter) *MarkAllReadUseCase {
	return &MarkAllReadUseCase{repo: repo, boardEventRepo: boardEventRepo, memberRepo: memberRepo, cache: cache}
}

func (uc *MarkAllReadUseCase) Execute(ctx context.Context, userID string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}

	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list board IDs for user %s: %v", userID, err)
	}

	if len(boardIDs) > 0 {
		if err := uc.boardEventRepo.MarkAllBoardsRead(ctx, userID, boardIDs); err != nil {
			log.Printf("failed to mark all boards read for user %s: %v", userID, err)
		}
	}

	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}

	// Инвалидируем Redis cache
	if uc.cache != nil {
		_ = uc.cache.Invalidate(ctx, userID)
	}

	return nil
}
