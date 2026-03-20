package integration

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestMembershipRepository_AddMember(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Add member
	err = memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Verify member exists
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, "user-456")
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if !isMember {
		t.Error("Expected user to be a member")
	}

	if role != domain.RoleMember {
		t.Errorf("Expected role member, got %s", role)
	}
}

func TestMembershipRepository_AddMember_Duplicate(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Add member
	memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)

	// Try to add same member again
	err = memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)
	if err != domain.ErrMemberExists {
		t.Errorf("Expected ErrMemberExists, got %v", err)
	}
}

func TestMembershipRepository_AddMember_InvalidRole(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Try to add member with invalid role
	err = memberRepo.AddMember(ctx, board.ID, "user-456", domain.Role("invalid"))
	if err != domain.ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got %v", err)
	}
}

func TestMembershipRepository_RemoveMember(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Add member
	memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)

	// Remove member
	err = memberRepo.RemoveMember(ctx, board.ID, "user-456")
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}

	// Verify member removed
	isMember, _, err := memberRepo.IsMember(ctx, board.ID, "user-456")
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if isMember {
		t.Error("Expected user to not be a member")
	}
}

func TestMembershipRepository_RemoveMember_NotFound(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Try to remove non-existent member
	err = memberRepo.RemoveMember(ctx, board.ID, "non-existent-user")
	if err != domain.ErrMemberNotFound {
		t.Errorf("Expected ErrMemberNotFound, got %v", err)
	}
}

func TestMembershipRepository_RemoveMember_CannotRemoveOwner(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board (owner automatically added)
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Try to remove owner
	err = memberRepo.RemoveMember(ctx, board.ID, "owner-123")
	if err != domain.ErrCannotRemoveOwner {
		t.Errorf("Expected ErrCannotRemoveOwner, got %v", err)
	}
}

func TestMembershipRepository_IsMember(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Check owner is member
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, "owner-123")
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if !isMember {
		t.Error("Expected owner to be a member")
	}

	if role != domain.RoleOwner {
		t.Errorf("Expected role owner, got %s", role)
	}

	// Check non-member
	isMember, _, err = memberRepo.IsMember(ctx, board.ID, "non-member")
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if isMember {
		t.Error("Expected non-member to not be a member")
	}
}

func TestMembershipRepository_ListMembers(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Add members
	memberRepo.AddMember(ctx, board.ID, "user-1", domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, "user-2", domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, "user-3", domain.RoleMember)

	// List members (limit 10, offset 0)
	members, err := memberRepo.ListMembers(ctx, board.ID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list members: %v", err)
	}

	// Should have 4 members (owner + 3 added)
	if len(members) != 4 {
		t.Errorf("Expected 4 members, got %d", len(members))
	}

	// Verify owner is first (ordered by joined_at)
	if members[0].UserID != "owner-123" {
		t.Errorf("Expected first member to be owner-123, got %s", members[0].UserID)
	}

	if members[0].Role != domain.RoleOwner {
		t.Errorf("Expected first member role to be owner, got %s", members[0].Role)
	}
}

func TestMembershipRepository_ListMembers_Pagination(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	board, _ := domain.NewBoard("Test Board", "Desc", "owner-123")
	boardRepo.Create(ctx, board)

	// Add 10 members
	for i := 1; i <= 10; i++ {
		memberRepo.AddMember(ctx, board.ID, fmt.Sprintf("user-%d", i), domain.RoleMember)
	}

	// Page 1 (limit 5, offset 0)
	page1, err := memberRepo.ListMembers(ctx, board.ID, 5, 0)
	if err != nil {
		t.Fatalf("Failed to list members (page 1): %v", err)
	}

	if len(page1) != 5 {
		t.Errorf("Expected 5 members on page 1, got %d", len(page1))
	}

	// Page 2 (limit 5, offset 5)
	page2, err := memberRepo.ListMembers(ctx, board.ID, 5, 5)
	if err != nil {
		t.Fatalf("Failed to list members (page 2): %v", err)
	}

	if len(page2) != 5 {
		t.Errorf("Expected 5 members on page 2, got %d", len(page2))
	}

	// Verify no duplicates
	ids := make(map[string]bool)
	for _, m := range page1 {
		ids[m.UserID] = true
	}
	for _, m := range page2 {
		if ids[m.UserID] {
			t.Errorf("Duplicate member ID %s in pagination", m.UserID)
		}
	}
}
