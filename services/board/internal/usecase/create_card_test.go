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

// MockCardRepository - мок для CardRepository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) Create(ctx context.Context, card *domain.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepository) GetByID(ctx context.Context, cardID, boardID string) (*domain.Card, error) {
	args := m.Called(ctx, cardID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Card), args.Error(1)
}

func (m *MockCardRepository) GetLastInColumn(ctx context.Context, columnID string) (*domain.Card, error) {
	args := m.Called(ctx, columnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Card), args.Error(1)
}

func (m *MockCardRepository) ListByColumnID(ctx context.Context, columnID string) ([]*domain.Card, error) {
	args := m.Called(ctx, columnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Card), args.Error(1)
}

func (m *MockCardRepository) Update(ctx context.Context, card *domain.Card) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockCardRepository) Delete(ctx context.Context, cardID, boardID string) error {
	args := m.Called(ctx, cardID, boardID)
	return args.Error(0)
}

func (m *MockCardRepository) BatchDelete(ctx context.Context, boardID string, cardIDs []string) error {
	args := m.Called(ctx, boardID, cardIDs)
	return args.Error(0)
}

func (m *MockCardRepository) CountByBoard(ctx context.Context, boardID string) (map[string]int, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockCardRepository) UnassignByUser(ctx context.Context, boardID, userID string) (int, error) {
	args := m.Called(ctx, boardID, userID)
	return args.Int(0), args.Error(1)
}

type MockActivityRepository struct {
	mock.Mock
}

func (m *MockActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockActivityRepository) ListByCardID(ctx context.Context, cardID, boardID string, limit int, cursor string) ([]*domain.Activity, string, error) {
	args := m.Called(ctx, cardID, boardID, limit, cursor)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Activity), args.String(1), args.Error(2)
}

func TestCreateCardUseCase_Execute(t *testing.T) {
	assigneeID := "user-789"

	tests := []struct {
		name        string
		columnID    string
		boardID     string
		userID      string
		title       string
		description string
		position    string
		assigneeID  *string
		setupMocks  func(*MockCardRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "успешное создание карточки с явной позицией",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "Test Card",
			description: "Test Description",
			position:    "n",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "успешное создание карточки без позиции (пустая колонка)",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "Test Card",
			description: "Test Description",
			position:    "",
			assigneeID:  &assigneeID,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-789").Return(true, domain.RoleMember, nil)
				cardRepo.On("GetLastInColumn", mock.Anything, "column-123").Return(nil, domain.ErrCardNotFound)
				cardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "успешное создание карточки без позиции (в конец колонки)",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "Test Card",
			description: "Test Description",
			position:    "",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				lastCard := &domain.Card{
					ID:          "card-last",
					ColumnID:    "column-123",
					Title:       "Last Card",
					Description: "",
					Position:    "n",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				cardRepo.On("GetLastInColumn", mock.Anything, "column-123").Return(lastCard, nil)
				cardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Card")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "пользователь не является участником",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-999",
			title:       "Test Card",
			description: "Test Description",
			position:    "n",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:        "пустой заголовок",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "",
			description: "Test Description",
			position:    "n",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrEmptyCardTitle,
		},
		{
			name:        "невалидная позиция lexorank",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "Test Card",
			description: "Test Description",
			position:    "INVALID@#$",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidLexorank,
		},
		{
			name:        "ошибка при получении последней карточки",
			columnID:    "column-123",
			boardID:     "board-123",
			userID:      "user-123",
			title:       "Test Card",
			description: "Test Description",
			position:    "",
			assigneeID:  nil,
			setupMocks: func(cardRepo *MockCardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("GetLastInColumn", mock.Anything, "column-123").Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cardRepo := new(MockCardRepository)
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(cardRepo, memberRepo, publisher)
			boardRepo.On("TouchUpdatedAt", mock.Anything, mock.Anything).Return(nil).Maybe()
			publisher.On("PublishCardCreated", mock.Anything, mock.Anything).Return(nil).Maybe()
			publisher.On("PublishCardAssigned", mock.Anything, mock.Anything).Return(nil).Maybe()

			activityRepo := new(MockActivityRepository)
			activityRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewCreateCardUseCase(cardRepo, boardRepo, memberRepo, activityRepo, publisher)
			card, err := useCase.Execute(context.Background(), tt.columnID, tt.boardID, tt.userID, tt.title, tt.description, tt.position, tt.assigneeID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				}
				assert.Nil(t, card)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, card)
				assert.Equal(t, tt.title, card.Title)
				assert.Equal(t, tt.description, card.Description)
				assert.Equal(t, tt.columnID, card.ColumnID)
				assert.NotEmpty(t, card.ID)
				assert.NotEmpty(t, card.Position)
			}

			cardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			publisher.AssertExpectations(t)
		})
	}
}
