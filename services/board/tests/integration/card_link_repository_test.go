package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/RomaLytar/yammi/services/board/internal/repository/postgres"
)

func TestCreateCardLink(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, cards
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	parentCard, _ := domain.NewCard(column.ID, "Parent Task", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, parentCard)

	childCard, _ := domain.NewCard(column.ID, "Child Task", "Desc", "b", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, childCard)

	// Create link
	link, err := domain.NewCardLink("", parentCard.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
	if err != nil {
		t.Fatalf("Failed to create domain card link: %v", err)
	}

	err = cardLinkRepo.Create(ctx, link)
	if err != nil {
		t.Fatalf("Failed to save card link: %v", err)
	}

	// Verify link exists
	loaded, err := cardLinkRepo.GetByID(ctx, link.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to load card link: %v", err)
	}

	if loaded.ParentID != parentCard.ID {
		t.Errorf("Expected parent_id %s, got %s", parentCard.ID, loaded.ParentID)
	}

	if loaded.ChildID != childCard.ID {
		t.Errorf("Expected child_id %s, got %s", childCard.ID, loaded.ChildID)
	}

	if loaded.BoardID != board.ID {
		t.Errorf("Expected board_id %s, got %s", board.ID, loaded.BoardID)
	}

	if loaded.LinkType != domain.LinkTypeSubtask {
		t.Errorf("Expected link_type subtask, got %s", loaded.LinkType)
	}
}

func TestCreateCardLink_Duplicate(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, cards
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	parentCard, _ := domain.NewCard(column.ID, "Parent Task", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, parentCard)

	childCard, _ := domain.NewCard(column.ID, "Child Task", "Desc", "b", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, childCard)

	// Create first link
	link1, _ := domain.NewCardLink("", parentCard.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
	cardLinkRepo.Create(ctx, link1)

	// Try to create duplicate
	link2, _ := domain.NewCardLink("", parentCard.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
	err := cardLinkRepo.Create(ctx, link2)

	if err != domain.ErrLinkAlreadyExists {
		t.Errorf("Expected ErrLinkAlreadyExists, got %v", err)
	}
}

func TestCreateCardLink_SelfLink(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, card
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	card, _ := domain.NewCard(column.ID, "Task", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, card)

	// Domain validation prevents self-link, but also test DB constraint
	// by manually creating a CardLink struct with same parent/child
	link := &domain.CardLink{
		ID:       uuid.NewString(),
		ParentID: card.ID,
		ChildID:  card.ID,
		BoardID:  board.ID,
		LinkType: domain.LinkTypeSubtask,
	}

	err := cardLinkRepo.Create(ctx, link)
	if err != domain.ErrSelfLink {
		t.Errorf("Expected ErrSelfLink from DB constraint, got %v", err)
	}
}

func TestDeleteCardLink(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, cards, link
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	parentCard, _ := domain.NewCard(column.ID, "Parent", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, parentCard)

	childCard, _ := domain.NewCard(column.ID, "Child", "Desc", "b", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, childCard)

	link, _ := domain.NewCardLink("", parentCard.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
	cardLinkRepo.Create(ctx, link)

	// Delete link
	err := cardLinkRepo.Delete(ctx, link.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to delete card link: %v", err)
	}

	// Verify deleted
	_, err = cardLinkRepo.GetByID(ctx, link.ID, board.ID)
	if err != domain.ErrCardLinkNotFound {
		t.Errorf("Expected ErrCardLinkNotFound after delete, got %v", err)
	}
}

func TestListChildren(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, parent + 3 children
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	parentCard, _ := domain.NewCard(column.ID, "Parent", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, parentCard)

	for i, pos := range []string{"b", "c", "d"} {
		child, _ := domain.NewCard(column.ID, "Child "+string(rune('1'+i)), "Desc", pos, nil, ownerID, nil, "", "")
		cardRepo.Create(ctx, child)

		link, _ := domain.NewCardLink("", parentCard.ID, child.ID, board.ID, domain.LinkTypeSubtask)
		cardLinkRepo.Create(ctx, link)
	}

	// List children
	links, err := cardLinkRepo.ListChildren(ctx, parentCard.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to list children: %v", err)
	}

	if len(links) != 3 {
		t.Errorf("Expected 3 children, got %d", len(links))
	}

	// All should have the same parent
	for _, l := range links {
		if l.ParentID != parentCard.ID {
			t.Errorf("Expected parent_id %s, got %s", parentCard.ID, l.ParentID)
		}
	}
}

func TestListParents(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, child + 2 parents
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	childCard, _ := domain.NewCard(column.ID, "Child", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, childCard)

	for i, pos := range []string{"b", "c"} {
		parent, _ := domain.NewCard(column.ID, "Parent "+string(rune('1'+i)), "Desc", pos, nil, ownerID, nil, "", "")
		cardRepo.Create(ctx, parent)

		link, _ := domain.NewCardLink("", parent.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
		cardLinkRepo.Create(ctx, link)
	}

	// List parents (without board_id filter)
	links, err := cardLinkRepo.ListParents(ctx, childCard.ID)
	if err != nil {
		t.Fatalf("Failed to list parents: %v", err)
	}

	if len(links) != 2 {
		t.Errorf("Expected 2 parents, got %d", len(links))
	}

	// All should have the same child
	for _, l := range links {
		if l.ChildID != childCard.ID {
			t.Errorf("Expected child_id %s, got %s", childCard.ID, l.ChildID)
		}
	}
}

func TestExists(t *testing.T) {
	t.Parallel()
	db := getSharedDB(t)

	boardRepo := postgres.NewBoardRepository(db)
	columnRepo := postgres.NewColumnRepository(db)
	cardRepo := postgres.NewCardRepository(db)
	cardLinkRepo := postgres.NewCardLinkRepository(db)
	ctx := context.Background()

	// Create board, column, cards
	ownerID := uuid.NewString()
	board, _ := domain.NewBoard("Test Board", "Desc", ownerID)
	boardRepo.Create(ctx, board)

	column, _ := domain.NewColumn(board.ID, "To Do", 0)
	columnRepo.Create(ctx, column)

	parentCard, _ := domain.NewCard(column.ID, "Parent", "Desc", "a", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, parentCard)

	childCard, _ := domain.NewCard(column.ID, "Child", "Desc", "b", nil, ownerID, nil, "", "")
	cardRepo.Create(ctx, childCard)

	// Not exists yet
	exists, err := cardLinkRepo.Exists(ctx, parentCard.ID, childCard.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}
	if exists {
		t.Error("Expected link to not exist yet")
	}

	// Create link
	link, _ := domain.NewCardLink("", parentCard.ID, childCard.ID, board.ID, domain.LinkTypeSubtask)
	cardLinkRepo.Create(ctx, link)

	// Now exists
	exists, err = cardLinkRepo.Exists(ctx, parentCard.ID, childCard.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}
	if !exists {
		t.Error("Expected link to exist after creation")
	}

	// Different direction should not exist
	exists, err = cardLinkRepo.Exists(ctx, childCard.ID, parentCard.ID, board.ID)
	if err != nil {
		t.Fatalf("Failed to check reverse exists: %v", err)
	}
	if exists {
		t.Error("Expected reverse link to not exist")
	}
}
