package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestBoardRepository_Create(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, err := domain.NewBoard("Test Board", "Description", ownerID)
	if err != nil {
		t.Fatalf("Failed to create domain board: %v", err)
	}

	err = repo.Create(ctx, board)
	if err != nil {
		t.Fatalf("Failed to save board: %v", err)
	}

	// Verify board exists
	loaded, err := repo.GetByID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	if loaded.Title != board.Title {
		t.Errorf("Expected title %s, got %s", board.Title, loaded.Title)
	}

	if loaded.Description != board.Description {
		t.Errorf("Expected description %s, got %s", board.Description, loaded.Description)
	}

	if loaded.OwnerID != board.OwnerID {
		t.Errorf("Expected owner ID %s, got %s", board.OwnerID, loaded.OwnerID)
	}

	if loaded.Version != 1 {
		t.Errorf("Expected version 1, got %d", loaded.Version)
	}

	// Verify owner is in board_members
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, ownerID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if !isMember {
		t.Error("Owner should be a member")
	}

	if role != domain.RoleOwner {
		t.Errorf("Expected role owner, got %s", role)
	}
}

func TestBoardRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.NewString())
	if err != domain.ErrBoardNotFound {
		t.Errorf("Expected ErrBoardNotFound, got %v", err)
	}
}

func TestBoardRepository_Update(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Original Title", "Original Description", ownerID)
	repo.Create(ctx, board)

	// Update board
	err := board.Update("Updated Title", "Updated Description")
	if err != nil {
		t.Fatalf("Failed to update domain board: %v", err)
	}

	err = repo.Update(ctx, board)
	if err != nil {
		t.Fatalf("Failed to save updated board: %v", err)
	}

	// Verify updates
	loaded, _ := repo.GetByID(ctx, board.ID)
	if loaded.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", loaded.Title)
	}

	if loaded.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %s", loaded.Description)
	}

	if loaded.Version != 2 {
		t.Errorf("Expected version 2, got %d", loaded.Version)
	}
}

func TestBoardRepository_OptimisticLocking(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test", "Desc", ownerID)
	repo.Create(ctx, board)

	// Load board twice (simulate concurrent access)
	board1, _ := repo.GetByID(ctx, board.ID)
	board2, _ := repo.GetByID(ctx, board.ID)

	// Update board1 (version 1 → 2)
	board1.Update("Updated Title 1", "Desc 1")
	err := repo.Update(ctx, board1)
	if err != nil {
		t.Fatalf("First update should succeed: %v", err)
	}

	// Verify board1 version is now 2
	if board1.Version != 2 {
		t.Errorf("Expected version 2, got %d", board1.Version)
	}

	// Update board2 (still version 1, should fail)
	board2.Update("Updated Title 2", "Desc 2")
	err = repo.Update(ctx, board2)
	if err != domain.ErrInvalidVersion {
		t.Errorf("Expected ErrInvalidVersion, got %v", err)
	}
}

func TestBoardRepository_Delete(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test", "Desc", ownerID)
	repo.Create(ctx, board)

	// Delete board
	err := repo.Delete(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete board: %v", err)
	}

	// Verify deleted
	_, err = repo.GetByID(ctx, board.ID)
	if err != domain.ErrBoardNotFound {
		t.Errorf("Expected ErrBoardNotFound after delete, got %v", err)
	}

	// Try to delete again
	err = repo.Delete(ctx, board.ID)
	if err != domain.ErrBoardNotFound {
		t.Errorf("Expected ErrBoardNotFound on second delete, got %v", err)
	}
}

func TestBoardRepository_CursorPagination(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	ownerID := uuid.NewString()
	userID := uuid.NewString()

	// Create 25 boards
	for i := 0; i < 25; i++ {
		board, _ := domain.NewBoard(fmt.Sprintf("Board %d", i), "", ownerID)
		repo.Create(ctx, board)
		memberRepo.AddMember(ctx, board.ID, userID, domain.RoleMember)
		time.Sleep(1 * time.Millisecond) // Ensure different created_at
	}

	// Page 1 (limit 10)
	boards, cursor, err := repo.ListByUserID(ctx, userID, 10, "", false, "", "updated_at")
	if err != nil {
		t.Fatalf("Failed to list boards: %v", err)
	}

	if len(boards) != 10 {
		t.Errorf("Expected 10 boards, got %d", len(boards))
	}

	if cursor == "" {
		t.Error("Expected cursor, got empty")
	}

	// Page 2 (using cursor)
	boards2, cursor2, err := repo.ListByUserID(ctx, userID, 10, cursor, false, "", "updated_at")
	if err != nil {
		t.Fatalf("Failed to list boards (page 2): %v", err)
	}

	if len(boards2) != 10 {
		t.Errorf("Expected 10 boards (page 2), got %d", len(boards2))
	}

	if cursor2 == "" {
		t.Error("Expected cursor2, got empty")
	}

	// Verify no duplicates
	ids := make(map[string]bool)
	for _, b := range boards {
		ids[b.ID] = true
	}
	for _, b := range boards2 {
		if ids[b.ID] {
			t.Errorf("Duplicate board ID %s in pagination", b.ID)
		}
	}

	// Page 3 (last page, should have 5 boards)
	boards3, cursor3, err := repo.ListByUserID(ctx, userID, 10, cursor2, false, "", "updated_at")
	if err != nil {
		t.Fatalf("Failed to list boards (page 3): %v", err)
	}

	if len(boards3) != 5 {
		t.Errorf("Expected 5 boards (page 3), got %d", len(boards3))
	}

	if cursor3 != "" {
		t.Error("Expected empty cursor on last page, got non-empty")
	}
}

func TestBoardRepository_ListByUserID_EmptyResult(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	repo := postgres.NewBoardRepository(db)
	ctx := context.Background()

	// List boards for user without any boards
	nonExistentUserID := uuid.NewString()
	boards, cursor, err := repo.ListByUserID(ctx, nonExistentUserID, 10, "", false, "", "updated_at")
	if err != nil {
		t.Fatalf("Failed to list boards: %v", err)
	}

	if len(boards) != 0 {
		t.Errorf("Expected 0 boards, got %d", len(boards))
	}

	if cursor != "" {
		t.Error("Expected empty cursor, got non-empty")
	}
}
