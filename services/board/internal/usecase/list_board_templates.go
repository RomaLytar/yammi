package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListBoardTemplatesUseCase struct {
	boardTmplRepo BoardTemplateRepository
}

func NewListBoardTemplatesUseCase(boardTmplRepo BoardTemplateRepository) *ListBoardTemplatesUseCase {
	return &ListBoardTemplatesUseCase{
		boardTmplRepo: boardTmplRepo,
	}
}

func (uc *ListBoardTemplatesUseCase) Execute(ctx context.Context, userID string) ([]*domain.BoardTemplate, error) {
	return uc.boardTmplRepo.ListByUserID(ctx, userID)
}
