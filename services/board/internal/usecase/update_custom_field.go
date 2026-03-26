package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateCustomFieldUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
	publisher       EventPublisher
}

func NewUpdateCustomFieldUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateCustomFieldUseCase {
	return &UpdateCustomFieldUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
		publisher:       publisher,
	}
}

func (uc *UpdateCustomFieldUseCase) Execute(ctx context.Context, fieldID, boardID, userID, name string, options []string, required bool) (*domain.CustomFieldDefinition, error) {
	// 1. Проверка доступа (только owner может обновлять определения)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return nil, domain.ErrNotOwner
	}

	// 2. Загружаем определение
	def, err := uc.customFieldRepo.GetDefinitionByID(ctx, fieldID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем (валидация внутри)
	if err := def.Update(name, options, required); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.customFieldRepo.UpdateDefinition(ctx, def); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCustomFieldUpdated(ctx, CustomFieldUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			FieldID:      def.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         def.Name,
		}); err != nil {
			slog.Error("failed to publish CustomFieldUpdated", "error", err, "field_id", def.ID, "board_id", boardID)
		}
	}()

	return def, nil
}
