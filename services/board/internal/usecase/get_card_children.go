package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardChildrenUseCase struct {
	cardLinkRepo CardLinkRepository
	memberRepo   MembershipRepository
}

func NewGetCardChildrenUseCase(cardLinkRepo CardLinkRepository, memberRepo MembershipRepository) *GetCardChildrenUseCase {
	return &GetCardChildrenUseCase{
		cardLinkRepo: cardLinkRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *GetCardChildrenUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.CardLink, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем дочерние связи
	return uc.cardLinkRepo.ListChildren(ctx, cardID, boardID)
}
