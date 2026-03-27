package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecuteAutomations_MatchingRule_AssignMember(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	userID := "user-456"
	card := &domain.Card{
		ID:         "card-123",
		ColumnID:   "col-done",
		Title:      "Test Card",
		Position:   "n",
		AssigneeID: nil,
		Priority:   domain.PriorityMedium,
		TaskType:   domain.TaskTypeTask,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	rule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Auto assign on done",
		Enabled:       true,
		TriggerType:   domain.TriggerCardMovedToColumn,
		TriggerConfig: map[string]string{"column_id": "col-done"},
		ActionType:    domain.ActionAssignMember,
		ActionConfig:  map[string]string{"user_id": userID},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{rule}, nil)
	cardRepo.On("GetByID", mock.Anything, "card-123", "board-123").Return(card, nil)
	cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
	automationRepo.On("CreateExecution", mock.Anything, mock.AnythingOfType("*domain.AutomationExecution")).Return(nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	assert.NoError(t, err)

	// Проверяем, что карточка обновлена с новым assignee
	cardRepo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(c *domain.Card) bool {
		return c.AssigneeID != nil && *c.AssigneeID == userID
	}))

	// Проверяем, что execution записан со статусом "success"
	automationRepo.AssertCalled(t, "CreateExecution", mock.Anything, mock.MatchedBy(func(exec *domain.AutomationExecution) bool {
		return exec.RuleID == "rule-123" && exec.Status == "success" && exec.CardID == "card-123" && exec.BoardID == "board-123"
	}))

	automationRepo.AssertExpectations(t)
	cardRepo.AssertExpectations(t)
}

func TestExecuteAutomations_NonMatchingColumnID(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	rule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Auto assign on done",
		Enabled:       true,
		TriggerType:   domain.TriggerCardMovedToColumn,
		TriggerConfig: map[string]string{"column_id": "col-done"},
		ActionType:    domain.ActionAssignMember,
		ActionConfig:  map[string]string{"user_id": "user-456"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{rule}, nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-other"})

	assert.NoError(t, err)

	// Не должно быть вызовов к cardRepo или CreateExecution
	cardRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
	cardRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	automationRepo.AssertNotCalled(t, "CreateExecution", mock.Anything, mock.Anything)

	automationRepo.AssertExpectations(t)
}

func TestExecuteAutomations_DisabledRule(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	// ListEnabledByBoardAndTrigger фильтрует disabled правила на уровне БД,
	// поэтому disabled правило не вернётся вообще
	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{}, nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	assert.NoError(t, err)

	cardRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
	cardRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	automationRepo.AssertNotCalled(t, "CreateExecution", mock.Anything, mock.Anything)

	automationRepo.AssertExpectations(t)
}

func TestExecuteAutomations_NoRules(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{}, nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	assert.NoError(t, err)

	cardRepo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
	automationRepo.AssertNotCalled(t, "CreateExecution", mock.Anything, mock.Anything)

	automationRepo.AssertExpectations(t)
}

func TestExecuteAutomations_ActionAddLabel(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	rule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Auto label on move",
		Enabled:       true,
		TriggerType:   domain.TriggerCardMovedToColumn,
		TriggerConfig: map[string]string{"column_id": "col-done"},
		ActionType:    domain.ActionAddLabel,
		ActionConfig:  map[string]string{"label_id": "label-789"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{rule}, nil)
	labelRepo.On("AddToCard", mock.Anything, "card-123", "board-123", "label-789").Return(nil)
	automationRepo.On("CreateExecution", mock.Anything, mock.AnythingOfType("*domain.AutomationExecution")).Return(nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	assert.NoError(t, err)

	labelRepo.AssertCalled(t, "AddToCard", mock.Anything, "card-123", "board-123", "label-789")

	// Проверяем execution со статусом "success"
	automationRepo.AssertCalled(t, "CreateExecution", mock.Anything, mock.MatchedBy(func(exec *domain.AutomationExecution) bool {
		return exec.RuleID == "rule-123" && exec.Status == "success"
	}))

	automationRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
}

func TestExecuteAutomations_ActionSetPriority(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	card := &domain.Card{
		ID:         "card-123",
		ColumnID:   "col-done",
		Title:      "Test Card",
		Position:   "n",
		AssigneeID: nil,
		Priority:   domain.PriorityMedium,
		TaskType:   domain.TaskTypeTask,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	rule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Set priority on move",
		Enabled:       true,
		TriggerType:   domain.TriggerCardMovedToColumn,
		TriggerConfig: map[string]string{"column_id": "col-done"},
		ActionType:    domain.ActionSetPriority,
		ActionConfig:  map[string]string{"priority": "critical"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{rule}, nil)
	cardRepo.On("GetByID", mock.Anything, "card-123", "board-123").Return(card, nil)
	cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
	automationRepo.On("CreateExecution", mock.Anything, mock.AnythingOfType("*domain.AutomationExecution")).Return(nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	assert.NoError(t, err)

	// Проверяем, что приоритет обновлён
	cardRepo.AssertCalled(t, "Update", mock.Anything, mock.MatchedBy(func(c *domain.Card) bool {
		return c.Priority == domain.PriorityCritical
	}))

	automationRepo.AssertCalled(t, "CreateExecution", mock.Anything, mock.MatchedBy(func(exec *domain.AutomationExecution) bool {
		return exec.RuleID == "rule-123" && exec.Status == "success"
	}))

	automationRepo.AssertExpectations(t)
	cardRepo.AssertExpectations(t)
}

func TestExecuteAutomations_ExecutionFailure(t *testing.T) {
	automationRepo := new(MockAutomationRuleRepository)
	cardRepo := new(MockCardRepository)
	labelRepo := new(MockLabelRepository)

	rule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Auto label on move",
		Enabled:       true,
		TriggerType:   domain.TriggerCardMovedToColumn,
		TriggerConfig: map[string]string{"column_id": "col-done"},
		ActionType:    domain.ActionAddLabel,
		ActionConfig:  map[string]string{"label_id": "label-789"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	automationRepo.On("ListEnabledByBoardAndTrigger", mock.Anything, "board-123", domain.TriggerCardMovedToColumn).
		Return([]*domain.AutomationRule{rule}, nil)
	labelRepo.On("AddToCard", mock.Anything, "card-123", "board-123", "label-789").
		Return(errors.New("label not found"))
	automationRepo.On("CreateExecution", mock.Anything, mock.AnythingOfType("*domain.AutomationExecution")).Return(nil)

	uc := NewExecuteAutomationsUseCase(automationRepo, cardRepo, labelRepo)
	err := uc.Execute(context.Background(), "board-123", "card-123",
		domain.TriggerCardMovedToColumn, map[string]string{"to_column_id": "col-done"})

	// Execute не возвращает ошибку — ошибки логируются, не пробрасываются
	assert.NoError(t, err)

	// Проверяем, что execution записан со статусом "failed" и сообщением об ошибке
	automationRepo.AssertCalled(t, "CreateExecution", mock.Anything, mock.MatchedBy(func(exec *domain.AutomationExecution) bool {
		return exec.RuleID == "rule-123" && exec.Status == "failed" && exec.ErrorMessage == "label not found"
	}))

	automationRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
}
