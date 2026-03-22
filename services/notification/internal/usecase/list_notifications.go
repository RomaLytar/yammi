package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

type ListNotificationsUseCase struct {
	repo NotificationRepository
}

func NewListNotificationsUseCase(repo NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: repo}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, int, error) {
	if userID == "" {
		return nil, "", 0, domain.ErrEmptyUserID
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	notifications, nextCursor, err := uc.repo.ListByUserID(ctx, userID, limit, cursor, typeFilter, search)
	if err != nil {
		return nil, "", 0, err
	}

	unreadCount, err := uc.repo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, "", 0, err
	}

	return notifications, nextCursor, unreadCount, nil
}
