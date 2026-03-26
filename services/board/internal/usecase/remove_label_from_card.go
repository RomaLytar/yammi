package usecase

import (
	"context"
	"log/slog"
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardLabelRemoved(ctx, CardLabelRemoved{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			CardID:       cardID,
			BoardID:      boardID,
			LabelID:      labelID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardLabelRemoved", "error", err, "card_id", cardID, "board_id", boardID)
		}
	}()

	return nil
}
