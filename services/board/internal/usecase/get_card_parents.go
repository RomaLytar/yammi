package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardParentsUseCase struct {
	cardLinkRepo CardLinkRepository
	memberRepo   MembershipRepository
}

func NewGetCardParentsUseCase(cardLinkRepo CardLinkRepository, memberRepo MembershipRepository) *GetCardParentsUseCase {
	return &GetCardParentsUseCase{
		cardLinkRepo: cardLinkRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *GetCardParentsUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.CardLink, error) {
	// 1. Проверка доступа (member доски может видеть родительские связи)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем родительские связи (без boardID — child может быть на любой доске)
	return uc.cardLinkRepo.ListParents(ctx, cardID)
}
