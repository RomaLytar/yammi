package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLabelRepository - мок для LabelRepository
type MockLabelRepository struct {
	mock.Mock
}

func (m *MockLabelRepository) Create(ctx context.Context, label *domain.Label) error {
	args := m.Called(ctx, label)
	return args.Error(0)
}

func (m *MockLabelRepository) GetByID(ctx context.Context, labelID, boardID string) (*domain.Label, error) {
	args := m.Called(ctx, labelID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Label), args.Error(1)
}

func (m *MockLabelRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Label, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Label), args.Error(1)
}

func (m *MockLabelRepository) Update(ctx context.Context, label *domain.Label) error {
	args := m.Called(ctx, label)
	return args.Error(0)
}

func (m *MockLabelRepository) Delete(ctx context.Context, labelID, boardID string) error {
	args := m.Called(ctx, labelID, boardID)
	return args.Error(0)
}

func (m *MockLabelRepository) AddToCard(ctx context.Context, cardID, boardID, labelID string) error {
	args := m.Called(ctx, cardID, boardID, labelID)
	return args.Error(0)
}

func (m *MockLabelRepository) RemoveFromCard(ctx context.Context, cardID, boardID, labelID string) error {
	args := m.Called(ctx, cardID, boardID, labelID)
	return args.Error(0)
}

func (m *MockLabelRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Label, error) {
	args := m.Called(ctx, cardID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Label), args.Error(1)
}

func (m *MockLabelRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	args := m.Called(ctx, boardID)
	return args.Int(0), args.Error(1)
}

func (m *MockLabelRepository) BatchCreate(ctx context.Context, labels []*domain.Label) error {
	args := m.Called(ctx, labels)
	return args.Error(0)
}

func (m *MockLabelRepository) CreateWithLimit(ctx context.Context, label *domain.Label, maxCount int) error {
	args := m.Called(ctx, label, maxCount)
	return args.Error(0)
}

func TestCreateLabel_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("CreateWithLimit", mock.Anything, mock.AnythingOfType("*domain.Label"), 50).Return(nil)
	publisher.On("PublishLabelCreated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "board-123", "user-123", "Bug", "#ef4444")

	assert.NoError(t, err)
	assert.NotNil(t, label)
	assert.Equal(t, "Bug", label.Name)
	assert.Equal(t, "#ef4444", label.Color)
	assert.Equal(t, "board-123", label.BoardID)
	assert.NotEmpty(t, label.ID)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestCreateLabel_NonMember_Denied(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "board-123", "user-999", "Bug", "#ef4444")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, label)

	memberRepo.AssertExpectations(t)
}

func TestCreateLabel_EmptyName_Error(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)

	uc := NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "board-123", "user-123", "", "#ef4444")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyLabelName, err)
	assert.Nil(t, label)

	memberRepo.AssertExpectations(t)
}

func TestCreateLabel_MaxLabelsReached(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("CreateWithLimit", mock.Anything, mock.AnythingOfType("*domain.Label"), 50).
		Return(domain.ErrMaxLabelsReached)

	uc := NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "board-123", "user-123", "Bug", "#ef4444")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxLabelsReached, err)
	assert.Nil(t, label)

	memberRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
}

func TestCreateLabel_InvalidColor(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("CountByBoardID", mock.Anything, "board-123").Return(5, nil)

	uc := NewCreateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "board-123", "user-123", "Bug", "invalid")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidColor, err)
	assert.Nil(t, label)

	memberRepo.AssertExpectations(t)
}

func TestDeleteLabel_OwnerCanDelete(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	labelRepo.On("Delete", mock.Anything, "label-123", "board-123").Return(nil)
	publisher.On("PublishLabelDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewDeleteLabelUseCase(labelRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "label-123", "board-123", "user-123")

	assert.NoError(t, err)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestDeleteLabel_MemberCannotDelete(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewDeleteLabelUseCase(labelRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "label-123", "board-123", "user-456")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)

	memberRepo.AssertExpectations(t)
}

func TestDeleteLabel_NonMember_Denied(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewDeleteLabelUseCase(labelRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "label-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	memberRepo.AssertExpectations(t)
}

func TestAddLabelToCard_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("GetByID", mock.Anything, "label-123", "board-123").Return(&domain.Label{ID: "label-123"}, nil)
	labelRepo.On("AddToCard", mock.Anything, "card-123", "board-123", "label-123").Return(nil)
	publisher.On("PublishCardLabelAdded", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, publisher)
	err := uc.Execute(context.Background(), "card-123", "board-123", "label-123", "user-123")

	assert.NoError(t, err)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestAddLabelToCard_NonMember_Denied(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, publisher)
	err := uc.Execute(context.Background(), "card-123", "board-123", "label-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	memberRepo.AssertExpectations(t)
}

func TestAddLabelToCard_AlreadyAssigned(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("GetByID", mock.Anything, "label-123", "board-123").Return(&domain.Label{ID: "label-123"}, nil)
	labelRepo.On("AddToCard", mock.Anything, "card-123", "board-123", "label-123").
		Return(domain.ErrLabelAlreadyOnCard)

	uc := NewAddLabelToCardUseCase(labelRepo, nil, memberRepo, publisher)
	err := uc.Execute(context.Background(), "card-123", "board-123", "label-123", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrLabelAlreadyOnCard, err)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestRemoveLabelFromCard_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("RemoveFromCard", mock.Anything, "card-123", "board-123", "label-123").Return(nil)
	publisher.On("PublishCardLabelRemoved", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewRemoveLabelFromCardUseCase(labelRepo, memberRepo, publisher)
	err := uc.Execute(context.Background(), "card-123", "board-123", "label-123", "user-123")

	assert.NoError(t, err)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardLabels_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("ListByCardID", mock.Anything, "card-123", "board-123").
		Return([]*domain.Label{
			{ID: "label-1", BoardID: "board-123", Name: "Bug", Color: "#ef4444"},
			{ID: "label-2", BoardID: "board-123", Name: "Feature", Color: "#3b82f6"},
		}, nil)

	uc := NewGetCardLabelsUseCase(labelRepo, memberRepo)
	labels, err := uc.Execute(context.Background(), "card-123", "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "Bug", labels[0].Name)
	assert.Equal(t, "Feature", labels[1].Name)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetCardLabels_NonMember_Denied(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetCardLabelsUseCase(labelRepo, memberRepo)
	labels, err := uc.Execute(context.Background(), "card-123", "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, labels)

	memberRepo.AssertExpectations(t)
}

func TestListLabels_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("ListByBoardID", mock.Anything, "board-123").
		Return([]*domain.Label{
			{ID: "label-1", BoardID: "board-123", Name: "Bug", Color: "#ef4444"},
		}, nil)

	uc := NewListLabelsUseCase(labelRepo, memberRepo)
	labels, err := uc.Execute(context.Background(), "board-123", "user-123")

	assert.NoError(t, err)
	assert.Len(t, labels, 1)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestUpdateLabel_Success(t *testing.T) {
	labelRepo := new(MockLabelRepository)
	memberRepo := new(MockMembershipRepository)
	publisher := new(MockEventPublisher)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	labelRepo.On("GetByID", mock.Anything, "label-123", "board-123").
		Return(&domain.Label{ID: "label-123", BoardID: "board-123", Name: "Bug", Color: "#ef4444"}, nil)
	labelRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Label")).Return(nil)
	publisher.On("PublishLabelUpdated", mock.Anything, mock.Anything).Return(nil).Maybe()

	uc := NewUpdateLabelUseCase(labelRepo, memberRepo, publisher)
	label, err := uc.Execute(context.Background(), "label-123", "board-123", "user-123", "Feature", "#3b82f6")

	assert.NoError(t, err)
	assert.NotNil(t, label)
	assert.Equal(t, "Feature", label.Name)
	assert.Equal(t, "#3b82f6", label.Color)

	labelRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}
