package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetBoardUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
}

func NewGetBoardUseCase(boardRepo BoardRepository, memberRepo MembershipRepository) *GetBoardUseCase {
	return &GetBoardUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetBoardUseCase) Execute(ctx context.Context, boardID, userID string) (*domain.Board, error) {
	// 1. Загружаем доску
	board, err := uc.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 2. Проверяем доступ (IsMember query)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	return board, nil
}
