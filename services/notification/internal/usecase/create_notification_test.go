package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

// --- Mock implementations ---

type mockNotificationRepo struct {
	createFn      func(ctx context.Context, n *domain.Notification) error
	batchCreateFn func(ctx context.Context, notifications []*domain.Notification) error
	listFn        func(ctx context.Context, userID string, limit int, cursor string, typeFilter string, search string) ([]*domain.Notification, string, error)
	markReadFn    func(ctx context.Context, userID string, ids []string) error
	markAllReadFn func(ctx context.Context, userID string) error
	unreadCountFn func(ctx context.Context, userID string) (int, error)
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

func (m *mockNotificationRepo) BatchCreate(ctx context.Context, notifications []*domain.Notification) error {
	if m.batchCreateFn != nil {
		return m.batchCreateFn(ctx, notifications)
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
	getFn      func(ctx context.Context, userID string) (*domain.NotificationSettings, error)
	batchGetFn func(ctx context.Context, userIDs []string) (map[string]*domain.NotificationSettings, error)
	upsertFn   func(ctx context.Context, settings *domain.NotificationSettings) error
}

func (m *mockSettingsRepo) Get(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return domain.DefaultSettings(userID), nil
}

func (m *mockSettingsRepo) BatchGet(ctx context.Context, userIDs []string) (map[string]*domain.NotificationSettings, error) {
	if m.batchGetFn != nil {
		return m.batchGetFn(ctx, userIDs)
	}
	result := make(map[string]*domain.NotificationSettings, len(userIDs))
	for _, uid := range userIDs {
		result[uid] = domain.DefaultSettings(uid)
	}
	return result, nil
}

func (m *mockSettingsRepo) Upsert(ctx context.Context, settings *domain.NotificationSettings) error {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, settings)
	}
	return nil
}

type mockPublisher struct {
	publishFn      func(ctx context.Context, n *domain.Notification) error
	publishBatchFn func(ctx context.Context, notifications []*domain.Notification) error
}

func (m *mockPublisher) PublishNotificationCreated(ctx context.Context, n *domain.Notification) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, n)
	}
	return nil
}

func (m *mockPublisher) PublishNotificationsBatch(ctx context.Context, notifications []*domain.Notification) error {
	if m.publishBatchFn != nil {
		return m.publishBatchFn(ctx, notifications)
	}
	return nil
}

type mockBoardEventRepo struct {
	createFn            func(ctx context.Context, event *domain.BoardEvent) error
	listForUserFn       func(ctx context.Context, userID string, boardIDs []string, limit int, cursor string) ([]*domain.Notification, string, error)
	markBoardReadFn     func(ctx context.Context, userID, boardID string) error
	markAllBoardsReadFn func(ctx context.Context, userID string, boardIDs []string) error
	getBoardIDByEventFn func(ctx context.Context, eventID string) (string, error)
}

func (m *mockBoardEventRepo) Create(ctx context.Context, event *domain.BoardEvent) error {
	if m.createFn != nil {
		return m.createFn(ctx, event)
	}
	return nil
}

func (m *mockBoardEventRepo) ListForUser(ctx context.Context, userID string, boardIDs []string, limit int, cursor string) ([]*domain.Notification, string, error) {
	if m.listForUserFn != nil {
		return m.listForUserFn(ctx, userID, boardIDs, limit, cursor)
	}
	return nil, "", nil
}

func (m *mockBoardEventRepo) MarkBoardRead(ctx context.Context, userID, boardID string) error {
	if m.markBoardReadFn != nil {
		return m.markBoardReadFn(ctx, userID, boardID)
	}
	return nil
}

func (m *mockBoardEventRepo) MarkAllBoardsRead(ctx context.Context, userID string, boardIDs []string) error {
	if m.markAllBoardsReadFn != nil {
		return m.markAllBoardsReadFn(ctx, userID, boardIDs)
	}
	return nil
}

func (m *mockBoardEventRepo) GetBoardIDByEventID(ctx context.Context, eventID string) (string, error) {
	if m.getBoardIDByEventFn != nil {
		return m.getBoardIDByEventFn(ctx, eventID)
	}
	return "", nil
}

type mockUnreadCounter struct {
	incrementFn     func(ctx context.Context, userID string) error
	incrementManyFn func(ctx context.Context, userIDs []string) error
	getFn           func(ctx context.Context, userID string) (int, error)
	resetFn         func(ctx context.Context, userID string) error
	decrementFn     func(ctx context.Context, userID string) error
}

func (m *mockUnreadCounter) Increment(ctx context.Context, userID string) error {
	if m.incrementFn != nil {
		return m.incrementFn(ctx, userID)
	}
	return nil
}

func (m *mockUnreadCounter) IncrementMany(ctx context.Context, userIDs []string) error {
	if m.incrementManyFn != nil {
		return m.incrementManyFn(ctx, userIDs)
	}
	return nil
}

func (m *mockUnreadCounter) Get(ctx context.Context, userID string) (int, error) {
	if m.getFn != nil {
		return m.getFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockUnreadCounter) Reset(ctx context.Context, userID string) error {
	if m.resetFn != nil {
		return m.resetFn(ctx, userID)
	}
	return nil
}

func (m *mockUnreadCounter) Decrement(ctx context.Context, userID string) error {
	if m.decrementFn != nil {
		return m.decrementFn(ctx, userID)
	}
	return nil
}

type mockBoardMemberRepo struct {
	addMemberFn        func(ctx context.Context, boardID, userID string) error
	removeMemberFn     func(ctx context.Context, boardID, userID string) error
	removeAllByBoardFn func(ctx context.Context, boardID string) error
	listMemberIDsFn    func(ctx context.Context, boardID string) ([]string, error)
	listBoardIDsFn     func(ctx context.Context, userID string) ([]string, error)
	truncateCacheFn    func(ctx context.Context) error
}

func (m *mockBoardMemberRepo) AddMember(ctx context.Context, boardID, userID string) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, boardID, userID)
	}
	return nil
}

func (m *mockBoardMemberRepo) RemoveMember(ctx context.Context, boardID, userID string) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, boardID, userID)
	}
	return nil
}

func (m *mockBoardMemberRepo) RemoveAllByBoard(ctx context.Context, boardID string) error {
	if m.removeAllByBoardFn != nil {
		return m.removeAllByBoardFn(ctx, boardID)
	}
	return nil
}

func (m *mockBoardMemberRepo) ListMemberIDs(ctx context.Context, boardID string) ([]string, error) {
	if m.listMemberIDsFn != nil {
		return m.listMemberIDsFn(ctx, boardID)
	}
	return nil, nil
}

func (m *mockBoardMemberRepo) ListBoardIDsByUser(ctx context.Context, userID string) ([]string, error) {
	if m.listBoardIDsFn != nil {
		return m.listBoardIDsFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockBoardMemberRepo) TruncateCache(ctx context.Context) error {
	if m.truncateCacheFn != nil {
		return m.truncateCacheFn(ctx)
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

	uc := NewCreateNotificationUseCase(repo, settings, publisher, nil, nil, nil)
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

	uc := NewCreateNotificationUseCase(repo, settings, nil, nil, nil, nil)
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

	uc := NewCreateNotificationUseCase(repo, settings, nil, nil, nil, nil)
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

	uc := NewCreateNotificationUseCase(repo, settings, nil, nil, nil, nil)
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

	uc := NewCreateNotificationUseCase(repo, settings, nil, nil, nil, nil)
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

	uc := NewCreateNotificationUseCase(repo, settings, nil, nil, nil, nil)
	err := uc.Execute(context.Background(), "user-1", domain.TypeWelcome, "Welcome", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateBoardEvent_Success(t *testing.T) {
	var eventCreated bool
	var incrementedUsers []string

	boardEventRepo := &mockBoardEventRepo{
		createFn: func(ctx context.Context, event *domain.BoardEvent) error {
			eventCreated = true
			if event.BoardID != "board-1" {
				t.Errorf("expected BoardID=board-1, got %s", event.BoardID)
			}
			if event.ActorID != "actor-1" {
				t.Errorf("expected ActorID=actor-1, got %s", event.ActorID)
			}
			return nil
		},
	}
	memberRepo := &mockBoardMemberRepo{
		listMemberIDsFn: func(ctx context.Context, boardID string) ([]string, error) {
			return []string{"actor-1", "user-2", "user-3"}, nil
		},
	}
	unreadCounter := &mockUnreadCounter{
		incrementManyFn: func(ctx context.Context, userIDs []string) error {
			incrementedUsers = userIDs
			return nil
		},
	}
	settings := &mockSettingsRepo{}

	uc := NewCreateNotificationUseCase(&mockNotificationRepo{}, settings, nil, boardEventRepo, unreadCounter, memberRepo)
	err := uc.CreateBoardEvent(context.Background(), "board-1", "actor-1", domain.TypeBoardUpdated, "Board updated", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eventCreated {
		t.Error("expected board event to be created")
	}
	if len(incrementedUsers) != 2 {
		t.Errorf("expected 2 users to be incremented, got %d", len(incrementedUsers))
	}
}

func TestCreateBoardEvent_EventRepoError(t *testing.T) {
	expectedErr := errors.New("db error")
	boardEventRepo := &mockBoardEventRepo{
		createFn: func(ctx context.Context, event *domain.BoardEvent) error {
			return expectedErr
		},
	}

	uc := NewCreateNotificationUseCase(&mockNotificationRepo{}, &mockSettingsRepo{}, nil, boardEventRepo, &mockUnreadCounter{}, &mockBoardMemberRepo{})
	err := uc.CreateBoardEvent(context.Background(), "board-1", "actor-1", domain.TypeBoardUpdated, "Board updated", "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected db error, got %v", err)
	}
}
