package usecase

import (
	"context"
	"errors"
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
		targetPosition int
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
			targetPosition: 0,
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

				cardsInColumn := []*domain.Card{
					{ID: "card-1", ColumnID: "column-2", Position: "m", Title: "Card 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: "card-2", ColumnID: "column-2", Position: "o", Title: "Card 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				cardRepo.On("ListByColumnID", mock.Anything, "column-2").Return(cardsInColumn, nil)
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
			targetPosition: 10,
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

				cardsInColumn := []*domain.Card{
					{ID: "card-1", ColumnID: "column-2", Position: "m", Title: "Card 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: "card-2", ColumnID: "column-2", Position: "o", Title: "Card 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				cardRepo.On("ListByColumnID", mock.Anything, "column-2").Return(cardsInColumn, nil)
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
			targetPosition: 1,
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

				cardsInColumn := []*domain.Card{
					{ID: "card-1", ColumnID: "column-2", Position: "m", Title: "Card 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
					{ID: "card-2", ColumnID: "column-2", Position: "o", Title: "Card 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				}
				cardRepo.On("ListByColumnID", mock.Anything, "column-2").Return(cardsInColumn, nil)
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
			targetPosition: 0,
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

				cardsInColumn := []*domain.Card{}
				cardRepo.On("ListByColumnID", mock.Anything, "column-2").Return(cardsInColumn, nil)
				cardRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:           "пользователь не является участником",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-999",
			targetPosition: 0,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:           "карточка не найдена",
			cardID:         "card-999",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: 0,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("GetByID", mock.Anything, "card-999").Return(nil, domain.ErrCardNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrCardNotFound,
		},
		{
			name:           "ошибка при получении карточек колонки",
			cardID:         "card-123",
			boardID:        "board-123",
			fromColumnID:   "column-1",
			toColumnID:     "column-2",
			userID:         "user-123",
			targetPosition: 0,
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
				cardRepo.On("ListByColumnID", mock.Anything, "column-2").Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(MockCardRepository)
			memberRepo := new(MockMembershipRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(cardRepo, memberRepo, publisher)

			useCase := NewMoveCardUseCase(cardRepo, memberRepo, publisher)
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
