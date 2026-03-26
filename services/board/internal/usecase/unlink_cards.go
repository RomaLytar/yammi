package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UnlinkCardsUseCase struct {
	cardLinkRepo CardLinkRepository
	memberRepo   MembershipRepository
	publisher    EventPublisher
}

func NewUnlinkCardsUseCase(cardLinkRepo CardLinkRepository, memberRepo MembershipRepository, publisher EventPublisher) *UnlinkCardsUseCase {
	return &UnlinkCardsUseCase{
		cardLinkRepo: cardLinkRepo,
		memberRepo:   memberRepo,
		publisher:    publisher,
	}
}

func (uc *UnlinkCardsUseCase) Execute(ctx context.Context, linkID, boardID, userID string) error {
	// 1. Проверка доступа (member может удалять связи)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Получаем связь для события
	link, err := uc.cardLinkRepo.GetByID(ctx, linkID, boardID)
	if err != nil {
		return err
	}

	// 3. Удаляем связь
	if err := uc.cardLinkRepo.Delete(ctx, linkID, boardID); err != nil {
		return err
	}

	// 4. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardUnlinked(ctx, CardUnlinked{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LinkID:       linkID,
			ParentID:     link.ParentID,
			ChildID:      link.ChildID,
			BoardID:      boardID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardUnlinked", "error", err, "link_id", linkID, "board_id", boardID)
		}
	}()

	return nil
}
