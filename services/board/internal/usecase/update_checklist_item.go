package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateChecklistItemUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
}

func NewUpdateChecklistItemUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository) *UpdateChecklistItemUseCase {
	return &UpdateChecklistItemUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
	}
}

func (uc *UpdateChecklistItemUseCase) Execute(ctx context.Context, itemID, boardID, userID, title string) (*domain.ChecklistItem, error) {
	// 1. Проверка доступа (member может обновлять элементы)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем элемент
	item, err := uc.checklistRepo.GetItemByID(ctx, itemID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем (валидация внутри)
	if err := item.Update(title); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.checklistRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}
