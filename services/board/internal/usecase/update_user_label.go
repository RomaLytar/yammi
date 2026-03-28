package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateUserLabelUseCase struct {
	userLabelRepo UserLabelRepository
}

func NewUpdateUserLabelUseCase(userLabelRepo UserLabelRepository) *UpdateUserLabelUseCase {
	return &UpdateUserLabelUseCase{
		userLabelRepo: userLabelRepo,
	}
}

func (uc *UpdateUserLabelUseCase) Execute(ctx context.Context, labelID, userID, name, color string) (*domain.UserLabel, error) {
	// 1. Загружаем метку
	label, err := uc.userLabelRepo.GetByID(ctx, labelID)
	if err != nil {
		return nil, err
	}

	// 2. Проверка владения (userID должен совпадать)
	if label.UserID != userID {
		return nil, domain.ErrAccessDenied
	}

	// 3. Обновляем (валидация внутри)
	if err := label.Update(name, color); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.userLabelRepo.Update(ctx, label); err != nil {
		return nil, err
	}

	return label, nil
}
