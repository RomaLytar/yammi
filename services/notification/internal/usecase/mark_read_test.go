package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

func TestMarkRead_Success(t *testing.T) {
	var calledIDs []string
	repo := &mockNotificationRepo{
		markReadFn: func(ctx context.Context, userID string, ids []string) error {
			calledIDs = ids
			return nil
		},
	}
	boardEventRepo := &mockBoardEventRepo{
		getBoardIDByEventFn: func(ctx context.Context, eventID string) (string, error) {
			return "", nil // not a board event
		},
	}
	unreadCounter := &mockUnreadCounter{}

	uc := NewMarkReadUseCase(repo, boardEventRepo, unreadCounter)
	err := uc.Execute(context.Background(), "user-1", []string{"n-1", "n-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calledIDs) != 2 {
		t.Errorf("expected 2 calls, got %d", len(calledIDs))
	}
}

func TestMarkRead_BoardEvent(t *testing.T) {
	var boardMarked bool
	boardEventRepo := &mockBoardEventRepo{
		getBoardIDByEventFn: func(ctx context.Context, eventID string) (string, error) {
			return "board-1", nil // is a board event
		},
		markBoardReadFn: func(ctx context.Context, userID, boardID string) error {
			boardMarked = true
			if boardID != "board-1" {
				t.Errorf("expected board-1, got %s", boardID)
			}
			return nil
		},
	}
	repo := &mockNotificationRepo{
		markReadFn: func(ctx context.Context, userID string, ids []string) error {
			t.Error("should not call direct mark read for board events")
			return nil
		},
	}
	unreadCounter := &mockUnreadCounter{}

	uc := NewMarkReadUseCase(repo, boardEventRepo, unreadCounter)
	err := uc.Execute(context.Background(), "user-1", []string{"event-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !boardMarked {
		t.Error("expected board to be marked as read")
	}
}

func TestMarkRead_EmptyUserID(t *testing.T) {
	uc := NewMarkReadUseCase(&mockNotificationRepo{}, &mockBoardEventRepo{}, &mockUnreadCounter{})
	err := uc.Execute(context.Background(), "", []string{"n-1"})
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestMarkRead_EmptyIDs(t *testing.T) {
	repo := &mockNotificationRepo{
		markReadFn: func(ctx context.Context, userID string, ids []string) error {
			t.Error("should not call repo for empty IDs")
			return nil
		},
	}

	uc := NewMarkReadUseCase(repo, &mockBoardEventRepo{}, &mockUnreadCounter{})
	err := uc.Execute(context.Background(), "user-1", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkAllRead_Success(t *testing.T) {
	var directCalled bool
	var boardsCalled bool
	var resetCalled bool
	repo := &mockNotificationRepo{
		markAllReadFn: func(ctx context.Context, userID string) error {
			directCalled = true
			return nil
		},
	}
	boardEventRepo := &mockBoardEventRepo{
		markAllBoardsReadFn: func(ctx context.Context, userID string, boardIDs []string) error {
			boardsCalled = true
			return nil
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return []string{"board-1", "board-2"}, nil
		},
	}
	unreadCounter := &mockUnreadCounter{
		resetFn: func(ctx context.Context, userID string) error {
			resetCalled = true
			return nil
		},
	}

	uc := NewMarkAllReadUseCase(repo, boardEventRepo, memberRepo, unreadCounter)
	err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !directCalled {
		t.Error("expected MarkAllAsRead to be called")
	}
	if !boardsCalled {
		t.Error("expected MarkAllBoardsRead to be called")
	}
	if !resetCalled {
		t.Error("expected unread counter Reset to be called")
	}
}

func TestMarkAllRead_EmptyUserID(t *testing.T) {
	uc := NewMarkAllReadUseCase(&mockNotificationRepo{}, &mockBoardEventRepo{}, &mockBoardMemberRepo{}, &mockUnreadCounter{})
	err := uc.Execute(context.Background(), "")
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}
