package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

const maxUserLabels = 50

type CreateUserLabelUseCase struct {
	userLabelRepo UserLabelRepository
}

func NewCreateUserLabelUseCase(userLabelRepo UserLabelRepository) *CreateUserLabelUseCase {
	return &CreateUserLabelUseCase{
		userLabelRepo: userLabelRepo,
	}
}

func (uc *CreateUserLabelUseCase) Execute(ctx context.Context, userID, name, color string) (*domain.UserLabel, error) {
	// 1. Создаем метку (валидация внутри)
	label, err := domain.NewUserLabel("", userID, name, color)
	if err != nil {
		return nil, err
	}

	// 2. Сохраняем с проверкой лимита в одном запросе (вместо COUNT + INSERT)
	if err := uc.userLabelRepo.CreateWithLimit(ctx, label, maxUserLabels); err != nil {
		return nil, err
	}

	return label, nil
}
