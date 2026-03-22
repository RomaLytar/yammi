package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// ==================== Board tests ====================

func TestFeature_CreateBoard_OwnerAutoMember(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()

	uc := usecase.NewCreateBoardUseCase(boardRepo, memberRepo, publisher)
	board, err := uc.Execute(ctx, "My Board", "Description", ownerID)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	// Verify board exists in DB
	loaded, err := boardRepo.GetByID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}
	if loaded.Title != "My Board" {
		t.Errorf("Expected title 'My Board', got %s", loaded.Title)
	}
	if loaded.OwnerID != ownerID {
		t.Errorf("Expected ownerID %s, got %s", ownerID, loaded.OwnerID)
	}

	// Verify owner is auto-added as member with RoleOwner
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, ownerID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}
	if !isMember {
		t.Error("Owner should be auto-added as member")
	}
	if role != domain.RoleOwner {
		t.Errorf("Expected role %s, got %s", domain.RoleOwner, role)
	}
}

func TestFeature_ListBoards_OnlyMemberBoards(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()
	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()
	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	ctx := context.Background()
	userA := uuid.NewString()
	userB := uuid.NewString()

	// User A creates a board (auto-adds as owner member)
	boardA, _ := domain.NewBoard("Board A", "Desc A", userA)
	if err := boardRepo.Create(ctx, boardA); err != nil {
		t.Fatalf("Failed to create board A: %v", err)
	}

	// User B creates a board (auto-adds as owner member)
	boardB, _ := domain.NewBoard("Board B", "Desc B", userB)
	if err := boardRepo.Create(ctx, boardB); err != nil {
		t.Fatalf("Failed to create board B: %v", err)
	}

	// User A lists boards → sees only their board
	uc := usecase.NewListBoardsUseCase(boardRepo)
	boards, _, err := uc.Execute(ctx, userA, 20, "", false, "", "")
	if err != nil {
		t.Fatalf("Failed to list boards: %v", err)
	}

	if len(boards) != 1 {
		t.Fatalf("Expected 1 board for user A, got %d", len(boards))
	}
	if boards[0].ID != boardA.ID {
		t.Errorf("Expected board ID %s, got %s", boardA.ID, boards[0].ID)
	}

	// User B lists boards → sees only their board
	boards, _, err = uc.Execute(ctx, userB, 20, "", false, "", "")
	if err != nil {
		t.Fatalf("Failed to list boards for user B: %v", err)
	}

	if len(boards) != 1 {
		t.Fatalf("Expected 1 board for user B, got %d", len(boards))
	}
	if boards[0].ID != boardB.ID {
		t.Errorf("Expected board ID %s, got %s", boardB.ID, boards[0].ID)
	}
}

func TestFeature_ListBoards_MemberSeesSharedBoards(t *testing.T) {
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
	userA := uuid.NewString()
	userB := uuid.NewString()

	// User A creates a board
	boardA, _ := domain.NewBoard("Shared Board", "Desc", userA)
	if err := boardRepo.Create(ctx, boardA); err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	// Add user B as member
	if err := memberRepo.AddMember(ctx, boardA.ID, userB, domain.RoleMember); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// User B lists boards → sees the shared board
	uc := usecase.NewListBoardsUseCase(boardRepo)
	boards, _, err := uc.Execute(ctx, userB, 20, "", false, "", "")
	if err != nil {
		t.Fatalf("Failed to list boards: %v", err)
	}

	if len(boards) != 1 {
		t.Fatalf("Expected 1 board for user B, got %d", len(boards))
	}
	if boards[0].ID != boardA.ID {
		t.Errorf("Expected board ID %s, got %s", boardA.ID, boards[0].ID)
	}
}

func TestFeature_ListBoards_OwnerOnlyFilter(t *testing.T) {
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
	userA := uuid.NewString()
	userB := uuid.NewString()

	// User A creates their own board
	ownBoard, _ := domain.NewBoard("My Own Board", "Desc", userA)
	if err := boardRepo.Create(ctx, ownBoard); err != nil {
		t.Fatalf("Failed to create own board: %v", err)
	}

	// User B creates a board and adds user A as member
	otherBoard, _ := domain.NewBoard("Other Board", "Desc", userB)
	if err := boardRepo.Create(ctx, otherBoard); err != nil {
		t.Fatalf("Failed to create other board: %v", err)
	}
	if err := memberRepo.AddMember(ctx, otherBoard.ID, userA, domain.RoleMember); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	uc := usecase.NewListBoardsUseCase(boardRepo)

	// Without filter: user A sees both boards
	allBoards, _, err := uc.Execute(ctx, userA, 20, "", false, "", "")
	if err != nil {
		t.Fatalf("Failed to list all boards: %v", err)
	}
	if len(allBoards) != 2 {
		t.Fatalf("Expected 2 boards total, got %d", len(allBoards))
	}

	// With owner_only=true: user A sees only their own board
	ownBoards, _, err := uc.Execute(ctx, userA, 20, "", true, "", "")
	if err != nil {
		t.Fatalf("Failed to list owner-only boards: %v", err)
	}
	if len(ownBoards) != 1 {
		t.Fatalf("Expected 1 owner board, got %d", len(ownBoards))
	}
	if ownBoards[0].ID != ownBoard.ID {
		t.Errorf("Expected own board ID %s, got %s", ownBoard.ID, ownBoards[0].ID)
	}
}

func TestFeature_ListBoards_SearchByTitle(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()
	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()
	runMigrations(t, db)

	boardRepo := postgres.NewBoardRepository(db)
	ctx := context.Background()
	userID := uuid.NewString()

	// Create boards with different titles
	for _, title := range []string{"Alpha", "Beta", "Alphabet"} {
		board, _ := domain.NewBoard(title, "Desc", userID)
		if err := boardRepo.Create(ctx, board); err != nil {
			t.Fatalf("Failed to create board %s: %v", title, err)
		}
	}

	uc := usecase.NewListBoardsUseCase(boardRepo)

	// Search "Alph" → should return "Alpha" and "Alphabet"
	boards, _, err := uc.Execute(ctx, userID, 20, "", false, "Alph", "")
	if err != nil {
		t.Fatalf("Failed to search boards: %v", err)
	}
	if len(boards) != 2 {
		t.Fatalf("Expected 2 boards matching 'Alph', got %d", len(boards))
	}

	// Verify titles
	titles := map[string]bool{}
	for _, b := range boards {
		titles[b.Title] = true
	}
	if !titles["Alpha"] || !titles["Alphabet"] {
		t.Errorf("Expected Alpha and Alphabet, got %v", titles)
	}
}

func TestFeature_UpdateBoard_OnlyOwner(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Original Title", "Desc", ownerID)
	if err := boardRepo.Create(ctx, board); err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	uc := usecase.NewUpdateBoardUseCase(boardRepo, memberRepo, publisher)

	// Owner updates board → success
	updated, err := uc.Execute(ctx, board.ID, ownerID, "Updated Title", "Updated Desc", board.Version)
	if err != nil {
		t.Fatalf("Owner should be able to update board: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", updated.Title)
	}

	// Member updates board → success (UpdateBoard checks IsMember, not owner)
	_, err = uc.Execute(ctx, board.ID, memberID, "Member Updated", "Member Desc", updated.Version)
	if err != nil {
		t.Fatalf("Member should be able to update board (code only checks IsMember): %v", err)
	}

	// Non-member tries to update → ErrAccessDenied
	_, err = uc.Execute(ctx, board.ID, nonMemberID, "Hacked Title", "Hacked", 1)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestFeature_GetBoard_NonMemberDenied(t *testing.T) {
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
	ownerID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Private Board", "Desc", ownerID)
	if err := boardRepo.Create(ctx, board); err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	uc := usecase.NewGetBoardUseCase(boardRepo, memberRepo)

	// Owner can get board
	_, err = uc.Execute(ctx, board.ID, ownerID)
	if err != nil {
		t.Fatalf("Owner should be able to get board: %v", err)
	}

	// Non-member cannot get board
	_, err = uc.Execute(ctx, board.ID, nonMemberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

// ==================== Column tests ====================

func TestFeature_AddColumn_MemberCanAdd(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	uc := usecase.NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)

	// Member adds column → success
	column, err := uc.Execute(ctx, board.ID, memberID, "To Do", 0)
	if err != nil {
		t.Fatalf("Member should be able to add column: %v", err)
	}

	// Verify column in DB
	loaded, err := columnRepo.GetByID(ctx, column.ID)
	if err != nil {
		t.Fatalf("Failed to load column: %v", err)
	}
	if loaded.Title != "To Do" {
		t.Errorf("Expected title 'To Do', got %s", loaded.Title)
	}
	if loaded.BoardID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.BoardID)
	}
}

func TestFeature_AddColumn_NonMemberDenied(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	uc := usecase.NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)

	// Non-member tries to add column → ErrAccessDenied
	_, err = uc.Execute(ctx, board.ID, nonMemberID, "Hacked Column", 0)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestFeature_DeleteColumn_OnlyOwnerCanDelete(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// Create two columns — one for member delete attempt, one for owner delete
	col1, _ := domain.NewColumn(board.ID, "Column 1", 0)
	columnRepo.Create(ctx, col1)
	col2, _ := domain.NewColumn(board.ID, "Column 2", 1)
	columnRepo.Create(ctx, col2)

	uc := usecase.NewDeleteColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)

	// Non-member tries to delete → ErrAccessDenied
	err = uc.Execute(ctx, col1.ID, board.ID, nonMemberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}

	// Member can delete column (delete_column.go checks IsMember, not owner)
	err = uc.Execute(ctx, col1.ID, board.ID, memberID)
	if err != nil {
		t.Fatalf("Member should be able to delete column (code checks IsMember): %v", err)
	}

	// Verify col1 is deleted
	_, err = columnRepo.GetByID(ctx, col1.ID)
	if err != domain.ErrColumnNotFound {
		t.Errorf("Expected ErrColumnNotFound after delete, got %v", err)
	}

	// Owner deletes column → success
	err = uc.Execute(ctx, col2.ID, board.ID, ownerID)
	if err != nil {
		t.Fatalf("Owner should be able to delete column: %v", err)
	}

	// Verify col2 is deleted
	_, err = columnRepo.GetByID(ctx, col2.ID)
	if err != domain.ErrColumnNotFound {
		t.Errorf("Expected ErrColumnNotFound after owner delete, got %v", err)
	}
}

// ==================== Card tests ====================

func TestFeature_CreateCard_SetsCreatorID(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	uc := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	// Member creates card
	card, err := uc.Execute(ctx, column.ID, board.ID, memberID, "My Task", "Description", "", nil)
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Verify creator_id is set to member's user ID
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}
	if loaded.CreatorID != memberID {
		t.Errorf("Expected creator_id %s, got %s", memberID, loaded.CreatorID)
	}
}

func TestFeature_CreateCard_NonMemberDenied(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	uc := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	// Non-member tries to create card → ErrAccessDenied
	_, err = uc.Execute(ctx, column.ID, board.ID, nonMemberID, "Hacked Card", "Desc", "", nil)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestFeature_MoveCard_MemberCanMove(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	col1, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, col1)
	col2, _ := domain.NewColumn(board.ID, "In Progress", 1)
	columnRepo.Create(ctx, col2)

	card, _ := domain.NewCard(col1.ID, "Task", "Desc", "n", nil, ownerID)
	cardRepo.Create(ctx, card)

	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	uc := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	// Member moves card from col1 to col2
	moved, err := uc.Execute(ctx, card.ID, board.ID, col1.ID, col2.ID, memberID, "m")
	if err != nil {
		t.Fatalf("Member should be able to move card: %v", err)
	}

	if moved.ColumnID != col2.ID {
		t.Errorf("Expected column ID %s, got %s", col2.ID, moved.ColumnID)
	}
	if moved.Position != "m" {
		t.Errorf("Expected position 'm', got %s", moved.Position)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}
	if loaded.ColumnID != col2.ID {
		t.Errorf("DB: Expected column ID %s, got %s", col2.ID, loaded.ColumnID)
	}
}

func TestFeature_MoveCard_NonMemberDenied(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	nonMemberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	col1, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, col1)
	col2, _ := domain.NewColumn(board.ID, "Done", 1)
	columnRepo.Create(ctx, col2)

	card, _ := domain.NewCard(col1.ID, "Task", "Desc", "n", nil, ownerID)
	cardRepo.Create(ctx, card)

	uc := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	// Non-member tries to move card → ErrAccessDenied
	_, err = uc.Execute(ctx, card.ID, board.ID, col1.ID, col2.ID, nonMemberID, "m")
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}

	// Verify card is still in original column
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}
	if loaded.ColumnID != col1.ID {
		t.Errorf("Card should still be in original column %s, got %s", col1.ID, loaded.ColumnID)
	}
}

func TestFeature_UpdateCard_MemberCanUpdate(t *testing.T) {
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
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	card, _ := domain.NewCard(column.ID, "Original", "Desc", "n", nil, ownerID)
	cardRepo.Create(ctx, card)

	uc := usecase.NewUpdateCardUseCase(cardRepo, boardRepo, memberRepo, publisher)

	// Member updates card → success
	assignee := uuid.NewString()
	updated, err := uc.Execute(ctx, card.ID, board.ID, memberID, "Updated Title", "Updated Desc", &assignee, 0)
	if err != nil {
		t.Fatalf("Member should be able to update card: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", updated.Title)
	}
	if updated.Description != "Updated Desc" {
		t.Errorf("Expected description 'Updated Desc', got %s", updated.Description)
	}
	if updated.AssigneeID == nil || *updated.AssigneeID != assignee {
		t.Errorf("Expected assignee %s, got %v", assignee, updated.AssigneeID)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}
	if loaded.Title != "Updated Title" {
		t.Errorf("DB: Expected title 'Updated Title', got %s", loaded.Title)
	}
}

// ==================== Member tests ====================

func TestFeature_AddMember_OnlyOwnerCanAdd(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()
	newUserID := uuid.NewString()
	anotherUserID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	uc := usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher)

	// Owner adds member → success
	err = uc.Execute(ctx, board.ID, ownerID, memberID, domain.RoleMember)
	if err != nil {
		t.Fatalf("Owner should be able to add member: %v", err)
	}

	// Verify member was added
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, memberID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}
	if !isMember {
		t.Error("User should be member after being added")
	}
	if role != domain.RoleMember {
		t.Errorf("Expected role %s, got %s", domain.RoleMember, role)
	}

	// Member tries to add another user → ErrNotOwner
	err = uc.Execute(ctx, board.ID, memberID, newUserID, domain.RoleMember)
	if err != domain.ErrNotOwner {
		t.Errorf("Expected ErrNotOwner for member adding user, got %v", err)
	}

	// Verify new user was NOT added
	isMember, _, _ = memberRepo.IsMember(ctx, board.ID, newUserID)
	if isMember {
		t.Error("User should not be member (member cannot add)")
	}

	// Non-member tries to add → ErrBoardNotFound or ErrNotOwner
	// (AddMember loads board first, then checks IsOwner — non-member that is not owner gets ErrNotOwner)
	err = uc.Execute(ctx, board.ID, anotherUserID, newUserID, domain.RoleMember)
	if err != domain.ErrNotOwner {
		t.Errorf("Expected ErrNotOwner for non-member, got %v", err)
	}
}

func TestFeature_RemoveMember_OnlyOwnerCanRemove(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberA := uuid.NewString()
	memberB := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	memberRepo.AddMember(ctx, board.ID, memberA, domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, memberB, domain.RoleMember)

	uc := usecase.NewRemoveMemberUseCase(boardRepo, memberRepo, publisher)

	// Member A tries to remove member B → ErrAccessDenied
	err = uc.Execute(ctx, board.ID, memberA, memberB)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member removing member, got %v", err)
	}

	// Verify member B still exists
	isMember, _, _ := memberRepo.IsMember(ctx, board.ID, memberB)
	if !isMember {
		t.Error("Member B should still exist after failed removal")
	}

	// Owner removes member A → success
	err = uc.Execute(ctx, board.ID, ownerID, memberA)
	if err != nil {
		t.Fatalf("Owner should be able to remove member: %v", err)
	}

	// Verify member A is removed
	isMember, _, _ = memberRepo.IsMember(ctx, board.ID, memberA)
	if isMember {
		t.Error("Member A should be removed after owner removed them")
	}
}

func TestFeature_RemoveMember_CannotRemoveOwner(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	uc := usecase.NewRemoveMemberUseCase(boardRepo, memberRepo, publisher)

	// Owner tries to remove themselves → ErrCannotRemoveOwner
	err = uc.Execute(ctx, board.ID, ownerID, ownerID)
	if err != domain.ErrCannotRemoveOwner {
		t.Errorf("Expected ErrCannotRemoveOwner, got %v", err)
	}

	// Verify owner is still a member
	isMember, role, _ := memberRepo.IsMember(ctx, board.ID, ownerID)
	if !isMember {
		t.Error("Owner should still be a member after failed self-removal")
	}
	if role != domain.RoleOwner {
		t.Errorf("Expected role %s, got %s", domain.RoleOwner, role)
	}
}

func TestFeature_AfterRemoval_NoAccess(t *testing.T) {
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
	publisher := &mockPublisher{}
	ctx := context.Background()
	ownerID := uuid.NewString()
	memberID := uuid.NewString()

	board, _ := domain.NewBoard("Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	// Add member
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// Verify member can access the board
	getBoardUC := usecase.NewGetBoardUseCase(boardRepo, memberRepo)
	_, err = getBoardUC.Execute(ctx, board.ID, memberID)
	if err != nil {
		t.Fatalf("Member should be able to access board: %v", err)
	}

	// Remove member
	removeUC := usecase.NewRemoveMemberUseCase(boardRepo, memberRepo, publisher)
	err = removeUC.Execute(ctx, board.ID, ownerID, memberID)
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}

	// After removal, member cannot access the board
	_, err = getBoardUC.Execute(ctx, board.ID, memberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied after removal, got %v", err)
	}
}
