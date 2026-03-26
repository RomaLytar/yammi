package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type SetCustomFieldValueUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
	publisher       EventPublisher
}

func NewSetCustomFieldValueUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository, publisher EventPublisher) *SetCustomFieldValueUseCase {
	return &SetCustomFieldValueUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
		publisher:       publisher,
	}
}

func (uc *SetCustomFieldValueUseCase) Execute(ctx context.Context, cardID, boardID, fieldID, userID string, valueText *string, valueNumber *float64, valueDate *time.Time) (*domain.CustomFieldValue, error) {
	// 1. Проверка доступа (member может устанавливать значения)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверяем, что определение поля существует
	def, err := uc.customFieldRepo.GetDefinitionByID(ctx, fieldID)
	if err != nil {
		return nil, err
	}

	// 3. Валидируем значение по типу поля
	value := domain.NewCustomFieldValue("", cardID, boardID, fieldID)

	switch def.FieldType {
	case domain.FieldTypeText:
		if valueText == nil {
			return nil, domain.ErrInvalidFieldValue
		}
		value.SetText(*valueText)
	case domain.FieldTypeNumber:
		if valueNumber == nil {
			return nil, domain.ErrInvalidFieldValue
		}
		value.SetNumber(*valueNumber)
	case domain.FieldTypeDate:
		if valueDate == nil {
			return nil, domain.ErrInvalidFieldValue
		}
		value.SetDate(*valueDate)
	case domain.FieldTypeDropdown:
		if valueText == nil {
			return nil, domain.ErrInvalidFieldValue
		}
		// Проверяем, что значение есть в допустимых options
		found := false
		for _, opt := range def.Options {
			if opt == *valueText {
				found = true
				break
			}
		}
		if !found {
			return nil, domain.ErrInvalidFieldValue
		}
		value.SetText(*valueText)
	default:
		return nil, domain.ErrInvalidFieldType
	}

	// 4. Сохраняем (upsert)
	if err := uc.customFieldRepo.SetValue(ctx, value); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCustomFieldValueSet(ctx, CustomFieldValueSet{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			FieldID:      fieldID,
			CardID:       cardID,
			BoardID:      boardID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CustomFieldValueSet", "error", err, "field_id", fieldID, "card_id", cardID, "board_id", boardID)
		}
	}()

	return value, nil
}
