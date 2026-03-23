package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteChecklistItemUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
}

func NewDeleteChecklistItemUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository) *DeleteChecklistItemUseCase {
	return &DeleteChecklistItemUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
	}
}

func (uc *DeleteChecklistItemUseCase) Execute(ctx context.Context, itemID, boardID, userID string) error {
	// 1. Проверка доступа (member может удалять элементы)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Удаляем элемент
	return uc.checklistRepo.DeleteItem(ctx, itemID, boardID)
}
