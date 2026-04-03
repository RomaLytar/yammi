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

type ChecklistCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ChecklistID  string    `json:"checklist_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
}

type ChecklistUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ChecklistID  string    `json:"checklist_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Title        string    `json:"title"`
}

type ChecklistDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ChecklistID  string    `json:"checklist_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type ChecklistItemToggled struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ItemID       string    `json:"item_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	IsChecked    bool      `json:"is_checked"`
}

type CustomFieldCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	FieldID      string    `json:"field_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
	FieldType    string    `json:"field_type"`
}

type CustomFieldUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	FieldID      string    `json:"field_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
}

type CustomFieldDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	FieldID      string    `json:"field_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type CustomFieldValueSet struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	FieldID      string    `json:"field_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type AutomationRuleCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	RuleID       string    `json:"rule_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
	TriggerType  string    `json:"trigger_type"`
	ActionType   string    `json:"action_type"`
}

type AutomationRuleUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	RuleID       string    `json:"rule_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Name         string    `json:"name"`
	Enabled      bool      `json:"enabled"`
}

type AutomationRuleDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	RuleID       string    `json:"rule_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

type AutomationExecuted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	ExecutionID  string    `json:"execution_id"`
	RuleID       string    `json:"rule_id"`
	BoardID      string    `json:"board_id"`
	CardID       string    `json:"card_id"`
	Status       string    `json:"status"`
}

type ReleaseCreatedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ReleaseID    string    `json:"release_id"`
	Name         string    `json:"name"`
	ActorID      string    `json:"actor_id"`
}

type ReleaseUpdatedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ReleaseID    string    `json:"release_id"`
	Name         string    `json:"name"`
	ActorID      string    `json:"actor_id"`
}

type ReleaseStartedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ReleaseID    string    `json:"release_id"`
	Name         string    `json:"name"`
	ActorID      string    `json:"actor_id"`
}

type ReleaseCompletedEvent struct {
	EventID             string    `json:"event_id"`
	EventVersion        int       `json:"event_version"`
	OccurredAt          time.Time `json:"occurred_at"`
	BoardID             string    `json:"board_id"`
	ReleaseID           string    `json:"release_id"`
	Name                string    `json:"name"`
	ActorID             string    `json:"actor_id"`
	CardsMovedToBacklog int       `json:"cards_moved_to_backlog"`
}

type ReleaseDeletedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	ReleaseID    string    `json:"release_id"`
	ActorID      string    `json:"actor_id"`
}

type CardReleaseAssignedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	CardID       string    `json:"card_id"`
	ReleaseID    string    `json:"release_id"`
	ActorID      string    `json:"actor_id"`
}

type CardReleaseRemovedEvent struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	BoardID      string    `json:"board_id"`
	CardID       string    `json:"card_id"`
	ReleaseID    string    `json:"release_id"`
	ActorID      string    `json:"actor_id"`
}

func generateEventID() string {
	return uuid.New().String()
}
