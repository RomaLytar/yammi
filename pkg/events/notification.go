package events

import "time"

const (
	SubjectNotificationCreated        = "notification.created"
	SubjectNotificationSettingsUpdated = "notification.settings.updated"
	StreamNotifications                = "NOTIFICATIONS"
)

// NotificationCreated — уведомление создано и готово к доставке через WebSocket.
type NotificationCreated struct {
	EventID      string            `json:"event_id"`
	EventVersion int               `json:"event_version"`
	OccurredAt   time.Time         `json:"occurred_at"`
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Type         string            `json:"type"`
	Title        string            `json:"title"`
	Message      string            `json:"message"`
	Metadata     map[string]string `json:"metadata"`
}

// NotificationSettingsUpdated — настройки уведомлений изменены (для инвалидации кеша).
type NotificationSettingsUpdated struct {
	EventID         string    `json:"event_id"`
	EventVersion    int       `json:"event_version"`
	OccurredAt      time.Time `json:"occurred_at"`
	UserID          string    `json:"user_id"`
	Enabled         bool      `json:"enabled"`
	RealtimeEnabled bool      `json:"realtime_enabled"`
}
