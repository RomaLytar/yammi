package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestColumnRepository_Create(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	// Create board first
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create column
	column, err := domain.NewColumn(board.ID, "To Do", 0)
	if err != nil {
		t.Fatalf("Failed to create domain column: %v", err)
	}

	err = columnRepo.Create(ctx, column)
	if err != nil {
		t.Fatalf("Failed to save column: %v", err)
	}

	// Verify column exists
	loaded, err := columnRepo.GetByID(ctx, column.ID)
	if err != nil {
		t.Fatalf("Failed to load column: %v", err)
	}

	if loaded.Title != column.Title {
		t.Errorf("Expected title %s, got %s", column.Title, loaded.Title)
	}

	if loaded.Position != column.Position {
		t.Errorf("Expected position %d, got %d", column.Position, loaded.Position)
	}

	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestColumnRepository_GetByID_NotFound(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	_, err = columnRepo.GetByID(ctx, uuid.NewString())
	if err != domain.ErrColumnNotFound {
		t.Errorf("Expected ErrColumnNotFound, got %v", err)
	}
}

func TestColumnRepository_ListByBoardID(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create columns with different positions
	columns := []struct {
		title    string
		position int
	}{
		{"To Do", 0},
		{"In Progress", 1},
		{"Done", 2},
	}

	for _, c := range columns {
		column, _ := domain.NewColumn(board.ID, c.title, c.position)
		columnRepo.Create(ctx, column)
	}

	// List columns
	loaded, err := columnRepo.ListByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to list columns: %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(loaded))
	}

	// Verify order by position
	for i, col := range loaded {
		if col.Title != columns[i].title {
			t.Errorf("Expected title %s at position %d, got %s", columns[i].title, i, col.Title)
		}
		if col.Position != columns[i].position {
			t.Errorf("Expected position %d at index %d, got %d", columns[i].position, i, col.Position)
		}
	}
}

func TestColumnRepository_Update(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Update title
	err = column.Update("Backlog")
	if err != nil {
		t.Fatalf("Failed to update domain column: %v", err)
	}

	err = columnRepo.Update(ctx, column)
	if err != nil {
		t.Fatalf("Failed to save updated column: %v", err)
	}

	// Verify title update
	loaded, _ := columnRepo.GetByID(ctx, column.ID)
	if loaded.Title != "Backlog" {
		t.Errorf("Expected title 'Backlog', got %s", loaded.Title)
	}

	// Update position
	err = column.UpdatePosition(5)
	if err != nil {
		t.Fatalf("Failed to update column position: %v", err)
	}

	err = columnRepo.Update(ctx, column)
	if err != nil {
		t.Fatalf("Failed to save updated position: %v", err)
	}

	// Verify position update
	loaded, _ = columnRepo.GetByID(ctx, column.ID)
	if loaded.Position != 5 {
		t.Errorf("Expected position 5, got %d", loaded.Position)
	}
}

func TestColumnRepository_Delete(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	// Create board and column
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Delete column
	err = columnRepo.Delete(ctx, column.ID)
	if err != nil {
		t.Fatalf("Failed to delete column: %v", err)
	}

	// Verify deleted
	_, err = columnRepo.GetByID(ctx, column.ID)
	if err != domain.ErrColumnNotFound {
		t.Errorf("Expected ErrColumnNotFound after delete, got %v", err)
	}

	// Try to delete again
	err = columnRepo.Delete(ctx, column.ID)
	if err != domain.ErrColumnNotFound {
		t.Errorf("Expected ErrColumnNotFound on second delete, got %v", err)
	}
}

func TestColumnRepository_CascadeDelete(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Create multiple columns
	for i := 0; i < 3; i++ {
		column, _ := domain.NewColumn(board.ID, "Column", i)
		columnRepo.Create(ctx, column)
	}

	// Delete board (should cascade delete all columns)
	err = boardRepo.Delete(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete board: %v", err)
	}

	// Verify all columns are deleted
	columns, err := columnRepo.ListByBoardID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to list columns: %v", err)
	}

	if len(columns) != 0 {
		t.Errorf("Expected 0 columns after cascade delete, got %d", len(columns))
	}
}
