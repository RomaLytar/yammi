package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetBacklogUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
}

func NewGetBacklogUseCase(cardRepo CardRepository, memberRepo MembershipRepository) *GetBacklogUseCase {
	return &GetBacklogUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetBacklogUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем бэклог
	return uc.cardRepo.ListBacklog(ctx, boardID)
}
