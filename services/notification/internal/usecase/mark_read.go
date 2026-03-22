package usecase

import (
	"context"
	"log"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type MarkReadUseCase struct {
	repo           NotificationRepository
	boardEventRepo BoardEventRepository
	unreadCounter  UnreadCounter
}

func NewMarkReadUseCase(
	repo NotificationRepository,
	boardEventRepo BoardEventRepository,
	unreadCounter UnreadCounter,
) *MarkReadUseCase {
	return &MarkReadUseCase{
		repo:           repo,
		boardEventRepo: boardEventRepo,
		unreadCounter:  unreadCounter,
	}
}

func (uc *MarkReadUseCase) Execute(ctx context.Context, userID string, ids []string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		// Проверяем, является ли ID board event
		boardID, err := uc.boardEventRepo.GetBoardIDByEventID(ctx, id)
		if err != nil {
			log.Printf("failed to check board event %s: %v", id, err)
			continue
		}

		if boardID != "" {
			// Board event — помечаем всю доску как прочитанную
			if err := uc.boardEventRepo.MarkBoardRead(ctx, userID, boardID); err != nil {
				log.Printf("failed to mark board %s as read for user %s: %v", boardID, userID, err)
				continue
			}
		} else {
			// Direct notification — помечаем конкретное уведомление
			if err := uc.repo.MarkAsRead(ctx, userID, []string{id}); err != nil {
				log.Printf("failed to mark notification %s as read for user %s: %v", id, userID, err)
				continue
			}
		}

		// Декрементируем Redis-счётчик
		if uc.unreadCounter != nil {
			if err := uc.unreadCounter.Decrement(ctx, userID); err != nil {
				log.Printf("failed to decrement unread counter for user %s: %v", userID, err)
			}
		}
	}

	return nil
}

type MarkAllReadUseCase struct {
	repo           NotificationRepository
	boardEventRepo BoardEventRepository
	memberRepo     BoardMemberRepository
	unreadCounter  UnreadCounter
}

func NewMarkAllReadUseCase(
	repo NotificationRepository,
	boardEventRepo BoardEventRepository,
	memberRepo BoardMemberRepository,
	unreadCounter UnreadCounter,
) *MarkAllReadUseCase {
	return &MarkAllReadUseCase{
		repo:           repo,
		boardEventRepo: boardEventRepo,
		memberRepo:     memberRepo,
		unreadCounter:  unreadCounter,
	}
}

func (uc *MarkAllReadUseCase) Execute(ctx context.Context, userID string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}

	// 1. Получаем доски пользователя
	boardIDs, err := uc.memberRepo.ListBoardIDsByUser(ctx, userID)
	if err != nil {
		log.Printf("failed to list board IDs for user %s: %v", userID, err)
	}

	// 2. Помечаем все board events как прочитанные
	if len(boardIDs) > 0 {
		if err := uc.boardEventRepo.MarkAllBoardsRead(ctx, userID, boardIDs); err != nil {
			log.Printf("failed to mark all boards read for user %s: %v", userID, err)
		}
	}

	// 3. Помечаем все direct notifications как прочитанные
	if err := uc.repo.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}

	// 4. Сбрасываем Redis-счётчик
	if uc.unreadCounter != nil {
		if err := uc.unreadCounter.Reset(ctx, userID); err != nil {
			log.Printf("failed to reset unread counter for user %s: %v", userID, err)
		}
	}

	return nil
}
