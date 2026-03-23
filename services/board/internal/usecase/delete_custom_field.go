package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteCustomFieldUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
	publisher       EventPublisher
}

func NewDeleteCustomFieldUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteCustomFieldUseCase {
	return &DeleteCustomFieldUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
		publisher:       publisher,
	}
}

func (uc *DeleteCustomFieldUseCase) Execute(ctx context.Context, fieldID, boardID, userID string) error {
	// 1. Проверка доступа (только owner может удалять определения)
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

	// 2. Удаляем определение (CASCADE удалит значения)
	if err := uc.customFieldRepo.DeleteDefinition(ctx, fieldID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishCustomFieldDeleted(context.Background(), CustomFieldDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			FieldID:      fieldID,
			BoardID:      boardID,
			ActorID:      userID,
		})
	}()

	return nil
}
