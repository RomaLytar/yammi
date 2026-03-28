package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateBoardTemplateUseCase struct {
	boardTmplRepo BoardTemplateRepository
}

func NewCreateBoardTemplateUseCase(boardTmplRepo BoardTemplateRepository) *CreateBoardTemplateUseCase {
	return &CreateBoardTemplateUseCase{
		boardTmplRepo: boardTmplRepo,
	}
}

func (uc *CreateBoardTemplateUseCase) Execute(ctx context.Context, userID, name, description string, columnsData []domain.BoardColumnTemplateData, labelsData []domain.LabelTemplateData) (*domain.BoardTemplate, error) {
	// 1. Создаем шаблон (валидация внутри)
	tmpl, err := domain.NewBoardTemplate("", userID, name, description, columnsData, labelsData)
	if err != nil {
		return nil, err
	}

	// 2. Сохраняем
	if err := uc.boardTmplRepo.Create(ctx, tmpl); err != nil {
		return nil, err
	}

	return tmpl, nil
}
