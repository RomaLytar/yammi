package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteLabelUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteLabelUseCase(labelRepo LabelRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteLabelUseCase {
	return &DeleteLabelUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteLabelUseCase) Execute(ctx context.Context, labelID, boardID, userID string) error {
	// 1. Проверка доступа (только owner может удалять метки)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return domain.ErrNotOwner
	}

	// 2. Удаляем метку (CASCADE удалит card_labels)
	if err := uc.labelRepo.Delete(ctx, labelID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishLabelDeleted(context.Background(), LabelDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LabelID:      labelID,
			BoardID:      boardID,
			ActorID:      userID,
		})
	}()

	return nil
}
