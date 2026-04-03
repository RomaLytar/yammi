package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetReleaseCardsUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
}

func NewGetReleaseCardsUseCase(cardRepo CardRepository, memberRepo MembershipRepository) *GetReleaseCardsUseCase {
	return &GetReleaseCardsUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetReleaseCardsUseCase) Execute(ctx context.Context, boardID, releaseID, userID string) ([]*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем карточки релиза
	return uc.cardRepo.ListByReleaseID(ctx, boardID, releaseID)
}
