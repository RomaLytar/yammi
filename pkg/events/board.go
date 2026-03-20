package events

import "time"

const (
	SubjectBoardCreated  = "board.created"
	SubjectBoardUpdated  = "board.updated"
	SubjectBoardDeleted  = "board.deleted"
	SubjectColumnCreated = "column.created"
	SubjectColumnUpdated = "column.updated"
	SubjectColumnDeleted = "column.deleted"
	SubjectCardCreated   = "card.created"
	SubjectCardUpdated   = "card.updated"
	SubjectCardMoved     = "card.moved"
	SubjectCardDeleted   = "card.deleted"
	SubjectMemberAdded   = "member.added"
	SubjectMemberRemoved = "member.removed"
	StreamBoards         = "BOARDS"
)

// BoardCreated событие создания доски
type BoardCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	OwnerID      string    `json:"owner_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

// BoardUpdated событие обновления доски
type BoardUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

// BoardDeleted событие удаления доски
type BoardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
}

// ColumnCreated событие создания колонки
type ColumnCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	Title        string    `json:"title"`
	Position     int       `json:"position"`
}

// ColumnUpdated событие обновления колонки
type ColumnUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	Title        string    `json:"title"`
}

// ColumnDeleted событие удаления колонки
type ColumnDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
}

// CardCreated событие создания карточки
type CardCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Position     string    `json:"position"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
}

// CardUpdated событие обновления карточки
type CardUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	AssigneeID   *string   `json:"assignee_id,omitempty"`
}

// CardMoved событие перемещения карточки
type CardMoved struct {
	EventID         string    `json:"event_id"`
	EventVersion    int       `json:"event_version"`
	OccurredAt      time.Time `json:"occurred_at"`
	CardID          string    `json:"card_id"`
	BoardID         string    `json:"board_id"`
	SourceColumnID  string    `json:"source_column_id"`
	TargetColumnID  string    `json:"target_column_id"`
	NewPosition     string    `json:"new_position"`
}

// CardDeleted событие удаления карточки
type CardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
}

// MemberAdded событие добавления участника
type MemberAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
}

// MemberRemoved событие удаления участника
type MemberRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
}
