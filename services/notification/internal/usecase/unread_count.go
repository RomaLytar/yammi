package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

type GetUnreadCountUseCase struct {
	repo NotificationRepository
}

func NewGetUnreadCountUseCase(repo NotificationRepository) *GetUnreadCountUseCase {
	return &GetUnreadCountUseCase{repo: repo}
}

func (uc *GetUnreadCountUseCase) Execute(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, domain.ErrEmptyUserID
	}
	return uc.repo.GetUnreadCount(ctx, userID)
}
