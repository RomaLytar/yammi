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

// ExecuteAuthorized загружает карточки без проверки доступа.
// Вызывать только когда доступ уже проверен (например, в MoveCard handler после moveCard.Execute).
func (uc *GetCardsUseCase) ExecuteAuthorized(ctx context.Context, columnID string) ([]*domain.Card, error) {
	return uc.cardRepo.ListByColumnID(ctx, columnID)
}
