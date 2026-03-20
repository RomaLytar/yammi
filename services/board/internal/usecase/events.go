package usecase

import (
	"time"

	"github.com/google/uuid"
)

// События для Board Service
type BoardCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	OwnerID      string    `json:"owner_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

type BoardUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

type BoardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
}

type ColumnAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	Title        string    `json:"title"`
	Position     int       `json:"position"`
}

type ColumnUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	Title        string    `json:"title"`
}

type ColumnDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
}

type ColumnsReordered struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	Columns      []string  `json:"columns"` // ordered column IDs
}

type CardCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	CardID       string    `json:"card_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Position     string    `json:"position"`
}

type CardUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	CardID       string    `json:"card_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	AssigneeID   *string   `json:"assignee_id"`
}

type CardMoved struct {
	EventID        string    `json:"event_id"`
	EventVersion   int       `json:"event_version"`
	OccurredAt     time.Time `json:"occurred_at"`
	BoardID        string    `json:"board_id"`
	CardID         string    `json:"card_id"`
	FromColumnID   string    `json:"from_column_id"`
	ToColumnID     string    `json:"to_column_id"`
	NewPosition    string    `json:"new_position"`
}

type CardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
}

type MemberAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
}

type MemberRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
}

func generateEventID() string {
	return uuid.New().String()
}
