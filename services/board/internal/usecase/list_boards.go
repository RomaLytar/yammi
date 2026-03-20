package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListBoardsUseCase struct {
	boardRepo BoardRepository
}

func NewListBoardsUseCase(boardRepo BoardRepository) *ListBoardsUseCase {
	return &ListBoardsUseCase{
		boardRepo: boardRepo,
	}
}

func (uc *ListBoardsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string) ([]*domain.Board, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default
	}

	return uc.boardRepo.ListByUserID(ctx, userID, limit, cursor)
}
