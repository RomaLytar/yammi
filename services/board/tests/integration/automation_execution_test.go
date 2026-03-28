package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

func TestAutomationExecution_AssignMemberOnMoveToColumn(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()

	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	// 1. Создаём доску и добавляем участника
	board, _ := domain.NewBoard("Automation Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// 2. Создаём две колонки: "Todo" и "Done"
	colTodo, _ := domain.NewColumn(board.ID, "Todo", 0)
	columnRepo.Create(ctx, colTodo)
	colDone, _ := domain.NewColumn(board.ID, "Done", 1)
	columnRepo.Create(ctx, colDone)

	// 3. Создаём правило автоматизации: при перемещении в "Done" — назначить memberID
	rule, err := domain.NewAutomationRule("", board.ID, "Auto assign on done",
		domain.TriggerCardMovedToColumn,
		map[string]string{"column_id": colDone.ID},
		domain.ActionAssignMember,
		map[string]string{"user_id": memberID},
		ownerID,
	)
	if err != nil {
		t.Fatalf("Failed to create automation rule: %v", err)
	}
	if err := ruleRepo.Create(ctx, rule); err != nil {
		t.Fatalf("Failed to save automation rule: %v", err)
	}

	// 4. Создаём карточку в "Todo"
	card, _ := domain.NewCard(colTodo.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	if err := cardRepo.Create(ctx, card); err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// 5. Создаём executor и move usecase
	automationExecutor := usecase.NewExecuteAutomationsUseCase(ruleRepo, cardRepo, labelRepo)
	moveUC := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher, automationExecutor)

	// 6. Перемещаем карточку в "Done"
	_, err = moveUC.Execute(ctx, card.ID, board.ID, colTodo.ID, colDone.ID, ownerID, "m")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	// 7. Ждём async выполнения автоматизации
	time.Sleep(500 * time.Millisecond)

	// 8. Проверяем: карточка должна быть назначена на memberID
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	if loaded.AssigneeID == nil {
		t.Fatal("Expected card to have assignee after automation, but AssigneeID is nil")
	}
	if *loaded.AssigneeID != memberID {
		t.Errorf("Expected assignee %s, got %s", memberID, *loaded.AssigneeID)
	}

	// 9. Проверяем запись в automation_executions
	execs, err := ruleRepo.ListExecutionsByRuleID(ctx, rule.ID, board.ID, 10)
	if err != nil {
		t.Fatalf("Failed to list executions: %v", err)
	}
	if len(execs) < 1 {
		t.Fatal("Expected at least 1 execution record")
	}

	foundSuccess := false
	for _, exec := range execs {
		if exec.Status == "success" && exec.CardID == card.ID {
			foundSuccess = true
			break
		}
	}
	if !foundSuccess {
		t.Error("Expected to find a success execution record for the card")
	}
}

func TestAutomationExecution_NoMatchingColumn_NoAction(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()

	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("No Match Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	colTodo, _ := domain.NewColumn(board.ID, "Todo", 0)
	columnRepo.Create(ctx, colTodo)
	colInProgress, _ := domain.NewColumn(board.ID, "In Progress", 1)
	columnRepo.Create(ctx, colInProgress)
	colDone, _ := domain.NewColumn(board.ID, "Done", 2)
	columnRepo.Create(ctx, colDone)

	// Правило срабатывает только при перемещении в "Done"
	rule, _ := domain.NewAutomationRule("", board.ID, "Auto assign on done",
		domain.TriggerCardMovedToColumn,
		map[string]string{"column_id": colDone.ID},
		domain.ActionAssignMember,
		map[string]string{"user_id": memberID},
		ownerID,
	)
	ruleRepo.Create(ctx, rule)

	card, _ := domain.NewCard(colTodo.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	automationExecutor := usecase.NewExecuteAutomationsUseCase(ruleRepo, cardRepo, labelRepo)
	moveUC := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher, automationExecutor)

	// Перемещаем в "In Progress" (не в "Done") — правило НЕ должно сработать
	_, err := moveUC.Execute(ctx, card.ID, board.ID, colTodo.ID, colInProgress.ID, ownerID, "m")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Карточка не должна быть назначена
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}
	if loaded.AssigneeID != nil {
		t.Errorf("Expected no assignee (rule shouldn't match), got %s", *loaded.AssigneeID)
	}

	// Не должно быть записей о выполнении
	execs, err := ruleRepo.ListExecutionsByRuleID(ctx, rule.ID, board.ID, 10)
	if err != nil {
		t.Fatalf("Failed to list executions: %v", err)
	}
	if len(execs) != 0 {
		t.Errorf("Expected 0 execution records for non-matching column, got %d", len(execs))
	}
}

func TestAutomationExecution_AddLabelOnMoveToColumn(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	activityRepo := postgres.NewActivityRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ruleRepo := postgres.NewAutomationRuleRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()

	ownerID := uuid.NewString()

	board, _ := domain.NewBoard("Label Automation Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	colTodo, _ := domain.NewColumn(board.ID, "Todo", 0)
	columnRepo.Create(ctx, colTodo)
	colDone, _ := domain.NewColumn(board.ID, "Done", 1)
	columnRepo.Create(ctx, colDone)

	// Создаём метку
	label, _ := domain.NewLabel("", board.ID, "Completed", "#22c55e")
	if err := labelRepo.Create(ctx, label); err != nil {
		t.Fatalf("Failed to create label: %v", err)
	}

	// Правило: при перемещении в "Done" — добавить метку "Completed"
	rule, _ := domain.NewAutomationRule("", board.ID, "Auto label on done",
		domain.TriggerCardMovedToColumn,
		map[string]string{"column_id": colDone.ID},
		domain.ActionAddLabel,
		map[string]string{"label_id": label.ID},
		ownerID,
	)
	ruleRepo.Create(ctx, rule)

	card, _ := domain.NewCard(colTodo.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	automationExecutor := usecase.NewExecuteAutomationsUseCase(ruleRepo, cardRepo, labelRepo)
	moveUC := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher, automationExecutor)

	// Перемещаем карточку в "Done"
	_, err := moveUC.Execute(ctx, card.ID, board.ID, colTodo.ID, colDone.ID, ownerID, "m")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Проверяем: метка должна быть назначена на карточку
	labels, err := labelRepo.ListByCardID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list card labels: %v", err)
	}
	if len(labels) < 1 {
		t.Fatal("Expected at least 1 label on card after automation")
	}

	foundLabel := false
	for _, l := range labels {
		if l.ID == label.ID {
			foundLabel = true
			break
		}
	}
	if !foundLabel {
		t.Errorf("Expected label %s to be on card, but it wasn't", label.ID)
	}

	// Проверяем запись о выполнении
	execs, err := ruleRepo.ListExecutionsByRuleID(ctx, rule.ID, board.ID, 10)
	if err != nil {
		t.Fatalf("Failed to list executions: %v", err)
	}
	if len(execs) < 1 {
		t.Fatal("Expected at least 1 execution record")
	}
	if execs[0].Status != "success" {
		t.Errorf("Expected execution status 'success', got %s", execs[0].Status)
	}
}
