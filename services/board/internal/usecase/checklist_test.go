package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChecklistRepository - мок для ChecklistRepository
type MockChecklistRepository struct {
	mock.Mock
}

func (m *MockChecklistRepository) CreateChecklist(ctx context.Context, checklist *domain.Checklist) error {
	args := m.Called(ctx, checklist)
	return args.Error(0)
}

func (m *MockChecklistRepository) GetChecklistByID(ctx context.Context, checklistID, boardID string) (*domain.Checklist, error) {
	args := m.Called(ctx, checklistID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Checklist), args.Error(1)
}

func (m *MockChecklistRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Checklist, error) {
	args := m.Called(ctx, cardID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Checklist), args.Error(1)
}

func (m *MockChecklistRepository) UpdateChecklist(ctx context.Context, checklist *domain.Checklist) error {
	args := m.Called(ctx, checklist)
	return args.Error(0)
}

func (m *MockChecklistRepository) DeleteChecklist(ctx context.Context, checklistID, boardID string) error {
	args := m.Called(ctx, checklistID, boardID)
	return args.Error(0)
}

func (m *MockChecklistRepository) CreateItem(ctx context.Context, item *domain.ChecklistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockChecklistRepository) GetItemByID(ctx context.Context, itemID, boardID string) (*domain.ChecklistItem, error) {
	args := m.Called(ctx, itemID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ChecklistItem), args.Error(1)
}

func (m *MockChecklistRepository) ListItemsByChecklistID(ctx context.Context, checklistID, boardID string) ([]domain.ChecklistItem, error) {
	args := m.Called(ctx, checklistID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ChecklistItem), args.Error(1)
}

func (m *MockChecklistRepository) UpdateItem(ctx context.Context, item *domain.ChecklistItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockChecklistRepository) DeleteItem(ctx context.Context, itemID, boardID string) error {
	args := m.Called(ctx, itemID, boardID)
	return args.Error(0)
}

func (m *MockChecklistRepository) ToggleItem(ctx context.Context, itemID, boardID string, isChecked bool) error {
	args := m.Called(ctx, itemID, boardID, isChecked)
	return args.Error(0)
}

func TestCreateChecklist_Success(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	checklistRepo.On("CreateChecklist", mock.Anything, mock.AnythingOfType("*domain.Checklist")).Return(nil)
	publisher.On("PublishChecklistCreated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewCreateChecklistUseCase(checklistRepo, memberRepo, publisher)
	checklist, err := uc.Execute(context.Background(), "card-123", "board-123", "user-123", "Review Tasks", 0)

	assert.NoError(t, err)
	assert.NotNil(t, checklist)
	assert.Equal(t, "Review Tasks", checklist.Title)
	assert.Equal(t, "card-123", checklist.CardID)
	assert.Equal(t, "board-123", checklist.BoardID)
	assert.NotEmpty(t, checklist.ID)

	checklistRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestCreateChecklist_NonMember_Denied(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewCreateChecklistUseCase(checklistRepo, memberRepo, publisher)
	checklist, err := uc.Execute(context.Background(), "card-123", "board-123", "user-999", "Review Tasks", 0)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, checklist)

	memberRepo.AssertExpectations(t)
}

func TestCreateChecklist_EmptyTitle(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)

	uc := NewCreateChecklistUseCase(checklistRepo, memberRepo, publisher)
	checklist, err := uc.Execute(context.Background(), "card-123", "board-123", "user-123", "", 0)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyChecklistTitle, err)
	assert.Nil(t, checklist)

	memberRepo.AssertExpectations(t)
}

func TestToggleChecklistItem_Success(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	checklistRepo.On("ToggleItem", mock.Anything, "item-123", "board-123", true).Return(nil)
	publisher.On("PublishChecklistItemToggled", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewToggleChecklistItemUseCase(checklistRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "item-123", "board-123", "user-123", true)

	assert.NoError(t, err)

	checklistRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestToggleChecklistItem_NonMember_Denied(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewToggleChecklistItemUseCase(checklistRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "item-123", "board-123", "user-999", true)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	memberRepo.AssertExpectations(t)
}

func TestDeleteChecklist_Success(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	checklistRepo.On("DeleteChecklist", mock.Anything, "checklist-123", "board-123").Return(nil)
	publisher.On("PublishChecklistDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewDeleteChecklistUseCase(checklistRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "checklist-123", "board-123", "user-123")

	assert.NoError(t, err)

	checklistRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetChecklists_Success(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	checklistRepo.On("ListByCardID", mock.Anything, "card-123", "board-123").
		Return([]*domain.Checklist{
			{ID: "cl-1", CardID: "card-123", BoardID: "board-123", Title: "Review Tasks", Items: []domain.ChecklistItem{
				{ID: "item-1", IsChecked: true},
				{ID: "item-2", IsChecked: false},
			}},
			{ID: "cl-2", CardID: "card-123", BoardID: "board-123", Title: "Deploy Tasks"},
		}, nil)

	uc := NewGetChecklistsUseCase(checklistRepo, memberRepo)
	checklists, err := uc.Execute(context.Background(), "card-123", "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, checklists, 2)
	assert.Equal(t, "Review Tasks", checklists[0].Title)
	assert.Equal(t, 50, checklists[0].Progress())
	assert.Equal(t, "Deploy Tasks", checklists[1].Title)

	checklistRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetChecklists_NonMember_Denied(t *testing.T) {
	checklistRepo := new(MockChecklistRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetChecklistsUseCase(checklistRepo, memberRepo)
	checklists, err := uc.Execute(context.Background(), "card-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, checklists)

	memberRepo.AssertExpectations(t)
}
