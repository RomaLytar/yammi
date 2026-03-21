package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	// removed pkg/events
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockColumnRepository - мок для ColumnRepository
type MockColumnRepository struct {
	mock.Mock
}

func (m *MockColumnRepository) Create(ctx context.Context, column *domain.Column) error {
	args := m.Called(ctx, column)
	return args.Error(0)
}

func (m *MockColumnRepository) GetByID(ctx context.Context, columnID string) (*domain.Column, error) {
	args := m.Called(ctx, columnID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Column), args.Error(1)
}

func (m *MockColumnRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Column, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Column), args.Error(1)
}

func (m *MockColumnRepository) Update(ctx context.Context, column *domain.Column) error {
	args := m.Called(ctx, column)
	return args.Error(0)
}

func (m *MockColumnRepository) Delete(ctx context.Context, columnID string) error {
	args := m.Called(ctx, columnID)
	return args.Error(0)
}

func TestAddColumnUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		boardID     string
		userID      string
		title       string
		position    int
		setupMocks  func(*MockColumnRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "успешное добавление колонки owner'ом",
			boardID:  "board-123",
			userID:   "user-123",
			title:    "To Do",
			position: 0,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				columnRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Column")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "успешное добавление колонки member'ом",
			boardID:  "board-123",
			userID:   "user-456",
			title:    "In Progress",
			position: 1,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
				columnRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Column")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "пользователь не является участником",
			boardID:  "board-123",
			userID:   "user-999",
			title:    "Done",
			position: 2,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:     "пустой заголовок",
			boardID:  "board-123",
			userID:   "user-123",
			title:    "",
			position: 0,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrEmptyColumnTitle,
		},
		{
			name:     "отрицательная позиция",
			boardID:  "board-123",
			userID:   "user-123",
			title:    "Test Column",
			position: -1,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidPosition,
		},
		{
			name:     "ошибка при сохранении",
			boardID:  "board-123",
			userID:   "user-123",
			title:    "Test Column",
			position: 0,
			setupMocks: func(columnRepo *MockColumnRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				columnRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Column")).Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			columnRepo := new(MockColumnRepository)
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(columnRepo, memberRepo, publisher)
			boardRepo.On("TouchUpdatedAt", mock.Anything, mock.Anything).Return(nil).Maybe()
			publisher.On("PublishColumnCreated", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewAddColumnUseCase(columnRepo, boardRepo, memberRepo, publisher)
			column, err := useCase.Execute(context.Background(), tt.boardID, tt.userID, tt.title, tt.position)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, column)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, column)
				assert.Equal(t, tt.title, column.Title)
				assert.Equal(t, tt.boardID, column.BoardID)
				assert.Equal(t, tt.position, column.Position)
				assert.NotEmpty(t, column.ID)
			}

			columnRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			publisher.AssertExpectations(t)
		})
	}
}
