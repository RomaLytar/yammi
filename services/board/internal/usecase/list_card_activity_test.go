package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCardActivityUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		cardID         string
		boardID        string
		userID         string
		limit          int
		cursor         string
		setupMocks     func(*MockActivityRepository, *MockMembershipRepository)
		wantErr        bool
		expectedErr    error
		expectedCount  int
		expectedCursor string
	}{
		{
			name:    "успешный список с результатами",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-123",
			limit:   20,
			cursor:  "",
			setupMocks: func(activityRepo *MockActivityRepository, memberRepo *MockMembershipRepository) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				activities := []*domain.Activity{
					{
						ID:          "act-1",
						CardID:      "card-123",
						BoardID:     "board-123",
						ActorID:     "user-123",
						Type:        domain.ActivityCardCreated,
						Description: "Карточка создана",
						Changes:     map[string]string{},
						CreatedAt:   now,
					},
					{
						ID:          "act-2",
						CardID:      "card-123",
						BoardID:     "board-123",
						ActorID:     "user-456",
						Type:        domain.ActivityCardMoved,
						Description: "Перемещена из 'To Do' в 'In Progress'",
						Changes:     map[string]string{"old_column": "To Do", "new_column": "In Progress"},
						CreatedAt:   now,
					},
				}
				activityRepo.On("ListByCardID", mock.Anything, "card-123", "board-123", 20, "").Return(activities, "next-cursor", nil)
			},
			wantErr:        false,
			expectedCount:  2,
			expectedCursor: "next-cursor",
		},
		{
			name:    "пользователь не является участником доски",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-999",
			limit:   20,
			cursor:  "",
			setupMocks: func(activityRepo *MockActivityRepository, memberRepo *MockMembershipRepository) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:    "пустой результат (нет активности)",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-123",
			limit:   20,
			cursor:  "",
			setupMocks: func(activityRepo *MockActivityRepository, memberRepo *MockMembershipRepository) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleMember, nil)
				activityRepo.On("ListByCardID", mock.Anything, "card-123", "board-123", 20, "").Return([]*domain.Activity{}, "", nil)
			},
			wantErr:        false,
			expectedCount:  0,
			expectedCursor: "",
		},
		{
			name:    "лимит 0 — передается как есть",
			cardID:  "card-123",
			boardID: "board-123",
			userID:  "user-123",
			limit:   0,
			cursor:  "",
			setupMocks: func(activityRepo *MockActivityRepository, memberRepo *MockMembershipRepository) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleMember, nil)
				activityRepo.On("ListByCardID", mock.Anything, "card-123", "board-123", 0, "").Return([]*domain.Activity{}, "", nil)
			},
			wantErr:        false,
			expectedCount:  0,
			expectedCursor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activityRepo := new(MockActivityRepository)
			memberRepo := new(MockMembershipRepository)

			tt.setupMocks(activityRepo, memberRepo)

			useCase := NewListCardActivityUseCase(activityRepo, memberRepo)
			activities, cursor, err := useCase.Execute(context.Background(), tt.cardID, tt.boardID, tt.userID, tt.limit, tt.cursor)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, activities)
			} else {
				assert.NoError(t, err)
				assert.Len(t, activities, tt.expectedCount)
				assert.Equal(t, tt.expectedCursor, cursor)
			}

			activityRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
		})
	}
}
