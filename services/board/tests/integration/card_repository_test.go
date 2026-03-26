package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestCardRepository_Create(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Create card
	assignee := uuid.NewString()
	card, err := domain.NewCard(column.ID, "Task 1", "Description", "n", &assignee, ownerID, nil, "", "")
	if err != nil {
		t.Fatalf("Failed to create domain card: %v", err)
	}

	err = cardRepo.Create(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Verify card exists
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	if loaded.Title != card.Title {
		t.Errorf("Expected title %s, got %s", card.Title, loaded.Title)
	}

	if loaded.Position != card.Position {
		t.Errorf("Expected position %s, got %s", card.Position, loaded.Position)
	}

	if loaded.ColumnID != column.ID {
		t.Errorf("Expected column ID %s, got %s", column.ID, loaded.ColumnID)
	}

	if loaded.AssigneeID == nil || *loaded.AssigneeID != assignee {
		t.Errorf("Expected assignee %s, got %v", assignee, loaded.AssigneeID)
	}
}

func TestCardRepository_CreateWithoutAssignee(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Create card without assignee
	card, err := domain.NewCard(column.ID, "Task 1", "Description", "n", nil, ownerID, nil, "", "")
	if err != nil {
		t.Fatalf("Failed to create domain card: %v", err)
	}

	err = cardRepo.Create(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Verify card has no assignee
	loaded, _ := cardRepo.GetByID(ctx, card.ID, board.ID)
	if loaded.AssigneeID != nil {
		t.Errorf("Expected nil assignee, got %v", loaded.AssigneeID)
	}
}

func TestCardRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	_, err := cardRepo.GetByID(ctx, uuid.NewString(), uuid.NewString())
	if err != domain.ErrCardNotFound {
		t.Errorf("Expected ErrCardNotFound, got %v", err)
	}
}

func TestCardRepository_ListByColumnID(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Create cards with lexorank positions
	positions := []string{"a", "m", "z"}
	for i, pos := range positions {
		card, _ := domain.NewCard(column.ID, fmt.Sprintf("Task %d", i), "", pos, nil, ownerID, nil, "", "")
		cardRepo.Create(ctx, card)
	}

	// List cards
	loaded, err := cardRepo.ListByColumnID(ctx, column.ID)
	if err != nil {
		t.Fatalf("Failed to list cards: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(loaded))
	}

	// Verify lexicographic ordering
	for i, card := range loaded {
		if card.Position != positions[i] {
			t.Errorf("Expected position %s at index %d, got %s", positions[i], i, card.Position)
		}
	}
}

func TestCardRepository_LexorankPositioning(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Test lexorank ordering with complex positions
	testCases := []struct {
		title    string
		position string
	}{
		{"First", "a"},
		{"Second", "am"},
		{"Third", "b"},
		{"Fourth", "c"},
		{"Fifth", "m"},
		{"Sixth", "z"},
	}

	for _, tc := range testCases {
		card, _ := domain.NewCard(column.ID, tc.title, "", tc.position, nil, ownerID, nil, "", "")
		cardRepo.Create(ctx, card)
	}

	// Load cards and verify ORDER BY position works correctly
	cards, err := cardRepo.ListByColumnID(ctx, column.ID)
	if err != nil {
		t.Fatalf("Failed to list cards: %v", err)
	}

	if len(cards) != len(testCases) {
		t.Errorf("Expected %d cards, got %d", len(testCases), len(cards))
	}

	// Verify lexicographic order
	for i, card := range cards {
		if card.Title != testCases[i].title {
			t.Errorf("Expected card %s at position %d, got %s", testCases[i].title, i, card.Title)
		}
	}
}

func TestCardRepository_Update(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Create card
	card, _ := domain.NewCard(column.ID, "Original Title", "Original Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Update card
	newAssignee := uuid.NewString()
	err := card.Update("Updated Title", "Updated Desc", &newAssignee, nil, "", "")
	if err != nil {
		t.Fatalf("Failed to update domain card: %v", err)
	}

	err = cardRepo.Update(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save updated card: %v", err)
	}

	// Verify updates
	loaded, _ := cardRepo.GetByID(ctx, card.ID, board.ID)
	if loaded.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", loaded.Title)
	}

	if loaded.Description != "Updated Desc" {
		t.Errorf("Expected description 'Updated Desc', got %s", loaded.Description)
	}

	if loaded.AssigneeID == nil || *loaded.AssigneeID != newAssignee {
		t.Errorf("Expected assignee %s, got %v", newAssignee, loaded.AssigneeID)
	}
}

func TestCardRepository_Move(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board and two columns
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column1, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column1)

	column2, _ := domain.NewColumn(board.ID, "In Progress", 1)
	columnRepo.Create(ctx, column2)

	// Create card in column1
	card, _ := domain.NewCard(column1.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Move card to column2
	err := card.Move(column2.ID, "m")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	err = cardRepo.Update(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save moved card: %v", err)
	}

	// Verify card moved
	loaded, _ := cardRepo.GetByID(ctx, card.ID, board.ID)
	if loaded.ColumnID != column2.ID {
		t.Errorf("Expected column ID %s, got %s", column2.ID, loaded.ColumnID)
	}

	if loaded.Position != "m" {
		t.Errorf("Expected position 'm', got %s", loaded.Position)
	}
}

func TestCardRepository_Delete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create board, column, and card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Delete card
	err := cardRepo.Delete(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete card: %v", err)
	}

	// Verify deleted
	_, err = cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != domain.ErrCardNotFound {
		t.Errorf("Expected ErrCardNotFound after delete, got %v", err)
	}
}

func TestCardRepository_Partitioning(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	ctx := context.Background()

	// Create 10 boards with 10 cards each
	ownerID := uuid.NewString()
	totalCards := 0
	for i := 0; i < 10; i++ {
		board, _ := domain.NewBoard(fmt.Sprintf("Board %d", i), "Desc", ownerID)
		boardRepo.Create(ctx, board)

		column, _ := domain.NewColumn(board.ID, "To Do", 0)
		columnRepo.Create(ctx, column)

		for j := 0; j < 10; j++ {
			card, _ := domain.NewCard(column.ID, fmt.Sprintf("Card %d", j), "", "n", nil, ownerID, nil, "", "")
			cardRepo.Create(ctx, card)
			totalCards++
		}
	}

	// Query partition distribution (filter by creator_id to isolate from other parallel tests)
	query := `SELECT tableoid::regclass, COUNT(*) FROM cards WHERE creator_id = $1 GROUP BY tableoid ORDER BY tableoid`
	rows, err := db.Query(query, ownerID)
	if err != nil {
		t.Fatalf("Failed to query partitions: %v", err)
	}
	defer rows.Close()

	partitionCounts := make(map[string]int)
	totalFromPartitions := 0

	for rows.Next() {
		var partition string
		var count int
		if err := rows.Scan(&partition, &count); err != nil {
			t.Fatalf("Failed to scan partition row: %v", err)
		}
		partitionCounts[partition] = count
		totalFromPartitions += count
	}

	// Verify all partitions are used
	expectedPartitions := []string{"cards_p0", "cards_p1", "cards_p2", "cards_p3"}
	for _, p := range expectedPartitions {
		if _, exists := partitionCounts[p]; !exists {
			t.Errorf("Partition %s has no cards", p)
		}
	}

	// Verify total count matches
	if totalFromPartitions != totalCards {
		t.Errorf("Expected %d total cards, got %d", totalCards, totalFromPartitions)
	}

	// Log distribution for debugging
	t.Logf("Partition distribution:")
	for partition, count := range partitionCounts {
		t.Logf("  %s: %d cards", partition, count)
	}
}
