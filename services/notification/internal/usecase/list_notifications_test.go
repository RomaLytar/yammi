package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

func TestListNotifications_Success(t *testing.T) {
	now := time.Now()
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
			return []*domain.Notification{
				{ID: "n-1", UserID: userID, Type: domain.TypeWelcome, Title: "Welcome", CreatedAt: now},
			}, "", nil
		},
	}
	boardEventRepo := &mockBoardEventRepo{}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return nil, nil
		},
	}
	unreadCounter := &mockUnreadCounter{
		getFn: func(ctx context.Context, userID string) (int, error) {
			return 3, nil
		},
	}

	uc := NewListNotificationsUseCase(repo, boardEventRepo, memberRepo, unreadCounter)
	notifications, nextCursor, unread, err := uc.Execute(context.Background(), "user-1", 20, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 1 {
		t.Errorf("expected 1 notification, got %d", len(notifications))
	}
	if nextCursor != "" {
		t.Errorf("expected empty cursor, got %s", nextCursor)
	}
	if unread != 3 {
		t.Errorf("expected 3 unread, got %d", unread)
	}
}

func TestListNotifications_EmptyUserID(t *testing.T) {
	uc := NewListNotificationsUseCase(&mockNotificationRepo{}, &mockBoardEventRepo{}, &mockBoardMemberRepo{}, &mockUnreadCounter{})
	_, _, _, err := uc.Execute(context.Background(), "", 20, "", "", "")
	if err == nil {
		t.Fatal("expected error for empty user_id")
	}
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestListNotifications_DefaultLimit(t *testing.T) {
	var capturedLimit int
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
			capturedLimit = limit
			return nil, "", nil
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return nil, nil
		},
	}
	unreadCounter := &mockUnreadCounter{}

	uc := NewListNotificationsUseCase(repo, &mockBoardEventRepo{}, memberRepo, unreadCounter)
	_, _, _, err := uc.Execute(context.Background(), "user-1", 0, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
}

func TestListNotifications_MaxLimit(t *testing.T) {
	var capturedLimit int
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
			capturedLimit = limit
			return nil, "", nil
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return nil, nil
		},
	}
	unreadCounter := &mockUnreadCounter{}

	uc := NewListNotificationsUseCase(repo, &mockBoardEventRepo{}, memberRepo, unreadCounter)
	_, _, _, err := uc.Execute(context.Background(), "user-1", 200, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedLimit)
	}
}

func TestListNotifications_RepoError(t *testing.T) {
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
			return nil, "", errors.New("db error")
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return nil, nil
		},
	}

	uc := NewListNotificationsUseCase(repo, &mockBoardEventRepo{}, memberRepo, &mockUnreadCounter{})
	_, _, _, err := uc.Execute(context.Background(), "user-1", 20, "", "", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListNotifications_MergesBoardEventsAndDirect(t *testing.T) {
	now := time.Now()
	repo := &mockNotificationRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
			return []*domain.Notification{
				{ID: "direct-1", UserID: userID, Type: domain.TypeWelcome, Title: "Welcome", CreatedAt: now.Add(-2 * time.Minute)},
			}, "", nil
		},
	}
	boardEventRepo := &mockBoardEventRepo{
		listForUserFn: func(ctx context.Context, userID string, boardIDs []string, limit int, cursor string) ([]*domain.Notification, string, error) {
			return []*domain.Notification{
				{ID: "event-1", UserID: userID, Type: domain.TypeCardCreated, Title: "Card created", CreatedAt: now.Add(-1 * time.Minute)},
			}, "", nil
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listBoardIDsFn: func(ctx context.Context, userID string) ([]string, error) {
			return []string{"board-1"}, nil
		},
	}
	unreadCounter := &mockUnreadCounter{
		getFn: func(ctx context.Context, userID string) (int, error) {
			return 2, nil
		},
	}

	uc := NewListNotificationsUseCase(repo, boardEventRepo, memberRepo, unreadCounter)
	notifications, _, unread, err := uc.Execute(context.Background(), "user-1", 20, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(notifications))
	}
	if unread != 2 {
		t.Errorf("expected 2 unread, got %d", unread)
	}
	// Verify sorted by created_at DESC (event-1 is newer)
	if len(notifications) == 2 && notifications[0].ID != "event-1" {
		t.Errorf("expected event-1 first (newer), got %s", notifications[0].ID)
	}
}
