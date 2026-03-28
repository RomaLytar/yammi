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
	// 1. Проверка лимита
	count, err := uc.userLabelRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= maxUserLabels {
		return nil, domain.ErrMaxUserLabelsReached
	}

	// 2. Создаем метку (валидация внутри)
	label, err := domain.NewUserLabel("", userID, name, color)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем
	if err := uc.userLabelRepo.Create(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}
