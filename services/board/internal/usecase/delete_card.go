package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteCardUseCase struct {
	cardRepo   CardRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteCardUseCase {
	return &DeleteCardUseCase{
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, cardIDs []string, boardID, userID string) error {
	// 1. Проверка доступа
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Если не owner, проверяем что каждая карточка принадлежит пользователю
	if role != domain.RoleOwner {
		for _, cardID := range cardIDs {
			card, err := uc.cardRepo.GetByID(ctx, cardID)
			if err != nil {
				return err
			}
			if card.CreatorID != userID {
				return domain.ErrAccessDenied
			}
		}
	}

	// 3. Batch delete
	if err := uc.cardRepo.BatchDelete(ctx, boardID, cardIDs); err != nil {
		return err
	}

	// 4. Обновляем updated_at доски
	_ = uc.boardRepo.TouchUpdatedAt(ctx, boardID)

	// 5. Публикуем события
	go func() {
		for _, cardID := range cardIDs {
			_ = uc.publisher.PublishCardDeleted(context.Background(), CardDeleted{
				EventID:      generateEventID(),
				EventVersion: 1,
				OccurredAt:   getCurrentTime(),
				BoardID:      boardID,
				CardID:       cardID,
				ActorID:      userID,
			})
		}
	}()

	return nil
}
