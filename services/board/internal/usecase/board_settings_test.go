package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBoardSettingsRepository - мок для BoardSettingsRepository
type MockBoardSettingsRepository struct {
	mock.Mock
}

func (m *MockBoardSettingsRepository) GetByBoardID(ctx context.Context, boardID string) (*domain.BoardSettings, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BoardSettings), args.Error(1)
}

func (m *MockBoardSettingsRepository) Upsert(ctx context.Context, settings *domain.BoardSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

// --- GetBoardSettings tests ---

func TestGetBoardSettings_Success(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	settingsRepo.On("GetByBoardID", mock.Anything, "board-123").
		Return(&domain.BoardSettings{BoardID: "board-123", UseBoardLabelsOnly: true}, nil)

	uc := NewGetBoardSettingsUseCase(settingsRepo, memberRepo)
	settings, err := uc.Execute(context.Background(), "board-123", "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "board-123", settings.BoardID)
	assert.True(t, settings.UseBoardLabelsOnly)

	settingsRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestGetBoardSettings_NonMember_Denied(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewGetBoardSettingsUseCase(settingsRepo, memberRepo)
	settings, err := uc.Execute(context.Background(), "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, settings)

	memberRepo.AssertExpectations(t)
}

// --- UpdateBoardSettings tests ---

func TestUpdateBoardSettings_OwnerSuccess(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleOwner, nil)
	settingsRepo.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.BoardSettings")).Return(nil)

	uc := NewUpdateBoardSettingsUseCase(settingsRepo, memberRepo)
	settings, err := uc.Execute(context.Background(), "board-123", "user-123", true, nil, 14, false)

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.True(t, settings.UseBoardLabelsOnly)

	settingsRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestUpdateBoardSettings_MemberDenied(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").
		Return(true, domain.RoleMember, nil)

	uc := NewUpdateBoardSettingsUseCase(settingsRepo, memberRepo)
	settings, err := uc.Execute(context.Background(), "board-123", "user-456", true, nil, 14, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrNotOwner, err)
	assert.Nil(t, settings)

	memberRepo.AssertExpectations(t)
}

func TestUpdateBoardSettings_NonMember_Denied(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewUpdateBoardSettingsUseCase(settingsRepo, memberRepo)
	settings, err := uc.Execute(context.Background(), "board-123", "user-999", true, nil, 14, false)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, settings)

	memberRepo.AssertExpectations(t)
}
