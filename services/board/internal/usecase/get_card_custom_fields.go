package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardCustomFieldsUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
}

func NewGetCardCustomFieldsUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository) *GetCardCustomFieldsUseCase {
	return &GetCardCustomFieldsUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
	}
}

func (uc *GetCardCustomFieldsUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.CustomFieldValue, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем значения кастомных полей карточки
	return uc.customFieldRepo.GetCardValues(ctx, cardID, boardID)
}
