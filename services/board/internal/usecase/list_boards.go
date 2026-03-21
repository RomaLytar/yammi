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

func (uc *ListBoardsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string, ownerOnly bool, search string, sortBy string) ([]*domain.Board, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // default
	}

	// Validate sortBy
	switch sortBy {
	case "created_at", "title":
		// valid
	default:
		sortBy = "updated_at"
	}

	return uc.boardRepo.ListByUserID(ctx, userID, limit, cursor, ownerOnly, search, sortBy)
}
