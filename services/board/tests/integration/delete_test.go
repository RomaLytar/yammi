package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

func TestDeleteBoard_OwnerCanDelete(t *testing.T) {
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

	board, _ := domain.NewBoard("Board to Delete", "Description", ownerID)
	if err := boardRepo.Create(ctx, board); err != nil {
		t.Fatalf("Failed to create board: %v", err)
	}

	uc := usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	if err := uc.Execute(ctx, []string{board.ID}, ownerID); err != nil {
		t.Fatalf("Owner should be able to delete board: %v", err)
	}

	_, err = boardRepo.GetByID(ctx, board.ID)
	if err != domain.ErrBoardNotFound {
		t.Errorf("Expected ErrBoardNotFound, got %v", err)
	}
}

func TestDeleteBoard_MemberCannotDelete(t *testing.T) {
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

	board, _ := domain.NewBoard("Protected Board", "Description", ownerID)
	boardRepo.Create(ctx, board)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	uc := usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	err = uc.Execute(ctx, []string{board.ID}, memberID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for member, got %v", err)
	}

	if _, err := boardRepo.GetByID(ctx, board.ID); err != nil {
		t.Errorf("Board should still exist: %v", err)
	}
}

func TestDeleteBoard_BatchDelete(t *testing.T) {
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

	var boardIDs []string
	for i := 0; i < 3; i++ {
		board, _ := domain.NewBoard("Board", "Desc", ownerID)
		boardRepo.Create(ctx, board)
		boardIDs = append(boardIDs, board.ID)
	}

	uc := usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	if err := uc.Execute(ctx, boardIDs, ownerID); err != nil {
		t.Fatalf("Failed to batch delete: %v", err)
	}

	for _, id := range boardIDs {
		if _, err := boardRepo.GetByID(ctx, id); err != domain.ErrBoardNotFound {
			t.Errorf("Board %s should be gone", id)
		}
	}
}

func TestDeleteBoard_CascadeDeletesCards(t *testing.T) {
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

	board, _ := domain.NewBoard("Board with Cards", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	card1, _ := domain.NewCard(column.ID, "Card 1", "Desc", "a", nil, ownerID)
	cardRepo.Create(ctx, card1)
	card2, _ := domain.NewCard(column.ID, "Card 2", "Desc", "b", nil, ownerID)
	cardRepo.Create(ctx, card2)

	uc := usecase.NewDeleteBoardUseCase(boardRepo, memberRepo, publisher)
	uc.Execute(ctx, []string{board.ID}, ownerID)

	if _, err := cardRepo.GetByID(ctx, card1.ID); err != domain.ErrCardNotFound {
		t.Errorf("Card 1 should be gone after cascade")
	}
	if _, err := cardRepo.GetByID(ctx, card2.ID); err != domain.ErrCardNotFound {
		t.Errorf("Card 2 should be gone after cascade")
	}
}

func TestDeleteCard_CreatorCanDelete(t *testing.T) {
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

	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	card, _ := domain.NewCard(column.ID, "My Card", "Desc", "n", nil, memberID)
	cardRepo.Create(ctx, card)

	uc := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	if err := uc.Execute(ctx, []string{card.ID}, board.ID, memberID); err != nil {
		t.Fatalf("Creator should delete own card: %v", err)
	}

	if _, err := cardRepo.GetByID(ctx, card.ID); err != domain.ErrCardNotFound {
		t.Errorf("Card should be gone")
	}
}

func TestDeleteCard_OwnerCanDeleteAnyCard(t *testing.T) {
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

	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	memberRepo.AddMember(ctx, board.ID, memberID, domain.RoleMember)

	card, _ := domain.NewCard(column.ID, "Member Card", "Desc", "n", nil, memberID)
	cardRepo.Create(ctx, card)

	uc := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	if err := uc.Execute(ctx, []string{card.ID}, board.ID, ownerID); err != nil {
		t.Fatalf("Owner should delete any card: %v", err)
	}

	if _, err := cardRepo.GetByID(ctx, card.ID); err != domain.ErrCardNotFound {
		t.Errorf("Card should be gone")
	}
}

func TestDeleteCard_MemberCannotDeleteOthersCard(t *testing.T) {
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
	memberA := uuid.NewString()
	memberB := uuid.NewString()

	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)
	memberRepo.AddMember(ctx, board.ID, memberA, domain.RoleMember)
	memberRepo.AddMember(ctx, board.ID, memberB, domain.RoleMember)

	card, _ := domain.NewCard(column.ID, "A's Card", "Desc", "n", nil, memberA)
	cardRepo.Create(ctx, card)

	uc := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	err = uc.Execute(ctx, []string{card.ID}, board.ID, memberB)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}

	if _, err := cardRepo.GetByID(ctx, card.ID); err != nil {
		t.Errorf("Card should still exist: %v", err)
	}
}

func TestDeleteCard_BatchDelete(t *testing.T) {
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

	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	var cardIDs []string
	for _, pos := range []string{"a", "b", "c"} {
		card, _ := domain.NewCard(column.ID, "Card", "Desc", pos, nil, ownerID)
		cardRepo.Create(ctx, card)
		cardIDs = append(cardIDs, card.ID)
	}

	uc := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	if err := uc.Execute(ctx, cardIDs, board.ID, ownerID); err != nil {
		t.Fatalf("Failed to batch delete: %v", err)
	}

	for _, id := range cardIDs {
		if _, err := cardRepo.GetByID(ctx, id); err != domain.ErrCardNotFound {
			t.Errorf("Card %s should be gone", id)
		}
	}
}

func TestDeleteCard_NonMemberCannotDelete(t *testing.T) {
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
	strangerID := uuid.NewString()

	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)
	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Card", "Desc", "n", nil, ownerID)
	cardRepo.Create(ctx, card)

	uc := usecase.NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
	err = uc.Execute(ctx, []string{card.ID}, board.ID, strangerID)
	if err != domain.ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied for non-member, got %v", err)
	}

	if _, err := cardRepo.GetByID(ctx, card.ID); err != nil {
		t.Errorf("Card should still exist: %v", err)
	}
}
