package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnassignCardUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		cardID      string
		boardID     string
		userID      string
		setupMocks  func(*MockCardRepository, *MockBoardRepository, *MockMembershipRepository, *MockActivityRepository, *MockEventPublisher)
		wantErr     bool
		expectedErr error
		wantNilAssignee bool
	}{
		{
			name:    "успешное снятие назначения",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, activityRepo *MockActivityRepository, publisher *MockEventPublisher) {
				assignee := "user-456"
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("GetByID", mock.Anything, "card-123", "board-123").Return(&domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					AssigneeID:  &assignee,
					CreatorID:   "user-123",
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
				activityRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr:         false,
			wantNilAssignee: true,
		},
		{
			name:    "актор не является участником доски",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-999",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, activityRepo *MockActivityRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:    "карточка уже без назначения — возвращается без ошибки",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, activityRepo *MockActivityRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("GetByID", mock.Anything, "card-123", "board-123").Return(&domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					AssigneeID:  nil,
					CreatorID:   "user-123",
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			wantErr:         false,
			wantNilAssignee: true,
		},
		{
			name:    "карточка не найдена",
			cardID:  "card-999",
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, activityRepo *MockActivityRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("GetByID", mock.Anything, "card-999", "board-123").Return(nil, domain.ErrCardNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrCardNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(MockCardRepository)
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)
			activityRepo := new(MockActivityRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
			boardRepo.On("TouchUpdatedAt", mock.Anything, mock.Anything).Return(nil).Maybe()
			publisher.On("PublishCardUnassigned", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewUnassignCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
			card, err := useCase.Execute(context.Background(), tt.cardID, tt.boardID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, card)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, card)
				if tt.wantNilAssignee {
					assert.Nil(t, card.AssigneeID)
				}
			}

			cardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			activityRepo.AssertExpectations(t)
		})
	}
}
