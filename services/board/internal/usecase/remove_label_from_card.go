package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type RemoveLabelFromCardUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewRemoveLabelFromCardUseCase(labelRepo LabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *RemoveLabelFromCardUseCase {
	return &RemoveLabelFromCardUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *RemoveLabelFromCardUseCase) Execute(ctx context.Context, cardID, boardID, labelID, userID string) error {
	// 1. Проверка доступа (member может снимать метки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Снимаем метку с карточки
	if err := uc.labelRepo.RemoveFromCard(ctx, cardID, boardID, labelID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishCardLabelRemoved(context.Background(), CardLabelRemoved{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			CardID:       cardID,
			BoardID:      boardID,
			LabelID:      labelID,
			ActorID:      userID,
		})
	}()

	return nil
}
