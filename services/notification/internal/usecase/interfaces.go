package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

// NotificationRepository — интерфейс хранилища уведомлений.
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	BatchCreate(ctx context.Context, notifications []*domain.Notification) error
	ListByUserID(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error)
	MarkAsRead(ctx context.Context, userID string, ids []string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

// SettingsRepository — интерфейс хранилища настроек уведомлений.
type SettingsRepository interface {
	Get(ctx context.Context, userID string) (*domain.NotificationSettings, error)
	BatchGet(ctx context.Context, userIDs []string) (map[string]*domain.NotificationSettings, error)
	Upsert(ctx context.Context, settings *domain.NotificationSettings) error
}

// BoardMemberRepository — локальный кеш участников досок для маршрутизации уведомлений.
type BoardMemberRepository interface {
	AddMember(ctx context.Context, boardID, userID string) error
	RemoveMember(ctx context.Context, boardID, userID string) error
	RemoveAllByBoard(ctx context.Context, boardID string) error
	ListMemberIDs(ctx context.Context, boardID string) ([]string, error)
}

// EventPublisher — публикация событий для WebSocket доставки.
type EventPublisher interface {
	PublishNotificationCreated(ctx context.Context, n *domain.Notification) error
	PublishNotificationsBatch(ctx context.Context, notifications []*domain.Notification) error
}

// SettingsEventPublisher — публикация событий при изменении настроек.
type SettingsEventPublisher interface {
	PublishSettingsUpdated(ctx context.Context, userID string, enabled, realtimeEnabled bool) error
}
