package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateLabelUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateLabelUseCase(labelRepo LabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateLabelUseCase {
	return &UpdateLabelUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *UpdateLabelUseCase) Execute(ctx context.Context, labelID, boardID, userID, name, color string) (*domain.Label, error) {
	// 1. Проверка доступа (member может обновлять метки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем метку
	label, err := uc.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем (валидация внутри)
	if err := label.Update(name, color); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.labelRepo.Update(ctx, label); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishLabelUpdated(context.Background(), LabelUpdated{
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
