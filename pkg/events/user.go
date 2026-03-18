package events

import "time"

const (
	SubjectUserCreated = "user.created"
	SubjectUserDeleted = "user.deleted"
	StreamUsers        = "USERS"
)

type UserCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
}

type UserDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	UserID       string    `json:"user_id"`
}
