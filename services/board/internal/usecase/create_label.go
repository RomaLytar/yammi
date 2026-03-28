package usecase

import (
	"context"
	"log/slog"
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

	// 2. Создаем метку (валидация внутри)
	label, err := domain.NewLabel("", boardID, name, color)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем с проверкой лимита в одном запросе (вместо COUNT + INSERT)
	if err := uc.labelRepo.CreateWithLimit(ctx, label, maxLabelsPerBoard); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishLabelCreated(ctx, LabelCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LabelID:      label.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         label.Name,
			Color:        label.Color,
		}); err != nil {
			slog.Error("failed to publish LabelCreated", "error", err, "label_id", label.ID, "board_id", boardID)
		}
	}()

	return label, nil
}
