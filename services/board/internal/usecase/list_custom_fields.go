package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListCustomFieldsUseCase struct {
	customFieldRepo CustomFieldRepository
	memberRepo      MembershipRepository
}

func NewListCustomFieldsUseCase(customFieldRepo CustomFieldRepository, memberRepo MembershipRepository) *ListCustomFieldsUseCase {
	return &ListCustomFieldsUseCase{
		customFieldRepo: customFieldRepo,
		memberRepo:      memberRepo,
	}
}

func (uc *ListCustomFieldsUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.CustomFieldDefinition, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем определения кастомных полей доски
	return uc.customFieldRepo.ListDefinitionsByBoardID(ctx, boardID)
}
