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

// MockBoardRepository - мок для BoardRepository
type MockBoardRepository struct {
	mock.Mock
}

func (m *MockBoardRepository) Create(ctx context.Context, board *domain.Board) error {
	args := m.Called(ctx, board)
	return args.Error(0)
}

func (m *MockBoardRepository) GetByID(ctx context.Context, boardID string) (*domain.Board, error) {
	args := m.Called(ctx, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Board), args.Error(1)
}

func (m *MockBoardRepository) ListByUserID(ctx context.Context, userID string, limit int, cursor string) ([]*domain.Board, string, error) {
	args := m.Called(ctx, userID, limit, cursor)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Board), args.String(1), args.Error(2)
}

func (m *MockBoardRepository) Update(ctx context.Context, board *domain.Board) error {
	args := m.Called(ctx, board)
	return args.Error(0)
}

func (m *MockBoardRepository) Delete(ctx context.Context, boardID string) error {
	args := m.Called(ctx, boardID)
	return args.Error(0)
}

// MockMembershipRepository - мок для MembershipRepository
type MockMembershipRepository struct {
	mock.Mock
}

func (m *MockMembershipRepository) AddMember(ctx context.Context, boardID, userID string, role domain.Role) error {
	args := m.Called(ctx, boardID, userID, role)
	return args.Error(0)
}

func (m *MockMembershipRepository) RemoveMember(ctx context.Context, boardID, userID string) error {
	args := m.Called(ctx, boardID, userID)
	return args.Error(0)
}

func (m *MockMembershipRepository) IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error) {
	args := m.Called(ctx, boardID, userID)
	return args.Bool(0), args.Get(1).(domain.Role), args.Error(2)
}

func (m *MockMembershipRepository) ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error) {
	args := m.Called(ctx, boardID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Member), args.Error(1)
}

// MockEventPublisher - мок для EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishBoardCreated(ctx context.Context, event // events.BoardCreated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBoardUpdated(ctx context.Context, event // events.BoardUpdated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishBoardDeleted(ctx context.Context, event // events.BoardDeleted) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishColumnCreated(ctx context.Context, event // events.ColumnCreated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishColumnUpdated(ctx context.Context, event // events.ColumnUpdated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishColumnDeleted(ctx context.Context, event // events.ColumnDeleted) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCardCreated(ctx context.Context, event // events.CardCreated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCardUpdated(ctx context.Context, event // events.CardUpdated) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCardMoved(ctx context.Context, event // events.CardMoved) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishCardDeleted(ctx context.Context, event // events.CardDeleted) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishMemberAdded(ctx context.Context, event // events.MemberAdded) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishMemberRemoved(ctx context.Context, event // events.MemberRemoved) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestCreateBoardUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		ownerID     string
		setupMocks  func(*MockBoardRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "успешное создание доски",
			title:       "Test Board",
			description: "Test Description",
			ownerID:     "user-123",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				boardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Board")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "пустой заголовок",
			title:       "",
			description: "Test Description",
			ownerID:     "user-123",
			setupMocks:  func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {},
			wantErr:     true,
			expectedErr: domain.ErrEmptyTitle,
		},
		{
			name:        "пустой ownerID",
			title:       "Test Board",
			description: "Test Description",
			ownerID:     "",
			setupMocks:  func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {},
			wantErr:     true,
			expectedErr: domain.ErrEmptyOwnerID,
		},
		{
			name:        "ошибка при сохранении в репозиторий",
			title:       "Test Board",
			description: "Test Description",
			ownerID:     "user-123",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				boardRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Board")).Return(errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)
			publisher := new(MockEventPublisher)

			tt.setupMocks(boardRepo, memberRepo, publisher)

			useCase := NewCreateBoardUseCase(boardRepo, memberRepo, publisher)
			board, err := useCase.Execute(context.Background(), tt.title, tt.description, tt.ownerID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, board)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, board)
				assert.Equal(t, tt.title, board.Title)
				assert.Equal(t, tt.description, board.Description)
				assert.Equal(t, tt.ownerID, board.OwnerID)
				assert.NotEmpty(t, board.ID)
				assert.Equal(t, 1, board.Version)
			}

			boardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			publisher.AssertExpectations(t)
		})
	}
}
