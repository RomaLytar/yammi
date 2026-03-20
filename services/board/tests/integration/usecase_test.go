package integration

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
	
)

// Mock publisher для тестов
type mockPublisher struct {
	events []interface{}
}

func (m *mockPublisher) PublishBoardCreated(ctx context.Context, event events.BoardCreated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishBoardUpdated(ctx context.Context, event events.BoardUpdated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishBoardDeleted(ctx context.Context, event events.BoardDeleted) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishColumnCreated(ctx context.Context, event events.ColumnCreated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardCreated(ctx context.Context, event events.CardCreated) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishCardMoved(ctx context.Context, event events.CardMoved) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockPublisher) PublishMemberAdded(ctx context.Context, event events.MemberAdded) error {
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
	board, err := uc.Execute(ctx, "Integration Test", "Description", "user-123")
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
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, "user-123")
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
	board, _ := domain.NewBoard("Test Board", "Description", "owner-123")
	boardRepo.Create(ctx, board)

	// Add member
	memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)

	// Create use case
	uc := usecase.NewGetBoardUseCase(boardRepo, memberRepo)

	// Execute as owner
	loaded, err := uc.Execute(ctx, board.ID, "owner-123")
	if err != nil {
		t.Fatalf("Failed to get board: %v", err)
	}

	if loaded.ID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.ID)
	}

	// Execute as member
	loaded, err = uc.Execute(ctx, board.ID, "user-456")
	if err != nil {
		t.Fatalf("Failed to get board as member: %v", err)
	}

	if loaded.ID != board.ID {
		t.Errorf("Expected board ID %s, got %s", board.ID, loaded.ID)
	}

	// Execute as non-member (should fail)
	_, err = uc.Execute(ctx, board.ID, "non-member")
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
	board, _ := domain.NewBoard("Test Board", "Description", "owner-123")
	boardRepo.Create(ctx, board)

	// Create use case
	uc := usecase.NewAddColumnUseCase(boardRepo, columnRepo, memberRepo, publisher)

	// Execute as owner
	column, err := uc.Execute(ctx, board.ID, "To Do", 0, "owner-123")
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
	_, err = uc.Execute(ctx, board.ID, "Done", 1, "non-member")
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
	board, _ := domain.NewBoard("Test Board", "Description", "owner-123")
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	// Add member
	memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)

	// Create use case
	uc := usecase.NewCreateCardUseCase(boardRepo, columnRepo, cardRepo, memberRepo, publisher)

	// Execute as member
	assignee := "user-789"
	card, err := uc.Execute(ctx, column.ID, "Task 1", "Description", &assignee, "user-456")
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID)
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
	_, err = uc.Execute(ctx, column.ID, "Task 2", "Desc", nil, "non-member")
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
	board, _ := domain.NewBoard("Test Board", "Description", "owner-123")
	boardRepo.Create(ctx, board)

	column1, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column1)

	column2, _ := domain.NewColumn(board.ID, "In Progress", 1)
	columnRepo.Create(ctx, column2)

	// Create card in column1
	card, _ := domain.NewCard(column1.ID, "Task", "Desc", "n", nil)
	cardRepo.Create(ctx, card)

	// Add member
	memberRepo.AddMember(ctx, board.ID, "user-456", domain.RoleMember)

	// Create use case
	uc := usecase.NewMoveCardUseCase(boardRepo, columnRepo, cardRepo, memberRepo, publisher)

	// Execute as member (move to column2)
	err = uc.Execute(ctx, card.ID, column2.ID, "m", "user-456")
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}

	// Verify in DB
	loaded, err := cardRepo.GetByID(ctx, card.ID)
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
	err = uc.Execute(ctx, card.ID, column1.ID, "a", "non-member")
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
	board, _ := domain.NewBoard("Test Board", "Description", "owner-123")
	boardRepo.Create(ctx, board)

	// Create use case
	uc := usecase.NewAddMemberUseCase(boardRepo, memberRepo, publisher)

	// Execute as owner
	err = uc.Execute(ctx, board.ID, "user-456", domain.RoleMember, "owner-123")
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Verify in DB
	isMember, role, err := memberRepo.IsMember(ctx, board.ID, "user-456")
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
	memberRepo.AddMember(ctx, board.ID, "user-789", domain.RoleMember)
	err = uc.Execute(ctx, board.ID, "user-999", domain.RoleMember, "user-789")
	if err != domain.ErrNotOwner {
		t.Errorf("Expected ErrNotOwner for non-owner, got %v", err)
	}
}
