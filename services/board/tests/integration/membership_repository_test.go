package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestMembershipRepository_AddMember(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add member
	userID := uuid.NewString()
	err := memberRepo.AddMember(ctx, board.ID, userID, domain.RoleMember)
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Verify member exists
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, userID)
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
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add member
	userID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, userID, domain.RoleMember)

	// Try to add same member again
	err := memberRepo.AddMember(ctx, board.ID, userID, domain.RoleMember)
	if err != domain.ErrMemberExists {
		t.Errorf("Expected ErrMemberExists, got %v", err)
	}
}

func TestMembershipRepository_AddMember_InvalidRole(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Try to add member with invalid role
	userID := uuid.NewString()
	err := memberRepo.AddMember(ctx, board.ID, userID, domain.Role("invalid"))
	if err != domain.ErrInvalidRole {
		t.Errorf("Expected ErrInvalidRole, got %v", err)
	}
}

func TestMembershipRepository_RemoveMember(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add member
	userID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, userID, domain.RoleMember)

	// Remove member
	err := memberRepo.RemoveMember(ctx, board.ID, userID)
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}

	// Verify member removed
	isMember, _, err := memberRepo.IsMember(ctx, board.ID, userID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if isMember {
		t.Error("Expected user to not be a member")
	}
}

func TestMembershipRepository_RemoveMember_NotFound(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Try to remove non-existent member
	nonExistentUserID := uuid.NewString()
	err := memberRepo.RemoveMember(ctx, board.ID, nonExistentUserID)
	if err != domain.ErrMemberNotFound {
		t.Errorf("Expected ErrMemberNotFound, got %v", err)
	}
}

func TestMembershipRepository_RemoveMember_CannotRemoveOwner(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board (owner automatically added)
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Try to remove owner
	err := memberRepo.RemoveMember(ctx, board.ID, ownerID)
	if err != domain.ErrCannotRemoveOwner {
		t.Errorf("Expected ErrCannotRemoveOwner, got %v", err)
	}
}

func TestMembershipRepository_IsMember(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Check owner is member
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, ownerID)
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
	nonMemberID := uuid.NewString()
	isMember, _, err = memberRepo.IsMember(ctx, board.ID, nonMemberID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if isMember {
		t.Error("Expected non-member to not be a member")
	}
}

func TestMembershipRepository_ListMembers(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add members
	user1ID := uuid.NewString()
	user2ID := uuid.NewString()
	user3ID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, user1ID, domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, user2ID, domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, user3ID, domain.RoleMember)

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
	if members[0].UserID != ownerID {
		t.Errorf("Expected first member to be owner, got %s", members[0].UserID)
	}

	if members[0].Role != domain.RoleOwner {
		t.Errorf("Expected first member role to be owner, got %s", members[0].Role)
	}
}

func TestMembershipRepository_ListMembers_Pagination(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)
	ctx := context.Background()

	// Create board
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add 10 members
	for i := 1; i <= 10; i++ {
		memberRepo.AddMember(ctx, board.ID, uuid.NewString(), domain.RoleMember)
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
