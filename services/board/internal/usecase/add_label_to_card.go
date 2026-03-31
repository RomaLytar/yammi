package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AddLabelToCardUseCase struct {
	labelRepo     LabelRepository
	userLabelRepo UserLabelRepository
	memberRepo    MembershipRepository
	publisher     EventPublisher
}

func NewAddLabelToCardUseCase(labelRepo LabelRepository, userLabelRepo UserLabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *AddLabelToCardUseCase {
	return &AddLabelToCardUseCase{
		labelRepo:     labelRepo,
		userLabelRepo: userLabelRepo,
		memberRepo:    memberRepo,
		publisher:     publisher,
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

	// 2. Проверяем, есть ли метка в board labels (с проверкой boardID)
	_, err = uc.labelRepo.GetByID(ctx, labelID, boardID)
	if err != nil {
		// Метка не найдена в board labels — проверяем глобальные метки
		userLabel, ulErr := uc.userLabelRepo.GetByID(ctx, labelID)
		if ulErr != nil {
			return domain.ErrLabelNotFound
		}

		// Копируем глобальную метку как метку доски (сохраняем тот же ID)
		boardLabel, createErr := domain.NewLabel(userLabel.ID, boardID, userLabel.Name, userLabel.Color)
		if createErr != nil {
			return createErr
		}
		if createErr = uc.labelRepo.Create(ctx, boardLabel); createErr != nil {
			return createErr
		}
		labelID = boardLabel.ID
	}

	// 3. Назначаем метку на карточку
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
