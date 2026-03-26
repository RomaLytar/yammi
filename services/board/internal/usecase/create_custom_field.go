package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

const maxCustomFieldsPerBoard = 30

type CreateCustomFieldUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
	publisher       EventPublisher
}

func NewCreateCustomFieldUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateCustomFieldUseCase {
	return &CreateCustomFieldUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
		publisher:       publisher,
	}
}

func (uc *CreateCustomFieldUseCase) Execute(ctx context.Context, boardID, userID, name string, fieldType domain.FieldType, options []string, position int, required bool) (*domain.CustomFieldDefinition, error) {
	// 1. Проверка доступа (только owner может создавать кастомные поля)
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

	// 2. Проверка лимита кастомных полей на доску
	count, err := uc.customFieldRepo.CountDefinitionsByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if count >= maxCustomFieldsPerBoard {
		return nil, domain.ErrMaxCustomFieldsReached
	}

	// 3. Создаем определение (валидация внутри)
	def, err := domain.NewCustomFieldDefinition("", boardID, name, fieldType, options, position, required)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.customFieldRepo.CreateDefinition(ctx, def); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCustomFieldCreated(ctx, CustomFieldCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			FieldID:      def.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         def.Name,
			FieldType:    string(def.FieldType),
		}); err != nil {
			slog.Error("failed to publish CustomFieldCreated", "error", err, "field_id", def.ID, "board_id", boardID)
		}
	}()

	return def, nil
}
