package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetCardLabelsUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
}

func NewGetCardLabelsUseCase(labelRepo LabelRepository, memberRepo MembershipRepository) *GetCardLabelsUseCase {
	return &GetCardLabelsUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetCardLabelsUseCase) Execute(ctx context.Context, cardID, boardID, userID string) ([]*domain.Label, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем метки карточки
	return uc.labelRepo.ListByCardID(ctx, cardID, boardID)
}
