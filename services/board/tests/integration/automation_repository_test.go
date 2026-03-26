package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestAutomationRuleRepository_Create(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create automation rule
	rule, err := domain.NewAutomationRule("", board.ID, "Auto label on move",
		domain.TriggerCardMovedToColumn, map[string]string{"column_id": uuid.NewString()},
		domain.ActionAddLabel, map[string]string{"label_id": uuid.NewString()},
		ownerID,
	)
	if err != nil {
		t.Fatalf("Failed to create domain rule: %v", err)
	}

	err = ruleRepo.Create(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to save rule: %v", err)
	}

	// Verify rule exists
	loaded, err := ruleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to load rule: %v", err)
	}

	if loaded.Name != "Auto label on move" {
		t.Errorf("Expected name 'Auto label on move', got %s", loaded.Name)
	}
	if loaded.TriggerType != domain.TriggerCardMovedToColumn {
		t.Errorf("Expected trigger type card_moved_to_column, got %s", loaded.TriggerType)
	}
	if loaded.ActionType != domain.ActionAddLabel {
		t.Errorf("Expected action type add_label, got %s", loaded.ActionType)
	}
	if !loaded.Enabled {
		t.Error("Expected rule to be enabled")
	}
	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestAutomationRuleRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	_, err := ruleRepo.GetByID(ctx, uuid.NewString())
	if err != domain.ErrAutomationRuleNotFound {
		t.Errorf("Expected ErrAutomationRuleNotFound, got %v", err)
	}
}

func TestAutomationRuleRepository_ListByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create 3 rules
	for i := 0; i < 3; i++ {
		rule, _ := domain.NewAutomationRule("", board.ID, "Rule "+uuid.NewString()[:4],
			domain.TriggerCardCreated, nil,
			domain.ActionSetPriority, map[string]string{"priority": "high"},
			ownerID,
		)
		ruleRepo.Create(ctx, rule)
	}

	rules, err := ruleRepo.ListByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}

	if len(rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(rules))
	}
}

func TestAutomationRuleRepository_ListEnabledByBoardAndTrigger(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create enabled rule with card_created trigger
	rule1, _ := domain.NewAutomationRule("", board.ID, "Enabled card_created",
		domain.TriggerCardCreated, nil, domain.ActionSetPriority, nil, ownerID)
	ruleRepo.Create(ctx, rule1)

	// Create disabled rule with card_created trigger
	rule2, _ := domain.NewAutomationRule("", board.ID, "Disabled card_created",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil, ownerID)
	ruleRepo.Create(ctx, rule2)
	rule2.Update("Disabled card_created", false, nil, nil)
	ruleRepo.Update(ctx, rule2)

	// Create enabled rule with different trigger
	rule3, _ := domain.NewAutomationRule("", board.ID, "Enabled label_added",
		domain.TriggerLabelAdded, nil, domain.ActionSetPriority, nil, ownerID)
	ruleRepo.Create(ctx, rule3)

	// Should only return 1 enabled card_created rule
	rules, err := ruleRepo.ListEnabledByBoardAndTrigger(ctx, board.ID, domain.TriggerCardCreated)
	if err != nil {
		t.Fatalf("Failed to list rules: %v", err)
	}

	if len(rules) != 1 {
		t.Errorf("Expected 1 enabled card_created rule, got %d", len(rules))
	}
}

func TestAutomationRuleRepository_Update(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	rule, _ := domain.NewAutomationRule("", board.ID, "Original",
		domain.TriggerCardCreated, map[string]string{"key": "val"},
		domain.ActionMoveCard, map[string]string{"target": "col-1"},
		ownerID,
	)
	ruleRepo.Create(ctx, rule)

	// Update
	rule.Update("Updated", false, map[string]string{"key": "new"}, map[string]string{"target": "col-2"})
	err := ruleRepo.Update(ctx, rule)
	if err != nil {
		t.Fatalf("Failed to update rule: %v", err)
	}

	loaded, _ := ruleRepo.GetByID(ctx, rule.ID)
	if loaded.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got %s", loaded.Name)
	}
	if loaded.Enabled {
		t.Error("Expected rule to be disabled")
	}
	if loaded.TriggerConfig["key"] != "new" {
		t.Errorf("Expected trigger_config key=new, got %s", loaded.TriggerConfig["key"])
	}
	if loaded.ActionConfig["target"] != "col-2" {
		t.Errorf("Expected action_config target=col-2, got %s", loaded.ActionConfig["target"])
	}
}

func TestAutomationRuleRepository_Delete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	rule, _ := domain.NewAutomationRule("", board.ID, "To Delete",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil, ownerID)
	ruleRepo.Create(ctx, rule)

	err := ruleRepo.Delete(ctx, rule.ID)
	if err != nil {
		t.Fatalf("Failed to delete rule: %v", err)
	}

	_, err = ruleRepo.GetByID(ctx, rule.ID)
	if err != domain.ErrAutomationRuleNotFound {
		t.Errorf("Expected ErrAutomationRuleNotFound after delete, got %v", err)
	}
}

func TestAutomationRuleRepository_CountByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	count, _ := ruleRepo.CountByBoardID(ctx, board.ID)
	if count != 0 {
		t.Errorf("Expected 0 rules, got %d", count)
	}

	for i := 0; i < 5; i++ {
		rule, _ := domain.NewAutomationRule("", board.ID, "Rule "+uuid.NewString()[:4],
			domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil, ownerID)
		ruleRepo.Create(ctx, rule)
	}

	count, _ = ruleRepo.CountByBoardID(ctx, board.ID)
	if count != 5 {
		t.Errorf("Expected 5 rules, got %d", count)
	}
}

func TestAutomationRuleRepository_CascadeDelete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	rule, _ := domain.NewAutomationRule("", board.ID, "Rule with executions",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil, ownerID)
	ruleRepo.Create(ctx, rule)

	// Create executions
	for i := 0; i < 3; i++ {
		exec := &domain.AutomationExecution{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			BoardID:    board.ID,
			Status:     "success",
			ExecutedAt: time.Now(),
		}
		ruleRepo.CreateExecution(ctx, exec)
	}

	// Delete board (should cascade to rules and executions)
	err := boardRepo.Delete(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete board: %v", err)
	}

	_, err = ruleRepo.GetByID(ctx, rule.ID)
	if err != domain.ErrAutomationRuleNotFound {
		t.Errorf("Expected ErrAutomationRuleNotFound after cascade delete, got %v", err)
	}
}

func TestAutomationRuleRepository_Executions(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	rule, _ := domain.NewAutomationRule("", board.ID, "Rule",
		domain.TriggerCardCreated, nil, domain.ActionMoveCard, nil, ownerID)
	ruleRepo.Create(ctx, rule)

	// Create executions
	exec1 := &domain.AutomationExecution{
		ID:         uuid.NewString(),
		RuleID:     rule.ID,
		BoardID:    board.ID,
		CardID:     uuid.NewString(),
		Status:     "success",
		ExecutedAt: time.Now().Add(-time.Minute),
	}
	ruleRepo.CreateExecution(ctx, exec1)

	exec2 := &domain.AutomationExecution{
		ID:           uuid.NewString(),
		RuleID:       rule.ID,
		BoardID:      board.ID,
		Status:       "failed",
		ErrorMessage: "card not found",
		ExecutedAt:   time.Now(),
	}
	ruleRepo.CreateExecution(ctx, exec2)

	// List executions
	execs, err := ruleRepo.ListExecutionsByRuleID(ctx, rule.ID, board.ID, 50)
	if err != nil {
		t.Fatalf("Failed to list executions: %v", err)
	}

	if len(execs) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(execs))
	}

	// Should be ordered by executed_at DESC (newest first)
	if execs[0].Status != "failed" {
		t.Errorf("Expected first execution to be 'failed' (newest), got %s", execs[0].Status)
	}
	if execs[1].Status != "success" {
		t.Errorf("Expected second execution to be 'success' (oldest), got %s", execs[1].Status)
	}
}
