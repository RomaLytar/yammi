package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAvailableLabels_AllLabels(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	labelRepo := new(MockLabelRepository)
	userLabelRepo := new(MockUserLabelRepository)
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	settingsRepo.On("GetByBoardID", mock.Anything, "board-123").
		Return(&domain.BoardSettings{BoardID: "board-123", UseBoardLabelsOnly: false}, nil)
	labelRepo.On("ListByBoardID", mock.Anything, "board-123").
		Return([]*domain.Label{
			{ID: "bl-1", BoardID: "board-123", Name: "Bug", Color: "#ef4444"},
		}, nil)
	boardRepo.On("GetByID", mock.Anything, "board-123").
		Return(&domain.Board{ID: "board-123", OwnerID: "owner-123", Title: "Test", Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)
	userLabelRepo.On("ListByUserID", mock.Anything, "owner-123").
		Return([]*domain.UserLabel{
			{ID: "ul-1", UserID: "owner-123", Name: "Global Bug", Color: "#3b82f6"},
		}, nil)

	uc := NewListAvailableLabelsUseCase(settingsRepo, labelRepo, userLabelRepo, boardRepo, memberRepo)
	result, err := uc.Execute(context.Background(), "board-123", "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.BoardLabels, 1)
	assert.Len(t, result.UserLabels, 1)
	assert.False(t, result.UseBoardLabelsOnly)
	assert.Equal(t, "Bug", result.BoardLabels[0].Name)
	assert.Equal(t, "Global Bug", result.UserLabels[0].Name)

	settingsRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
	userLabelRepo.AssertExpectations(t)
	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestListAvailableLabels_BoardLabelsOnly(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	labelRepo := new(MockLabelRepository)
	userLabelRepo := new(MockUserLabelRepository)
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").
		Return(true, domain.RoleMember, nil)
	settingsRepo.On("GetByBoardID", mock.Anything, "board-123").
		Return(&domain.BoardSettings{BoardID: "board-123", UseBoardLabelsOnly: true}, nil)
	labelRepo.On("ListByBoardID", mock.Anything, "board-123").
		Return([]*domain.Label{
			{ID: "bl-1", BoardID: "board-123", Name: "Bug", Color: "#ef4444"},
		}, nil)
	boardRepo.On("GetByID", mock.Anything, "board-123").
		Return(&domain.Board{ID: "board-123", OwnerID: "owner-123", Title: "Test", Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)

	uc := NewListAvailableLabelsUseCase(settingsRepo, labelRepo, userLabelRepo, boardRepo, memberRepo)
	result, err := uc.Execute(context.Background(), "board-123", "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.BoardLabels, 1)
	assert.Nil(t, result.UserLabels)
	assert.True(t, result.UseBoardLabelsOnly)

	// userLabelRepo.ListByUserID НЕ должен вызываться при UseBoardLabelsOnly=true
	userLabelRepo.AssertNotCalled(t, "ListByUserID")

	settingsRepo.AssertExpectations(t)
	labelRepo.AssertExpectations(t)
	boardRepo.AssertExpectations(t)
	memberRepo.AssertExpectations(t)
}

func TestListAvailableLabels_NonMember_Denied(t *testing.T) {
	settingsRepo := new(MockBoardSettingsRepository)
	labelRepo := new(MockLabelRepository)
	userLabelRepo := new(MockUserLabelRepository)
	boardRepo := new(MockBoardRepository)
	memberRepo := new(MockMembershipRepository)

	memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").
		Return(false, domain.Role(""), nil)

	uc := NewListAvailableLabelsUseCase(settingsRepo, labelRepo, userLabelRepo, boardRepo, memberRepo)
	result, err := uc.Execute(context.Background(), "board-123", "user-999")

	assert.Error(t, err)
	assert.Equal(t, domain.ErrAccessDenied, err)
	assert.Nil(t, result)

	memberRepo.AssertExpectations(t)
}
