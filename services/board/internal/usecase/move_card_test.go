package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	// removed pkg/events
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMoveCardUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		cardID         string
		boardID        string
		fromColumnID   string
		toColumnID     string
		userID         string
		targetPosition string
		setupMocks     func(*MockCardRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "успешное перемещение в начало колонки",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: "a",
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)

				card := &domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(card, nil)

				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "успешное перемещение в конец колонки",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: "m",
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)

				card := &domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(card, nil)

				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "успешное перемещение между карточками",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: "b",
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)

				card := &domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(card, nil)

				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "успешное перемещение в пустую колонку",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: "a",
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)

				card := &domain.Card{
					ID:          "card-123",
					ColumnID:    "column-1",
					Title:       "Test Card",
					Description: "Description",
					Position:    "n",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(card, nil)
				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(MockCardRepository)
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(cardRepo, memberRepo, publisher)
			cardRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
			boardRepo.On("TouchUpdatedAt", mock.Anything, mock.Anything).Return(nil).Maybe()
			publisher.On("PublishCardMoved", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewMoveCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
			card, err := useCase.Execute(context.Background(), tt.cardID, tt.boardID, tt.fromColumnID, tt.toColumnID, tt.userID, tt.targetPosition)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, card)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, card)
				assert.Equal(t, tt.toColumnID, card.ColumnID)
				assert.NotEmpty(t, card.Position)
			}

			cardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			publisher.AssertExpectations(t)
		})
	}
}
