package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

func TestMarkRead_Success(t *testing.T) {
	var calledIDs []string
	repo := &mockNotificationRepo{
		markReadFn: func(ctx context.Context, userID string, ids []string) error {
			calledIDs = ids
			return nil
		},
	}

	uc := NewMarkReadUseCase(repo)
	err := uc.Execute(context.Background(), "user-1", []string{"n-1", "n-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calledIDs) != 2 {
		t.Errorf("expected 2 IDs, got %d", len(calledIDs))
	}
}

func TestMarkRead_EmptyUserID(t *testing.T) {
	uc := NewMarkReadUseCase(&mockNotificationRepo{})
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

	uc := NewMarkReadUseCase(repo)
	err := uc.Execute(context.Background(), "user-1", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarkAllRead_Success(t *testing.T) {
	var called bool
	repo := &mockNotificationRepo{
		markAllReadFn: func(ctx context.Context, userID string) error {
			called = true
			return nil
		},
	}

	uc := NewMarkAllReadUseCase(repo)
	err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected MarkAllAsRead to be called")
	}
}

func TestMarkAllRead_EmptyUserID(t *testing.T) {
	uc := NewMarkAllReadUseCase(&mockNotificationRepo{})
	err := uc.Execute(context.Background(), "")
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}
