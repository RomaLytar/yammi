package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAutomationRuleRepository - мок для AutomationRuleRepository
type MockAutomationRuleRepository struct {
	mock.Mock
}

func (m *MockAutomationRuleRepository) Create(ctx context.Context, rule *domain.AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAutomationRuleRepository) GetByID(ctx context.Context, ruleID, boardID string) (*domain.AutomationRule, error) {
	args := m.Called(ctx, ruleID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AutomationRule), args.Error(1)
}

func (m *MockAutomationRuleRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.AutomationRule, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AutomationRule), args.Error(1)
}

func (m *MockAutomationRuleRepository) ListEnabledByBoardAndTrigger(ctx context.Context, boardID string, triggerType domain.TriggerType) ([]*domain.AutomationRule, error) {
	args := m.Called(ctx, boardID, triggerType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AutomationRule), args.Error(1)
}

func (m *MockAutomationRuleRepository) Update(ctx context.Context, rule *domain.AutomationRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockAutomationRuleRepository) Delete(ctx context.Context, ruleID, boardID string) error {
	args := m.Called(ctx, ruleID, boardID)
	return args.Error(0)
}

func (m *MockAutomationRuleRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	args := m.Called(ctx, boardID)
	return args.Int(0), args.Error(1)
}

func (m *MockAutomationRuleRepository) CreateExecution(ctx context.Context, exec *domain.AutomationExecution) error {
	args := m.Called(ctx, exec)
	return args.Error(0)
}

func (m *MockAutomationRuleRepository) ListExecutionsByRuleID(ctx context.Context, ruleID, boardID string, limit int) ([]*domain.AutomationExecution, error) {
	args := m.Called(ctx, ruleID, boardID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AutomationExecution), args.Error(1)
}

func TestCreateRule_Success(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("CountByBoardID", mock.Anything, "board-123").Return(5, nil)
	ruleRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AutomationRule")).Return(nil)
	publisher.On("PublishAutomationRuleCreated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-123", "Auto move",
		domain.TriggerCardMovedToColumn, map[string]string{"column_id": "col-1"},
		domain.ActionAddLabel, map[string]string{"label_id": "label-1"})

	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "Auto move", rule.Name)
	assert.Equal(t, domain.TriggerCardMovedToColumn, rule.TriggerType)
	assert.Equal(t, domain.ActionAddLabel, rule.ActionType)
	assert.True(t, rule.Enabled)
	assert.NotEmpty(t, rule.ID)

	ruleRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestCreateRule_NonMember(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-999", "Rule",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
}

func TestCreateRule_NotOwner(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-456", "Rule",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
}

func TestCreateRule_EmptyName(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("CountByBoardID", mock.Anything, "board-123").Return(5, nil)

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-123", "",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyRuleName, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
}

func TestCreateRule_InvalidTrigger(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("CountByBoardID", mock.Anything, "board-123").Return(5, nil)

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-123", "Rule",
		domain.TriggerType("invalid"), nil, domain.ActionMoveCard, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidTriggerType, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
}

func TestCreateRule_MaxReached(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("CountByBoardID", mock.Anything, "board-123").Return(25, nil)

	uc := NewCreateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "board-123", "user-123", "Rule",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxRulesReached, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
	ruleRepo.AssertExpectations(t)
}

func TestUpdateRule_Owner(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	existingRule := &domain.AutomationRule{
		ID:            "rule-123",
		BoardID:       "board-123",
		Name:          "Original",
		Enabled:       true,
		TriggerType:   domain.TriggerCardCreated,
		TriggerConfig: map[string]string{},
		ActionType:    domain.ActionMoveCard,
		ActionConfig:  map[string]string{"target_column_id": "col-1"},
		CreatedBy:     "user-123",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("GetByID", mock.Anything, "rule-123", "board-123").Return(existingRule, nil)
	ruleRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.AutomationRule")).Return(nil)
	publisher.On("PublishAutomationRuleUpdated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewUpdateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "rule-123", "board-123", "user-123",
		"Updated", false, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, rule)
	assert.Equal(t, "Updated", rule.Name)
	assert.False(t, rule.Enabled)

	ruleRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestUpdateRule_NotOwner(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewUpdateAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	rule, err := uc.Execute(context.Background(), "rule-123", "board-123", "user-456",
		"Updated", true, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)
	assert.Nil(t, rule)

	memberRepo.AssertExpectations(t)
}

func TestDeleteRule_Owner(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	ruleRepo.On("Delete", mock.Anything, "rule-123", "board-123").Return(nil)
	publisher.On("PublishAutomationRuleDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewDeleteAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "rule-123", "board-123", "user-123")

	assert.NoError(t, err)

	ruleRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestDeleteRule_Member_Denied(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewDeleteAutomationRuleUseCase(ruleRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "rule-123", "board-123", "user-456")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)

	memberRepo.AssertExpectations(t)
}

func TestListRules_Member(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)
	ruleRepo.On("ListByBoardID", mock.Anything, "board-123").
		Return([]*domain.AutomationRule{
			{ID: "rule-1", BoardID: "board-123", Name: "Rule 1"},
			{ID: "rule-2", BoardID: "board-123", Name: "Rule 2"},
		}, nil)

	uc := NewListAutomationRulesUseCase(ruleRepo, memberRepo)
	rules, err := uc.Execute(context.Background(), "board-123", "user-456")

	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "Rule 1", rules[0].Name)

	ruleRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetHistory_Success(t *testing.T) {
	ruleRepo := new(MockAutomationRuleRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	ruleRepo.On("ListExecutionsByRuleID", mock.Anything, "rule-123", "board-123", 50).
		Return([]*domain.AutomationExecution{
			{ID: "exec-1", RuleID: "rule-123", BoardID: "board-123", Status: "success"},
			{ID: "exec-2", RuleID: "rule-123", BoardID: "board-123", Status: "failed", ErrorMessage: "card not found"},
		}, nil)

	uc := NewGetAutomationHistoryUseCase(ruleRepo, memberRepo)
	execs, err := uc.Execute(context.Background(), "rule-123", "board-123", "user-123", 50)

	assert.NoError(t, err)
	assert.Len(t, execs, 2)
	assert.Equal(t, "success", execs[0].Status)
	assert.Equal(t, "failed", execs[1].Status)

	ruleRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}
