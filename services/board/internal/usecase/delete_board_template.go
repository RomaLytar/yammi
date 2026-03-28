package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteBoardTemplateUseCase struct {
	boardTmplRepo BoardTemplateRepository
}

func NewDeleteBoardTemplateUseCase(boardTmplRepo BoardTemplateRepository) *DeleteBoardTemplateUseCase {
	return &DeleteBoardTemplateUseCase{
		boardTmplRepo: boardTmplRepo,
	}
}

func (uc *DeleteBoardTemplateUseCase) Execute(ctx context.Context, templateID, userID string) error {
	// 1. Получаем шаблон для проверки ownership
	tmpl, err := uc.boardTmplRepo.GetByID(ctx, templateID)
	if err != nil {
		return err
	}

	// 2. Только создатель шаблона может удалить
	if tmpl.UserID != userID {
		return domain.ErrAccessDenied
	}

	// 3. Удаляем
	return uc.boardTmplRepo.Delete(ctx, templateID)
}
