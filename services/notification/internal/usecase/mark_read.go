package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

type MarkReadUseCase struct {
	repo NotificationRepository
}

func NewMarkReadUseCase(repo NotificationRepository) *MarkReadUseCase {
	return &MarkReadUseCase{repo: repo}
}

func (uc *MarkReadUseCase) Execute(ctx context.Context, userID string, ids []string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}
	if len(ids) == 0 {
		return nil
	}
	return uc.repo.MarkAsRead(ctx, userID, ids)
}

type MarkAllReadUseCase struct {
	repo NotificationRepository
}

func NewMarkAllReadUseCase(repo NotificationRepository) *MarkAllReadUseCase {
	return &MarkAllReadUseCase{repo: repo}
}

func (uc *MarkAllReadUseCase) Execute(ctx context.Context, userID string) error {
	if userID == "" {
		return domain.ErrEmptyUserID
	}
	return uc.repo.MarkAllAsRead(ctx, userID)
}
