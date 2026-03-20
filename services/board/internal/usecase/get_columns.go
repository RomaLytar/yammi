package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetColumnsUseCase struct {
	columnRepo ColumnRepository
	memberRepo MembershipRepository
}

func NewGetColumnsUseCase(columnRepo ColumnRepository, memberRepo MembershipRepository) *GetColumnsUseCase {
	return &GetColumnsUseCase{
		columnRepo: columnRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetColumnsUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.Column, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем колонки
	return uc.columnRepo.ListByBoardID(ctx, boardID)
}
