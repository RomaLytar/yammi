package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
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
	ListBoardIDsByUser(ctx context.Context, userID string) ([]string, error)
	TruncateCache(ctx context.Context) error
}

// BoardEventRepository — хранилище board events (event-sourcing вместо fan-out).
type BoardEventRepository interface {
	Create(ctx context.Context, event *domain.BoardEvent) error
	ListForUser(ctx context.Context, userID string, boardIDs []string, limit int, cursor, typeFilter, search string) ([]*domain.Notification, string, error)
	MarkBoardRead(ctx context.Context, userID, boardID string) error
	MarkAllBoardsRead(ctx context.Context, userID string, boardIDs []string) error
	GetBoardIDByEventID(ctx context.Context, eventID string) (string, error)
	GetUnreadCountBySeq(ctx context.Context, userID string, boardIDs []string) (int, error)
}

// UnreadCounter — счётчик непрочитанных уведомлений в Redis (O(1) вместо SQL COUNT).
// UnreadCounter — Redis lazy cache для unread count.
// НЕ источник истины. Вычисляется из SQL seq diff при cache miss.
type UnreadCounter interface {
	Get(ctx context.Context, userID string) (int, error)    // -1 при cache miss
	Set(ctx context.Context, userID string, count int) error // кэшировать вычисленное значение
	Invalidate(ctx context.Context, userID string) error     // удалить кэш (при mark read)
}

// EventPublisher — публикация событий для WebSocket доставки.
type EventPublisher interface {
	PublishNotificationCreated(ctx context.Context, n *domain.Notification) error
	PublishNotificationsBatch(ctx context.Context, notifications []*domain.Notification) error
	PublishBoardEventNotification(ctx context.Context, event *domain.BoardEvent) error
}

// SettingsEventPublisher — публикация событий при изменении настроек.
type SettingsEventPublisher interface {
	PublishSettingsUpdated(ctx context.Context, userID string, enabled, realtimeEnabled bool) error
}
