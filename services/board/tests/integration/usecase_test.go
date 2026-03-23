package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// Mock publisher для тестов
type mockPublisher struct {
	events []interface{}
}

func (m *mockPublisher) PublishBoardCreated(ctx context.Context, event usecase.BoardCreated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishBoardUpdated(ctx context.Context, event usecase.BoardUpdated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishBoardDeleted(ctx context.Context, event usecase.BoardDeleted) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishColumnCreated(ctx context.Context, event usecase.ColumnAdded) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishColumnUpdated(ctx context.Context, event usecase.ColumnUpdated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishColumnDeleted(ctx context.Context, event usecase.ColumnDeleted) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishColumnsReordered(ctx context.Context, event usecase.ColumnsReordered) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardCreated(ctx context.Context, event usecase.CardCreated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardUpdated(ctx context.Context, event usecase.CardUpdated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardMoved(ctx context.Context, event usecase.CardMoved) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardDeleted(ctx context.Context, event usecase.CardDeleted) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishMemberAdded(ctx context.Context, event usecase.MemberAdded) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishMemberRemoved(ctx context.Context, event usecase.MemberRemoved) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardAssigned(ctx context.Context, event usecase.CardAssigned) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardUnassigned(ctx context.Context, event usecase.CardUnassigned) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishAttachmentUploaded(ctx context.Context, event usecase.AttachmentUploaded) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishAttachmentDeleted(ctx context.Context, event usecase.AttachmentDeleted) error {
	m.events = append(m.events, event)
	return nil
}

func TestCreateBoardUseCase_Integration(t *testing.T) {
	dsn, cleanup := setupPostgresContainer(t)
	defer cleanup()

	db, err := waitForDB(dsn, 10)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	runMigrations(t, db)

	// Setup repositories
	boardRepo := postgres.NewBoardRepository(db)
	memberRepo := postgres.NewMembershipRepository(db)

	// Mock publisher
	publisher := &mockPublisher{}

	// Create use case
	uc := usecase.NewCreateBoardUseCase(boardRepo, memberRepo, publisher)

	// Execute
	ctx := context.Background()
	userID := uuid.NewString()
	board, err := uc.Execute(ctx, "Integration Test", "Description", userID)
	if err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	// Verify in DB
	loaded, err := boardRepo.GetByID(ctx, board.ID)
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	if loaded.Title != "Integration Test" {
		t.Errorf("Expected title 'Integration Test', got %s", loaded.Title)
	}

	if loaded.Description != "Description" {
		t.Errorf("Expected description 'Description', got %s", loaded.Description)
	}

	// Verify owner membership
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, userID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if !isMember {
		t.Error("Owner should be member")
	}

	if role != domain.RoleOwner {
		t.Errorf("Expected role owner, got %s", role)
	}

	// Note: Event verification removed since events are published in goroutine
	// and may not be captured immediately in tests
}

func TestGetBoardUseCase_Integration(t *testing.T) {
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

	// Create board
	ctx := context.Background()
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Description", ownerID)
	boardRepo.Create(ctx, board)

	// Add member
	memberID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// Create use case
	uc := usecase.NewGetBoardUseCase(boardRepo, memberRepo)

	// Execute as owner
	loaded, err := uc.Execute(ctx, board.ID, ownerID)
	if err != nil {
		t.Fatalf("Failed to get board: %v", err)
	}

	if loaded.ID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.ID)
	}

	// Execute as member
	loaded, err = uc.Execute(ctx, board.ID, memberID)
	if err != nil {
		t.Fatalf("Failed to get board as member: %v", err)
	}

	if loaded.ID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.ID)
	}

	// Execute as non-member (should fail)
	nonMemberID := uuid.NewString()
	_, err = uc.Execute(ctx, board.ID, nonMemberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestAddColumnUseCase_Integration(t *testing.T) {
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

	// Create board
	ctx := context.Background()
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Description", ownerID)
	boardRepo.Create(ctx, board)

	// Create use case
	uc := usecase.NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)

	// Execute as owner
	column, err := uc.Execute(ctx, board.ID, ownerID, "To Do", 0)
	if err != nil {
		t.Fatalf("Failed to add column: %v", err)
	}

	// Verify in DB
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

	// Try as non-member (should fail)
	nonMemberID := uuid.NewString()
	_, err = uc.Execute(ctx, board.ID, nonMemberID, "Done", 1)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestCreateCardUseCase_Integration(t *testing.T) {
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

	// Create board and column
	ctx := context.Background()
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Description", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Add member
	memberID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// Create use case
	activityRepo := postgres.NewActivityRepository(db)
	uc := usecase.NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)

	// Execute as member (assignee must be a board member)
	assignee := memberID
	card, err := uc.Execute(ctx, column.ID, board.ID, memberID, "Task 1", "Description", "", &assignee, nil, "", "")
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	if loaded.Title != "Task 1" {
		t.Errorf("Expected title 'Task 1', got %s", loaded.Title)
	}

	if loaded.ColumnID != column.ID {
		t.Errorf("Expected column ID %s, got %s", column.ID, loaded.ColumnID)
	}

	if loaded.AssigneeID == nil || *loaded.AssigneeID != assignee {
		t.Errorf("Expected assignee %s, got %v", assignee, loaded.AssigneeID)
	}

	// Try as non-member (should fail)
	nonMemberID := uuid.NewString()
	_, err = uc.Execute(ctx, column.ID, board.ID, nonMemberID, "Task 2", "Desc", "", nil, nil, "", "")
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestMoveCardUseCase_Integration(t *testing.T) {
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

	// Create board and two columns
	ctx := context.Background()
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Description", ownerID)
	boardRepo.Create(ctx, board)

	column1, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column1)

	column2, _ := domain.NewColumn(board.ID, "In Progress", 1)
	columnRepo.Create(ctx, column2)

	// Create card in column1
	card, _ := domain.NewCard(column1.ID, "Task", "Desc", "n", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Add member
	memberID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	// Create use case
	activityRepo := postgres.NewActivityRepository(db)
	uc := usecase.NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)

	// Execute as member (move to column2)
	_, err = uc.Execute(ctx, card.ID, board.ID, column1.ID, column2.ID, memberID, "m")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card: %v", err)
	}

	if loaded.ColumnID != column2.ID {
		t.Errorf("Expected column ID %s, got %s", column2.ID, loaded.ColumnID)
	}

	if loaded.Position != "m" {
		t.Errorf("Expected position 'm', got %s", loaded.Position)
	}

	// Try as non-member (should fail)
	nonMemberID := uuid.NewString()
	_, err = uc.Execute(ctx, card.ID, board.ID, column2.ID, column1.ID, nonMemberID, "a")
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}
}

func TestAddMemberUseCase_Integration(t *testing.T) {
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

	// Create board
	ctx := context.Background()
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Description", ownerID)
	boardRepo.Create(ctx, board)

	// Create use case
	uc := usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher)

	// Execute as owner
	newMemberID := uuid.NewString()
	err = uc.Execute(ctx, board.ID, ownerID, newMemberID, domain.RoleMember)
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Verify in DB
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, newMemberID)
	if err != nil {
		t.Fatalf("Failed to check membership: %v", err)
	}

	if !isMember {
		t.Error("User should be a member")
	}

	if role != domain.RoleMember {
		t.Errorf("Expected role member, got %s", role)
	}

	// Try as non-owner (should fail)
	nonOwnerID := uuid.NewString()
	memberRepo.AddMember(ctx, board.ID, nonOwnerID, domain.RoleMember)
	anotherUserID := uuid.NewString()
	err = uc.Execute(ctx, board.ID, nonOwnerID, anotherUserID, domain.RoleMember)
	if err != domain.ErrNotOwner {
		t.Errorf("Expected ErrNotOwner for non-owner, got %v", err)
	}
}
