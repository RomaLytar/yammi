package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestLabelRepository_Create(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create label
	label, err := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create domain label: %v", err)
	}

	err = labelRepo.Create(ctx, label)
	if err != nil {
		t.Fatalf("Failed to save label: %v", err)
	}

	// Verify label exists
	loaded, err := labelRepo.GetByID(ctx, label.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load label: %v", err)
	}

	if loaded.Name != "Bug" {
		t.Errorf("Expected name Bug, got %s", loaded.Name)
	}

	if loaded.Color != "#ef4444" {
		t.Errorf("Expected color #ef4444, got %s", loaded.Color)
	}

	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestLabelRepository_Create_Duplicate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create first label
	label1, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label1)

	// Try to create duplicate (same name, same board)
	label2, _ := domain.NewLabel("", board.ID, "Bug", "#3b82f6")
	err := labelRepo.Create(ctx, label2)

	if err != domain.ErrLabelExists {
		t.Errorf("Expected ErrLabelExists, got %v", err)
	}
}

func TestLabelRepository_ListByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create labels
	names := []string{"Bug", "Feature", "Enhancement"}
	for _, name := range names {
		label, _ := domain.NewLabel("", board.ID, name, "#6b7280")
		labelRepo.Create(ctx, label)
	}

	// List labels
	labels, err := labelRepo.ListByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to list labels: %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
}

func TestLabelRepository_Update(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board and label
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	label, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label)

	// Update label
	label.Update("Feature", "#3b82f6")
	err := labelRepo.Update(ctx, label)
	if err != nil {
		t.Fatalf("Failed to update label: %v", err)
	}

	// Verify updates
	loaded, _ := labelRepo.GetByID(ctx, label.ID, board.ID)
	if loaded.Name != "Feature" {
		t.Errorf("Expected name Feature, got %s", loaded.Name)
	}
	if loaded.Color != "#3b82f6" {
		t.Errorf("Expected color #3b82f6, got %s", loaded.Color)
	}
}

func TestLabelRepository_Delete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board and label
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	label, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label)

	// Delete label
	err := labelRepo.Delete(ctx, label.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete label: %v", err)
	}

	// Verify deleted
	_, err = labelRepo.GetByID(ctx, label.ID, board.ID)
	if err != domain.ErrLabelNotFound {
		t.Errorf("Expected ErrLabelNotFound after delete, got %v", err)
	}
}

func TestLabelRepository_AddToCard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board, column, card, label
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	label, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label)

	// Add label to card
	err := labelRepo.AddToCard(ctx, card.ID, board.ID, label.ID)
	if err != nil {
		t.Fatalf("Failed to add label to card: %v", err)
	}

	// Verify label is on card
	labels, err := labelRepo.ListByCardID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list card labels: %v", err)
	}

	if len(labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(labels))
	}

	if labels[0].ID != label.ID {
		t.Errorf("Expected label ID %s, got %s", label.ID, labels[0].ID)
	}
}

func TestLabelRepository_AddToCard_Duplicate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board, column, card, label
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	label, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label)

	// Add label to card (first time)
	labelRepo.AddToCard(ctx, card.ID, board.ID, label.ID)

	// Try to add same label again
	err := labelRepo.AddToCard(ctx, card.ID, board.ID, label.ID)
	if err != domain.ErrLabelAlreadyOnCard {
		t.Errorf("Expected ErrLabelAlreadyOnCard, got %v", err)
	}
}

func TestLabelRepository_RemoveFromCard(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board, column, card, label
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	label, _ := domain.NewLabel("", board.ID, "Bug", "#ef4444")
	labelRepo.Create(ctx, label)

	// Add and then remove label from card
	labelRepo.AddToCard(ctx, card.ID, board.ID, label.ID)
	err := labelRepo.RemoveFromCard(ctx, card.ID, board.ID, label.ID)
	if err != nil {
		t.Fatalf("Failed to remove label from card: %v", err)
	}

	// Verify label is removed
	labels, _ := labelRepo.ListByCardID(ctx, card.ID, board.ID)
	if len(labels) != 0 {
		t.Errorf("Expected 0 labels after removal, got %d", len(labels))
	}
}

func TestLabelRepository_ListByCardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board, column, card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task 1", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Create multiple labels and assign to card
	names := []string{"Bug", "Feature", "Enhancement"}
	for _, name := range names {
		label, _ := domain.NewLabel("", board.ID, name, "#6b7280")
		labelRepo.Create(ctx, label)
		labelRepo.AddToCard(ctx, card.ID, board.ID, label.ID)
	}

	// List card labels
	labels, err := labelRepo.ListByCardID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list card labels: %v", err)
	}

	if len(labels) != 3 {
		t.Errorf("Expected 3 labels on card, got %d", len(labels))
	}
}

func TestLabelRepository_CountByBoardID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	labelRepo := postgres.NewLabelRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Initially 0 labels
	count, err := labelRepo.CountByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to count labels: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 labels, got %d", count)
	}

	// Create labels
	for i := 0; i < 5; i++ {
		label, _ := domain.NewLabel("", board.ID, uuid.NewString(), "#6b7280")
		labelRepo.Create(ctx, label)
	}

	// Count should be 5
	count, err = labelRepo.CountByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to count labels: %v", err)
	}
	if count != 5 {
		t.Errorf("Expected 5 labels, got %d", count)
	}
}
