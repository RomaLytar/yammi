package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserLabelRepository - мок для UserLabelRepository
type MockUserLabelRepository struct {
	mock.Mock
}

func (m *MockUserLabelRepository) Create(ctx context.Context, label *domain.UserLabel) error {
	args := m.Called(ctx, label)
	return args.Error(0)
}

func (m *MockUserLabelRepository) GetByID(ctx context.Context, labelID string) (*domain.UserLabel, error) {
	args := m.Called(ctx, labelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserLabel), args.Error(1)
}

func (m *MockUserLabelRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.UserLabel, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.UserLabel), args.Error(1)
}

func (m *MockUserLabelRepository) Update(ctx context.Context, label *domain.UserLabel) error {
	args := m.Called(ctx, label)
	return args.Error(0)
}

func (m *MockUserLabelRepository) Delete(ctx context.Context, labelID string) error {
	args := m.Called(ctx, labelID)
	return args.Error(0)
}

func (m *MockUserLabelRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// --- CreateUserLabel tests ---

func TestCreateUserLabel_Success(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("CountByUserID", mock.Anything, "user-123").Return(5, nil)
	userLabelRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.UserLabel")).Return(nil)

	uc := NewCreateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "user-123", "Bug", "#ef4444")

	assert.NoError(t, err)
	assert.NotNil(t, label)
	assert.Equal(t, "Bug", label.Name)
	assert.Equal(t, "#ef4444", label.Color)
	assert.Equal(t, "user-123", label.UserID)
	assert.NotEmpty(t, label.ID)

	userLabelRepo.AssertExpectations(t)
}

func TestCreateUserLabel_EmptyName(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("CountByUserID", mock.Anything, "user-123").Return(5, nil)

	uc := NewCreateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "user-123", "", "#ef4444")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrEmptyLabelName, err)
	assert.Nil(t, label)
}

func TestCreateUserLabel_InvalidColor(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("CountByUserID", mock.Anything, "user-123").Return(5, nil)

	uc := NewCreateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "user-123", "Bug", "invalid")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidColor, err)
	assert.Nil(t, label)
}

func TestCreateUserLabel_MaxLabelsReached(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("CountByUserID", mock.Anything, "user-123").Return(50, nil)

	uc := NewCreateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "user-123", "Bug", "#ef4444")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrMaxUserLabelsReached, err)
	assert.Nil(t, label)

	userLabelRepo.AssertExpectations(t)
}

// --- ListUserLabels tests ---

func TestListUserLabels_Success(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("ListByUserID", mock.Anything, "user-123").
		Return([]*domain.UserLabel{
			{ID: "label-1", UserID: "user-123", Name: "Bug", Color: "#ef4444"},
			{ID: "label-2", UserID: "user-123", Name: "Feature", Color: "#3b82f6"},
		}, nil)

	uc := NewListUserLabelsUseCase(userLabelRepo)
	labels, err := uc.Execute(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.Len(t, labels, 2)
	assert.Equal(t, "Bug", labels[0].Name)
	assert.Equal(t, "Feature", labels[1].Name)

	userLabelRepo.AssertExpectations(t)
}

// --- UpdateUserLabel tests ---

func TestUpdateUserLabel_Success(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-123").
		Return(&domain.UserLabel{ID: "label-123", UserID: "user-123", Name: "Bug", Color: "#ef4444"}, nil)
	userLabelRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.UserLabel")).Return(nil)

	uc := NewUpdateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "label-123", "user-123", "Feature", "#3b82f6")

	assert.NoError(t, err)
	assert.NotNil(t, label)
	assert.Equal(t, "Feature", label.Name)
	assert.Equal(t, "#3b82f6", label.Color)

	userLabelRepo.AssertExpectations(t)
}

func TestUpdateUserLabel_NotOwner(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-123").
		Return(&domain.UserLabel{ID: "label-123", UserID: "user-123", Name: "Bug", Color: "#ef4444"}, nil)

	uc := NewUpdateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "label-123", "user-999", "Feature", "#3b82f6")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, label)

	userLabelRepo.AssertExpectations(t)
}

func TestUpdateUserLabel_NotFound(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-999").
		Return(nil, domain.ErrUserLabelNotFound)

	uc := NewUpdateUserLabelUseCase(userLabelRepo)
	label, err := uc.Execute(context.Background(), "label-999", "user-123", "Feature", "#3b82f6")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserLabelNotFound, err)
	assert.Nil(t, label)

	userLabelRepo.AssertExpectations(t)
}

// --- DeleteUserLabel tests ---

func TestDeleteUserLabel_Success(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-123").
		Return(&domain.UserLabel{ID: "label-123", UserID: "user-123", Name: "Bug", Color: "#ef4444"}, nil)
	userLabelRepo.On("Delete", mock.Anything, "label-123").Return(nil)

	uc := NewDeleteUserLabelUseCase(userLabelRepo)
	err := uc.Execute(context.Background(), "label-123", "user-123")

	assert.NoError(t, err)

	userLabelRepo.AssertExpectations(t)
}

func TestDeleteUserLabel_NotOwner(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-123").
		Return(&domain.UserLabel{ID: "label-123", UserID: "user-123", Name: "Bug", Color: "#ef4444"}, nil)

	uc := NewDeleteUserLabelUseCase(userLabelRepo)
	err := uc.Execute(context.Background(), "label-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)

	userLabelRepo.AssertExpectations(t)
}

func TestDeleteUserLabel_NotFound(t *testing.T) {
	userLabelRepo := new(MockUserLabelRepository)

	userLabelRepo.On("GetByID", mock.Anything, "label-999").
		Return(nil, domain.ErrUserLabelNotFound)

	uc := NewDeleteUserLabelUseCase(userLabelRepo)
	err := uc.Execute(context.Background(), "label-999", "user-123")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrUserLabelNotFound, err)

	userLabelRepo.AssertExpectations(t)
}
