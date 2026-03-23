package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAutomationRule_Valid(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "board-123", "Move done cards",
		TriggerCardMovedToColumn,
		map[string]string{"column_id": "col-123"},
		ActionAddLabel,
		map[string]string{"label_id": "label-456"},
		"user-123",
	)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.NotEmpty(t, rule.ID)
	assert.Equal(t, "board-123", rule.BoardID)
	assert.Equal(t, "Move done cards", rule.Name)
	assert.True(t, rule.Enabled)
	assert.Equal(t, TriggerCardMovedToColumn, rule.TriggerType)
	assert.Equal(t, "col-123", rule.TriggerConfig["column_id"])
	assert.Equal(t, ActionAddLabel, rule.ActionType)
	assert.Equal(t, "label-456", rule.ActionConfig["label_id"])
	assert.Equal(t, "user-123", rule.CreatedBy)
	assert.False(t, rule.CreatedAt.IsZero())
	assert.False(t, rule.UpdatedAt.IsZero())
}

func TestNewAutomationRule_WithID(t *testing.T) {
	rule, err := NewAutomationRule(
		"custom-id", "board-123", "Rule",
		TriggerCardCreated, nil,
		ActionSetPriority, nil,
		"user-123",
	)

	assert.NoError(t, err)
	assert.Equal(t, "custom-id", rule.ID)
	assert.NotNil(t, rule.TriggerConfig)
	assert.NotNil(t, rule.ActionConfig)
}

func TestNewAutomationRule_EmptyName(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "board-123", "",
		TriggerCardCreated, nil,
		ActionMoveCard, nil,
		"user-123",
	)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyRuleName, err)
	assert.Nil(t, rule)
}

func TestNewAutomationRule_InvalidTrigger(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "board-123", "Rule",
		TriggerType("invalid_trigger"), nil,
		ActionMoveCard, nil,
		"user-123",
	)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTriggerType, err)
	assert.Nil(t, rule)
}

func TestNewAutomationRule_InvalidAction(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "board-123", "Rule",
		TriggerCardCreated, nil,
		ActionType("invalid_action"), nil,
		"user-123",
	)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidActionType, err)
	assert.Nil(t, rule)
}

func TestNewAutomationRule_EmptyBoardID(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "", "Rule",
		TriggerCardCreated, nil,
		ActionMoveCard, nil,
		"user-123",
	)

	assert.Error(t, err)
	assert.Equal(t, ErrBoardNotFound, err)
	assert.Nil(t, rule)
}

func TestNewAutomationRule_EmptyCreatedBy(t *testing.T) {
	rule, err := NewAutomationRule(
		"", "board-123", "Rule",
		TriggerCardCreated, nil,
		ActionMoveCard, nil,
		"",
	)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyOwnerID, err)
	assert.Nil(t, rule)
}

func TestAutomationRule_Update(t *testing.T) {
	rule, _ := NewAutomationRule(
		"", "board-123", "Original",
		TriggerCardCreated, map[string]string{"key": "val"},
		ActionMoveCard, map[string]string{"target": "col-1"},
		"user-123",
	)

	err := rule.Update("Updated", false, map[string]string{"key": "new"}, map[string]string{"target": "col-2"})

	assert.NoError(t, err)
	assert.Equal(t, "Updated", rule.Name)
	assert.False(t, rule.Enabled)
	assert.Equal(t, "new", rule.TriggerConfig["key"])
	assert.Equal(t, "col-2", rule.ActionConfig["target"])
}

func TestAutomationRule_Update_EmptyName(t *testing.T) {
	rule, _ := NewAutomationRule(
		"", "board-123", "Original",
		TriggerCardCreated, nil,
		ActionMoveCard, nil,
		"user-123",
	)

	err := rule.Update("", true, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, ErrEmptyRuleName, err)
	assert.Equal(t, "Original", rule.Name)
}

func TestAutomationRule_Update_Enable(t *testing.T) {
	rule, _ := NewAutomationRule(
		"", "board-123", "Rule",
		TriggerCardCreated, nil,
		ActionMoveCard, nil,
		"user-123",
	)

	// Disable
	err := rule.Update("Rule", false, nil, nil)
	assert.NoError(t, err)
	assert.False(t, rule.Enabled)

	// Re-enable
	err = rule.Update("Rule", true, nil, nil)
	assert.NoError(t, err)
	assert.True(t, rule.Enabled)
}

func TestAutomationRule_Update_NilConfigs_KeepsExisting(t *testing.T) {
	rule, _ := NewAutomationRule(
		"", "board-123", "Rule",
		TriggerCardCreated, map[string]string{"key": "val"},
		ActionMoveCard, map[string]string{"target": "col-1"},
		"user-123",
	)

	err := rule.Update("Rule", true, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "val", rule.TriggerConfig["key"])
	assert.Equal(t, "col-1", rule.ActionConfig["target"])
}

func TestTriggerType_IsValid(t *testing.T) {
	validTypes := []TriggerType{
		TriggerCardMovedToColumn,
		TriggerCardCreated,
		TriggerDueDatePassed,
		TriggerLabelAdded,
		TriggerChecklistCompleted,
	}

	for _, tt := range validTypes {
		assert.True(t, tt.IsValid(), "expected %s to be valid", tt)
	}

	invalidTypes := []TriggerType{
		"",
		"invalid",
		"CARD_MOVED_TO_COLUMN",
	}

	for _, tt := range invalidTypes {
		assert.False(t, tt.IsValid(), "expected %s to be invalid", tt)
	}
}

func TestActionType_IsValid(t *testing.T) {
	validTypes := []ActionType{
		ActionMoveCard,
		ActionAssignMember,
		ActionAddLabel,
		ActionSetPriority,
	}

	for _, at := range validTypes {
		assert.True(t, at.IsValid(), "expected %s to be valid", at)
	}

	invalidTypes := []ActionType{
		"",
		"invalid",
		"MOVE_CARD",
	}

	for _, at := range invalidTypes {
		assert.False(t, at.IsValid(), "expected %s to be invalid", at)
	}
}
