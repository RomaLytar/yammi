package usecase

import (
	"context"
	"log/slog"
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

	// 2. Удаляем определение (CASCADE удалит значения, с проверкой boardID)
	if err := uc.customFieldRepo.DeleteDefinition(ctx, fieldID, boardID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCustomFieldDeleted(ctx, CustomFieldDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			FieldID:      fieldID,
			BoardID:      boardID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CustomFieldDeleted", "error", err, "field_id", fieldID, "board_id", boardID)
		}
	}()

	return nil
}
