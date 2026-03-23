package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListLabelsUseCase struct {
	labelRepo  LabelRepository
	memberRepo MembershipRepository
}

func NewListLabelsUseCase(labelRepo LabelRepository, memberRepo MembershipRepository) *ListLabelsUseCase {
	return &ListLabelsUseCase{
		labelRepo:  labelRepo,
		memberRepo: memberRepo,
	}
}

func (uc *ListLabelsUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.Label, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем метки доски
	return uc.labelRepo.ListByBoardID(ctx, boardID)
}
