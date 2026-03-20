package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardsUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
}

func NewGetCardsUseCase(cardRepo CardRepository, memberRepo MembershipRepository) *GetCardsUseCase {
	return &GetCardsUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetCardsUseCase) Execute(ctx context.Context, columnID, boardID, userID string) ([]*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточки
	return uc.cardRepo.ListByColumnID(ctx, columnID)
}
