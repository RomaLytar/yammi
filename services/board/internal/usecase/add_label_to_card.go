package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AddLabelToCardUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewAddLabelToCardUseCase(labelRepo LabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *AddLabelToCardUseCase {
	return &AddLabelToCardUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *AddLabelToCardUseCase) Execute(ctx context.Context, cardID, boardID, labelID, userID string) error {
	// 1. Проверка доступа (member может назначать метки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Назначаем метку на карточку
	if err := uc.labelRepo.AddToCard(ctx, cardID, boardID, labelID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardLabelAdded(ctx, CardLabelAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			CardID:       cardID,
			BoardID:      boardID,
			LabelID:      labelID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardLabelAdded", "error", err, "card_id", cardID, "board_id", boardID)
		}
	}()

	return nil
}
