package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListUserLabelsUseCase struct {
	userLabelRepo UserLabelRepository
}

func NewListUserLabelsUseCase(userLabelRepo UserLabelRepository) *ListUserLabelsUseCase {
	return &ListUserLabelsUseCase{
		userLabelRepo: userLabelRepo,
	}
}

func (uc *ListUserLabelsUseCase) Execute(ctx context.Context, userID string) ([]*domain.UserLabel, error) {
	return uc.userLabelRepo.ListByUserID(ctx, userID)
}
