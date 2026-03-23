package domain

import (
	"time"

	"github.com/google/uuid"
)

// TriggerType определяет тип триггера правила автоматизации
type TriggerType string

const (
	TriggerCardMovedToColumn  TriggerType = "card_moved_to_column"
	TriggerCardCreated        TriggerType = "card_created"
	TriggerDueDatePassed      TriggerType = "due_date_passed"
	TriggerLabelAdded         TriggerType = "label_added"
	TriggerChecklistCompleted TriggerType = "checklist_completed"
)

// ActionType определяет тип действия правила автоматизации
type ActionType string

const (
	ActionMoveCard     ActionType = "move_card"
	ActionAssignMember ActionType = "assign_member"
	ActionAddLabel     ActionType = "add_label"
	ActionSetPriority  ActionType = "set_priority"
)

// AutomationRule — правило автоматизации доски (trigger -> action)
type AutomationRule struct {
	ID            string
	BoardID       string
	Name          string
	Enabled       bool
	TriggerType   TriggerType
	TriggerConfig map[string]string // e.g. {"column_id": "uuid"}
	ActionType    ActionType
	ActionConfig  map[string]string // e.g. {"target_column_id": "uuid"}
	CreatedBy     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AutomationExecution — запись о выполнении правила автоматизации
type AutomationExecution struct {
	ID             string
	RuleID         string
	BoardID        string
	CardID         string
	TriggerEventID string
	Status         string // "success", "failed", "skipped"
	ErrorMessage   string
	ExecutedAt     time.Time
}

// IsValid проверяет валидность типа триггера
func (tt TriggerType) IsValid() bool {
	switch tt {
	case TriggerCardMovedToColumn, TriggerCardCreated, TriggerDueDatePassed,
		TriggerLabelAdded, TriggerChecklistCompleted:
		return true
	}
	return false
}

// IsValid проверяет валидность типа действия
func (at ActionType) IsValid() bool {
	switch at {
	case ActionMoveCard, ActionAssignMember, ActionAddLabel, ActionSetPriority:
		return true
	}
	return false
}

// NewAutomationRule создает новое правило автоматизации с валидацией
func NewAutomationRule(id, boardID, name string, triggerType TriggerType, triggerConfig map[string]string, actionType ActionType, actionConfig map[string]string, createdBy string) (*AutomationRule, error) {
	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if name == "" {
		return nil, ErrEmptyRuleName
	}

	if !triggerType.IsValid() {
		return nil, ErrInvalidTriggerType
	}

	if !actionType.IsValid() {
		return nil, ErrInvalidActionType
	}

	if createdBy == "" {
		return nil, ErrEmptyOwnerID
	}

	if id == "" {
		id = uuid.NewString()
	}

	if triggerConfig == nil {
		triggerConfig = make(map[string]string)
	}

	if actionConfig == nil {
		actionConfig = make(map[string]string)
	}

	now := time.Now()
	return &AutomationRule{
		ID:            id,
		BoardID:       boardID,
		Name:          name,
		Enabled:       true,
		TriggerType:   triggerType,
		TriggerConfig: triggerConfig,
		ActionType:    actionType,
		ActionConfig:  actionConfig,
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Update обновляет правило автоматизации
func (r *AutomationRule) Update(name string, enabled bool, triggerConfig map[string]string, actionConfig map[string]string) error {
	if name == "" {
		return ErrEmptyRuleName
	}

	r.Name = name
	r.Enabled = enabled

	if triggerConfig != nil {
		r.TriggerConfig = triggerConfig
	}

	if actionConfig != nil {
		r.ActionConfig = actionConfig
	}

	r.UpdatedAt = time.Now()
	return nil
}
