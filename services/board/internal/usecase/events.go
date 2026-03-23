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
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

type BoardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type ColumnAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
	Position     int       `json:"position"`
}

type ColumnUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
}

type ColumnDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	ActorID      string    `json:"actor_id"`
}

type ColumnsReordered struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Columns      []string  `json:"columns"` // ordered column IDs
}

type CardCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Position     string    `json:"position"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
}

type CardUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
}

type CardMoved struct {
	EventID        string    `json:"event_id"`
	EventVersion   int       `json:"event_version"`
	OccurredAt     time.Time `json:"occurred_at"`
	CardID         string    `json:"card_id"`
	BoardID        string    `json:"board_id"`
	ActorID        string    `json:"actor_id"`
	FromColumnID   string    `json:"from_column_id"`
	ToColumnID     string    `json:"to_column_id"`
	NewPosition    string    `json:"new_position"`
}

type CardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type MemberAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	ActorID      string    `json:"actor_id"`
	Role         string    `json:"role"`
	BoardTitle   string    `json:"board_title"`
}

type MemberRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	ActorID      string    `json:"actor_id"`
	BoardTitle   string    `json:"board_title"`
}

type CardAssigned struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	ActorID      string    `json:"actor_id"`
	AssigneeID   string    `json:"assignee_id"`
	PrevAssignee *string   `json:"prev_assignee,omitempty"`
	CardTitle    string    `json:"card_title"`
}

type CardUnassigned struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ColumnID     string    `json:"column_id"`
	ActorID      string    `json:"actor_id"`
	PrevAssignee string    `json:"prev_assignee"`
	CardTitle    string    `json:"card_title"`
}

type AttachmentUploaded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	AttachmentID string    `json:"attachment_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
}

type AttachmentDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	AttachmentID string    `json:"attachment_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	FileName     string    `json:"file_name"`
}

type LabelCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LabelID      string    `json:"label_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
}

type LabelUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LabelID      string    `json:"label_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
}

type LabelDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LabelID      string    `json:"label_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type CardLabelAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	LabelID      string    `json:"label_id"`
	ActorID      string    `json:"actor_id"`
}

type CardLabelRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	LabelID      string    `json:"label_id"`
	ActorID      string    `json:"actor_id"`
}

func generateEventID() string {
	return uuid.New().String()
}
