package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteUserLabelUseCase struct {
	userLabelRepo UserLabelRepository
}

func NewDeleteUserLabelUseCase(userLabelRepo UserLabelRepository) *DeleteUserLabelUseCase {
	return &DeleteUserLabelUseCase{
		userLabelRepo: userLabelRepo,
	}
}

func (uc *DeleteUserLabelUseCase) Execute(ctx context.Context, labelID, userID string) error {
	// 1. Загружаем метку
	label, err := uc.userLabelRepo.GetByID(ctx, labelID)
	if err != nil {
		return err
	}

	// 2. Проверка владения
	if label.UserID != userID {
		return domain.ErrAccessDenied
	}

	// 3. Удаляем
	return uc.userLabelRepo.Delete(ctx, labelID)
}
