package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
}

func NewGetCardUseCase(cardRepo CardRepository, memberRepo MembershipRepository) *GetCardUseCase {
	return &GetCardUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetCardUseCase) Execute(ctx context.Context, cardID, boardID, userID string) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточку
	return uc.cardRepo.GetByID(ctx, cardID)
}
