package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteCardUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteCardUseCase(cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteCardUseCase {
	return &DeleteCardUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, cardID, boardID, columnID, userID string) error {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Удаляем
	if err := uc.cardRepo.Delete(ctx, cardID); err != nil {
		return err
	}

	// 3. Публикуем событие
	go func() {
		_ = uc.publisher.PublishCardDeleted(context.Background(), CardDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			CardID:       cardID,
			ColumnID:     columnID,
		})
	}()

	return nil
}
