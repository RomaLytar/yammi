package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetChecklistsUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
}

func NewGetChecklistsUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository) *GetChecklistsUseCase {
	return &GetChecklistsUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
	}
}

func (uc *GetChecklistsUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.Checklist, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем чеклисты карточки (с элементами и прогрессом)
	return uc.checklistRepo.ListByCardID(ctx, cardID, boardID)
}
