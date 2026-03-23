package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateChecklistItemUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
}

func NewCreateChecklistItemUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository) *CreateChecklistItemUseCase {
	return &CreateChecklistItemUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
	}
}

func (uc *CreateChecklistItemUseCase) Execute(ctx context.Context, checklistID, boardID, userID, title string, position int) (*domain.ChecklistItem, error) {
	// 1. Проверка доступа (member может создавать элементы)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Создаем элемент (валидация внутри)
	item, err := domain.NewChecklistItem("", checklistID, boardID, title, position)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем
	if err := uc.checklistRepo.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}
