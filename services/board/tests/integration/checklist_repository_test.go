package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestChecklistRepository_CreateChecklist(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Create checklist
	checklist, err := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	if err != nil {
		t.Fatalf("Failed to create domain checklist: %v", err)
	}

	err = checklistRepo.CreateChecklist(ctx, checklist)
	if err != nil {
		t.Fatalf("Failed to save checklist: %v", err)
	}

	// Verify checklist exists
	loaded, err := checklistRepo.GetChecklistByID(ctx, checklist.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load checklist: %v", err)
	}

	if loaded.Title != "Review Tasks" {
		t.Errorf("Expected title Review Tasks, got %s", loaded.Title)
	}

	if loaded.CardID != card.ID {
		t.Errorf("Expected card ID %s, got %s", card.ID, loaded.CardID)
	}

	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestChecklistRepository_GetChecklistByID_NotFound(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	_, err := checklistRepo.GetChecklistByID(ctx, uuid.NewString(), uuid.NewString())
	if err != domain.ErrChecklistNotFound {
		t.Errorf("Expected ErrChecklistNotFound, got %v", err)
	}
}

func TestChecklistRepository_ListByCardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Create multiple checklists
	titles := []string{"Review Tasks", "Deploy Tasks", "QA Tasks"}
	for i, title := range titles {
		cl, _ := domain.NewChecklist("", card.ID, board.ID, title, i)
		checklistRepo.CreateChecklist(ctx, cl)
	}

	// List checklists
	checklists, err := checklistRepo.ListByCardID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list checklists: %v", err)
	}

	if len(checklists) != 3 {
		t.Errorf("Expected 3 checklists, got %d", len(checklists))
	}
}

func TestChecklistRepository_UpdateChecklist(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	// Update checklist
	checklist.Update("Deploy Tasks")
	err := checklistRepo.UpdateChecklist(ctx, checklist)
	if err != nil {
		t.Fatalf("Failed to update checklist: %v", err)
	}

	// Verify updates
	loaded, _ := checklistRepo.GetChecklistByID(ctx, checklist.ID, board.ID)
	if loaded.Title != "Deploy Tasks" {
		t.Errorf("Expected title Deploy Tasks, got %s", loaded.Title)
	}
}

func TestChecklistRepository_DeleteChecklist_CascadeItems(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	// Add items
	item1, _ := domain.NewChecklistItem("", checklist.ID, board.ID, "Item 1", 0)
	checklistRepo.CreateItem(ctx, item1)
	item2, _ := domain.NewChecklistItem("", checklist.ID, board.ID, "Item 2", 1)
	checklistRepo.CreateItem(ctx, item2)

	// Delete checklist (should cascade delete items)
	err := checklistRepo.DeleteChecklist(ctx, checklist.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete checklist: %v", err)
	}

	// Verify checklist is deleted
	_, err = checklistRepo.GetChecklistByID(ctx, checklist.ID, board.ID)
	if err != domain.ErrChecklistNotFound {
		t.Errorf("Expected ErrChecklistNotFound after delete, got %v", err)
	}

	// Verify items are deleted
	items, err := checklistRepo.ListItemsByChecklistID(ctx, checklist.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list items after delete: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items after cascade delete, got %d", len(items))
	}
}

func TestChecklistRepository_CreateItem(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	// Create item
	item, _ := domain.NewChecklistItem("", checklist.ID, board.ID, "Write tests", 0)
	err := checklistRepo.CreateItem(ctx, item)
	if err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	// Verify item exists
	loaded, err := checklistRepo.GetItemByID(ctx, item.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load item: %v", err)
	}

	if loaded.Title != "Write tests" {
		t.Errorf("Expected title Write tests, got %s", loaded.Title)
	}

	if loaded.IsChecked {
		t.Error("Expected IsChecked to be false by default")
	}
}

func TestChecklistRepository_ToggleItem(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist, item
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	item, _ := domain.NewChecklistItem("", checklist.ID, board.ID, "Write tests", 0)
	checklistRepo.CreateItem(ctx, item)

	// Toggle on
	err := checklistRepo.ToggleItem(ctx, item.ID, board.ID, true)
	if err != nil {
		t.Fatalf("Failed to toggle item: %v", err)
	}

	loaded, _ := checklistRepo.GetItemByID(ctx, item.ID, board.ID)
	if !loaded.IsChecked {
		t.Error("Expected IsChecked to be true after toggle")
	}

	// Toggle off
	err = checklistRepo.ToggleItem(ctx, item.ID, board.ID, false)
	if err != nil {
		t.Fatalf("Failed to toggle item off: %v", err)
	}

	loaded, _ = checklistRepo.GetItemByID(ctx, item.ID, board.ID)
	if loaded.IsChecked {
		t.Error("Expected IsChecked to be false after toggle off")
	}
}

func TestChecklistRepository_DeleteItem(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist, item
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	item, _ := domain.NewChecklistItem("", checklist.ID, board.ID, "Write tests", 0)
	checklistRepo.CreateItem(ctx, item)

	// Delete item
	err := checklistRepo.DeleteItem(ctx, item.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Verify deleted
	_, err = checklistRepo.GetItemByID(ctx, item.ID, board.ID)
	if err != domain.ErrChecklistItemNotFound {
		t.Errorf("Expected ErrChecklistItemNotFound after delete, got %v", err)
	}
}

func TestChecklistRepository_ListItemsByChecklistID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	checklistRepo := postgres.NewChecklistRepository(db)
	ctx := context.Background()

	// Create board, column, card, checklist
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	checklist, _ := domain.NewChecklist("", card.ID, board.ID, "Review Tasks", 0)
	checklistRepo.CreateChecklist(ctx, checklist)

	// Create items
	titles := []string{"Write tests", "Code review", "Update docs"}
	for i, title := range titles {
		item, _ := domain.NewChecklistItem("", checklist.ID, board.ID, title, i)
		checklistRepo.CreateItem(ctx, item)
	}

	// List items
	items, err := checklistRepo.ListItemsByChecklistID(ctx, checklist.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list items: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Verify ordering
	if items[0].Title != "Write tests" {
		t.Errorf("Expected first item title Write tests, got %s", items[0].Title)
	}
	if items[1].Title != "Code review" {
		t.Errorf("Expected second item title Code review, got %s", items[1].Title)
	}
	if items[2].Title != "Update docs" {
		t.Errorf("Expected third item title Update docs, got %s", items[2].Title)
	}
}
