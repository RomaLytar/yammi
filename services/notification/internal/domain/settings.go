package domain

import "time"

type NotificationSettings struct {
	UserID          string
	Enabled         bool
	RealtimeEnabled bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DefaultSettings возвращает настройки по умолчанию для пользователя.
func DefaultSettings(userID string) *NotificationSettings {
	now := time.Now()
	return &NotificationSettings{
		UserID:          userID,
		Enabled:         true,
		RealtimeEnabled: true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
