package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

// --- Mock implementations ---

type mockNotificationRepo struct {
	createFn       func(ctx context.Context, n *domain.Notification) error
	listFn         func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error)
	markReadFn     func(ctx context.Context, userID string, ids []string) error
	markAllReadFn  func(ctx context.Context, userID string) error
	unreadCountFn  func(ctx context.Context, userID string) (int, error)
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

func (m *mockNotificationRepo) ListByUserID(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, limit, cursor, typeFilter, search)
	}
	return nil, "", nil
}

func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, userID string, ids []string) error {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, userID, ids)
	}
	return nil
}

func (m *mockNotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	if m.markAllReadFn != nil {
		return m.markAllReadFn(ctx, userID)
	}
	return nil
}

func (m *mockNotificationRepo) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	if m.unreadCountFn != nil {
		return m.unreadCountFn(ctx, userID)
	}
	return 0, nil
}

type mockSettingsRepo struct {
	getFn    func(ctx context.Context, userID string) (*domain.NotificationSettings, error)
	upsertFn func(ctx context.Context, settings *domain.NotificationSettings) error
}

func (m *mockSettingsRepo) Get(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return domain.DefaultSettings(userID), nil
}

func (m *mockSettingsRepo) Upsert(ctx context.Context, settings *domain.NotificationSettings) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, settings)
	}
	return nil
}

type mockPublisher struct {
	publishFn func(ctx context.Context, n *domain.Notification) error
}

func (m *mockPublisher) PublishNotificationCreated(ctx context.Context, n *domain.Notification) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, n)
	}
	return nil
}

// --- Tests ---

func TestCreateNotification_Success(t *testing.T) {
	var created bool
	repo := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *domain.Notification) error {
			created = true
			if n.UserID != "user-1" {
				t.Errorf("expected UserID=user-1, got %s", n.UserID)
			}
			if n.Type != domain.TypeBoardCreated {
				t.Errorf("expected type board_created, got %s", n.Type)
			}
			return nil
		},
	}
	settings := &mockSettingsRepo{}
	publisher := &mockPublisher{}

	uc := NewCreateNotificationUseCase(repo, settings, publisher)
	err := uc.Execute(context.Background(), "user-1", domain.TypeBoardCreated, "Board created", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected notification to be created in repo")
	}
}

func TestCreateNotification_DisabledSettings(t *testing.T) {
	repo := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *domain.Notification) error {
			t.Error("should not create notification when disabled")
			return nil
		},
	}
	settings := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
			s := domain.DefaultSettings(userID)
			s.Enabled = false
			return s, nil
		},
	}

	uc := NewCreateNotificationUseCase(repo, settings, nil)
	err := uc.Execute(context.Background(), "user-1", domain.TypeBoardCreated, "Board created", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateNotification_RepoError(t *testing.T) {
	expectedErr := errors.New("db error")
	repo := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *domain.Notification) error {
			return expectedErr
		},
	}
	settings := &mockSettingsRepo{}

	uc := NewCreateNotificationUseCase(repo, settings, nil)
	err := uc.Execute(context.Background(), "user-1", domain.TypeBoardCreated, "Board created", "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected db error, got %v", err)
	}
}

func TestCreateNotification_ValidationError(t *testing.T) {
	repo := &mockNotificationRepo{}
	settings := &mockSettingsRepo{}

	uc := NewCreateNotificationUseCase(repo, settings, nil)
	err := uc.Execute(context.Background(), "", domain.TypeBoardCreated, "Board created", "", nil)
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestCreateNotification_SettingsError_UsesDefaults(t *testing.T) {
	var created bool
	repo := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *domain.Notification) error {
			created = true
			return nil
		},
	}
	settings := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
			return nil, errors.New("settings db error")
		},
	}

	uc := NewCreateNotificationUseCase(repo, settings, nil)
	err := uc.Execute(context.Background(), "user-1", domain.TypeWelcome, "Welcome", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected notification to be created using defaults")
	}
}

func TestCreateNotification_NilPublisher(t *testing.T) {
	repo := &mockNotificationRepo{}
	settings := &mockSettingsRepo{}

	uc := NewCreateNotificationUseCase(repo, settings, nil)
	err := uc.Execute(context.Background(), "user-1", domain.TypeWelcome, "Welcome", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
