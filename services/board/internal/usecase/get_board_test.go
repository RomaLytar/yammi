package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBoardUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		boardID     string
		userID      string
		setupMocks  func(*MockBoardRepository, *MockMembershipRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "успешное получение доски для owner",
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository) {
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
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
			},
			wantErr: false,
		},
		{
			name:    "успешное получение доски для member",
			boardID: "board-123",
			userID:  "user-456",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository) {
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
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
			},
			wantErr: false,
		},
		{
			name:    "доска не найдена",
			boardID: "board-999",
			userID:  "user-123",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository) {
				boardRepo.On("GetByID", mock.Anything, "board-999").Return(nil, domain.ErrBoardNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrBoardNotFound,
		},
		{
			name:    "пользователь не является участником",
			boardID: "board-123",
			userID:  "user-999",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository) {
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
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:    "ошибка при проверке членства",
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository) {
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
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(false, domain.Role(""), errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boardRepo := new(MockBoardRepository)
			memberRepo := new(MockMembershipRepository)

			tt.setupMocks(boardRepo, memberRepo)

			useCase := NewGetBoardUseCase(boardRepo, memberRepo)
			board, err := useCase.Execute(context.Background(), tt.boardID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, board)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, board)
				assert.Equal(t, tt.boardID, board.ID)
			}

			boardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
		})
	}
}
