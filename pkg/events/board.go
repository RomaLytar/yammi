package events

import "time"

const (
	SubjectBoardCreated     = "board.created"
	SubjectBoardUpdated     = "board.updated"
	SubjectBoardDeleted     = "board.deleted"
	SubjectColumnCreated    = "column.created"
	SubjectColumnUpdated    = "column.updated"
	SubjectColumnDeleted    = "column.deleted"
	SubjectColumnsReordered = "columns.reordered"
	SubjectCardCreated      = "card.created"
	SubjectCardUpdated      = "card.updated"
	SubjectCardMoved        = "card.moved"
	SubjectCardDeleted      = "card.deleted"
	SubjectCardAssigned     = "card.assigned"
	SubjectCardUnassigned   = "card.unassigned"
	SubjectMemberAdded          = "member.added"
	SubjectMemberRemoved        = "member.removed"
	SubjectAttachmentUploaded   = "attachment.uploaded"
	SubjectAttachmentDeleted    = "attachment.deleted"
	SubjectLabelCreated         = "label.created"
	SubjectLabelUpdated         = "label.updated"
	SubjectLabelDeleted         = "label.deleted"
	SubjectCardLabelAdded       = "card.label.added"
	SubjectCardLabelRemoved     = "card.label.removed"
	SubjectCardLinked           = "card.linked"
	SubjectCardUnlinked         = "card.unlinked"
	StreamBoards                = "BOARDS"
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
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
}

// BoardDeleted событие удаления доски
type BoardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

// ColumnCreated событие создания колонки
type ColumnCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
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
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
}

// ColumnDeleted событие удаления колонки
type ColumnDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

// ColumnsReordered событие переупорядочивания колонок
type ColumnsReordered struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Columns      []string  `json:"columns"`
}

// CardCreated событие создания карточки
type CardCreated struct {
	EventID      string     `json:"event_id"`
	EventVersion int        `json:"event_version"`
	OccurredAt   time.Time  `json:"occurred_at"`
	CardID       string     `json:"card_id"`
	ColumnID     string     `json:"column_id"`
	BoardID      string     `json:"board_id"`
	ActorID      string     `json:"actor_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Position     string     `json:"position"`
	AssigneeID   *string    `json:"assignee_id,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	Priority     string     `json:"priority"`
	TaskType     string     `json:"task_type"`
}

// CardUpdated событие обновления карточки
type CardUpdated struct {
	EventID      string     `json:"event_id"`
	EventVersion int        `json:"event_version"`
	OccurredAt   time.Time  `json:"occurred_at"`
	CardID       string     `json:"card_id"`
	ColumnID     string     `json:"column_id"`
	BoardID      string     `json:"board_id"`
	ActorID      string     `json:"actor_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	AssigneeID   *string    `json:"assignee_id,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	Priority     string     `json:"priority"`
	TaskType     string     `json:"task_type"`
}

// CardMoved событие перемещения карточки
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

// CardDeleted событие удаления карточки
type CardDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	ColumnID     string    `json:"column_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

// CardAssigned событие назначения карточки
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

// CardUnassigned событие снятия назначения с карточки
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

// MemberAdded событие добавления участника
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

// MemberRemoved событие удаления участника
type MemberRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	UserID       string    `json:"user_id"`
	ActorID      string    `json:"actor_id"`
	BoardTitle   string    `json:"board_title"`
}

// AttachmentUploaded событие загрузки вложения
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

// AttachmentDeleted событие удаления вложения
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

// LabelCreated событие создания метки
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

// LabelUpdated событие обновления метки
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

// LabelDeleted событие удаления метки
type LabelDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LabelID      string    `json:"label_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

// CardLabelAdded событие назначения метки на карточку
type CardLabelAdded struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	LabelID      string    `json:"label_id"`
	ActorID      string    `json:"actor_id"`
}

// CardLabelRemoved событие снятия метки с карточки
type CardLabelRemoved struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	LabelID      string    `json:"label_id"`
	ActorID      string    `json:"actor_id"`
}

// CardLinked событие создания связи между карточками
type CardLinked struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LinkID       string    `json:"link_id"`
	ParentID     string    `json:"parent_id"`
	ChildID      string    `json:"child_id"`
	BoardID      string    `json:"board_id"`
	LinkType     string    `json:"link_type"`
	ActorID      string    `json:"actor_id"`
}

// CardUnlinked событие удаления связи между карточками
type CardUnlinked struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	LinkID       string    `json:"link_id"`
	ParentID     string    `json:"parent_id"`
	ChildID      string    `json:"child_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}
