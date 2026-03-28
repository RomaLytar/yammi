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

func TestAddMemberUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		boardID      string
		userID       string
		memberUserID string
		role         domain.Role
		setupMocks   func(*MockBoardRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "успешное добавление member'а owner'ом",
			boardID:      "board-123",
			userID:       "user-123",
			memberUserID: "user-456",
			role:         domain.RoleMember,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
				memberRepo.On("AddMember", mock.Anything, "board-123", "user-456", domain.RoleMember).Return(&domain.Member{UserID: "user-456", Role: domain.RoleMember}, nil)
			},
			wantErr: false,
		},
		{
			name:         "успешное добавление owner'а owner'ом",
			boardID:      "board-123",
			userID:       "user-123",
			memberUserID: "user-789",
			role:         domain.RoleOwner,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
				memberRepo.On("AddMember", mock.Anything, "board-123", "user-789", domain.RoleOwner).Return(&domain.Member{UserID: "user-789", Role: domain.RoleOwner}, nil)
			},
			wantErr: false,
		},
		{
			name:         "попытка добавить участника не owner'ом",
			boardID:      "board-123",
			userID:       "user-456",
			memberUserID: "user-789",
			role:         domain.RoleMember,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrNotOwner,
		},
		{
			name:         "невалидная роль",
			boardID:      "board-123",
			userID:       "user-123",
			memberUserID: "user-456",
			role:         domain.Role("invalid"),
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrInvalidRole,
		},
		{
			name:         "доска не найдена",
			boardID:      "board-999",
			userID:       "user-123",
			memberUserID: "user-456",
			role:         domain.RoleMember,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				boardRepo.On("GetByID", mock.Anything, "board-999").Return(nil, domain.ErrBoardNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrBoardNotFound,
		},
		{
			name:         "участник уже существует",
			boardID:      "board-123",
			userID:       "user-123",
			memberUserID: "user-456",
			role:         domain.RoleMember,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
				memberRepo.On("AddMember", mock.Anything, "board-123", "user-456", domain.RoleMember).Return(nil, domain.ErrMemberExists)
			},
			wantErr:     true,
			expectedErr: domain.ErrMemberExists,
		},
		{
			name:         "ошибка при добавлении участника",
			boardID:      "board-123",
			userID:       "user-123",
			memberUserID: "user-456",
			role:         domain.RoleMember,
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				board := &domain.Board{
					ID:          "board-123",
					Title:       "Test Board",
					Description: "Test Description",
					OwnerID:     "user-123",
					Version:     1,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				boardRepo.On("GetByID", mock.Anything, "board-123").Return(board, nil)
				memberRepo.On("AddMember", mock.Anything, "board-123", "user-456", domain.RoleMember).Return(nil, errors.New("database error"))
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
			publisher.On("PublishMemberAdded", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewAddMemberUseCase(boardRepo, memberRepo, publisher)
			member, err := useCase.Execute(context.Background(), tt.boardID, tt.userID, tt.memberUserID, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, member)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, member)
			}

			boardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			publisher.AssertExpectations(t)
		})
	}
}
