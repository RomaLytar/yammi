package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

const maxLabelsPerBoard = 50

type CreateLabelUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewCreateLabelUseCase(labelRepo LabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateLabelUseCase {
	return &CreateLabelUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *CreateLabelUseCase) Execute(ctx context.Context, boardID, userID, name, color string) (*domain.Label, error) {
	// 1. Проверка доступа (member может создавать метки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверка лимита меток на доску
	count, err := uc.labelRepo.CountByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if count >= maxLabelsPerBoard {
		return nil, domain.ErrMaxLabelsReached
	}

	// 3. Создаем метку (валидация внутри)
	label, err := domain.NewLabel("", boardID, name, color)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.labelRepo.Create(ctx, label); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishLabelCreated(context.Background(), LabelCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LabelID:      label.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         label.Name,
			Color:        label.Color,
		})
	}()

	return label, nil
}
