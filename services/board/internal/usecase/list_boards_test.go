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

func TestListBoardsUseCase_Execute(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		limit          int
		cursor         string
		setupMocks     func(*MockBoardRepository)
		wantErr        bool
		expectedErr    error
		expectedLimit  int
		expectedCount  int
		expectedCursor string
	}{
		{
			name:   "успешное получение списка досок с дефолтным лимитом",
			userID: "user-123",
			limit:  0,
			cursor: "",
			setupMocks: func(boardRepo *MockBoardRepository) {
				boards := []*domain.Board{
					{
						ID:          "board-1",
						Title:       "Board 1",
						Description: "Description 1",
						OwnerID:     "user-123",
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
					{
						ID:          "board-2",
						Title:       "Board 2",
						Description: "Description 2",
						OwnerID:     "user-123",
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				boardRepo.On("ListByUserID", mock.Anything, "user-123", 20, "", false, "", "updated_at").Return(boards, "cursor-next", nil)
			},
			wantErr:        false,
			expectedLimit:  20,
			expectedCount:  2,
			expectedCursor: "cursor-next",
		},
		{
			name:   "успешное получение списка досок с кастомным лимитом",
			userID: "user-123",
			limit:  50,
			cursor: "cursor-123",
			setupMocks: func(boardRepo *MockBoardRepository) {
				boards := []*domain.Board{
					{
						ID:          "board-3",
						Title:       "Board 3",
						Description: "Description 3",
						OwnerID:     "user-123",
						Version:     1,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					},
				}
				boardRepo.On("ListByUserID", mock.Anything, "user-123", 50, "cursor-123", false, "", "updated_at").Return(boards, "cursor-next-2", nil)
			},
			wantErr:        false,
			expectedLimit:  50,
			expectedCount:  1,
			expectedCursor: "cursor-next-2",
		},
		{
			name:   "лимит больше 100 - используется дефолтный",
			userID: "user-123",
			limit:  150,
			cursor: "",
			setupMocks: func(boardRepo *MockBoardRepository) {
				boards := []*domain.Board{}
				boardRepo.On("ListByUserID", mock.Anything, "user-123", 20, "", false, "", "updated_at").Return(boards, "", nil)
			},
			wantErr:        false,
			expectedLimit:  20,
			expectedCount:  0,
			expectedCursor: "",
		},
		{
			name:   "отрицательный лимит - используется дефолтный",
			userID: "user-123",
			limit:  -5,
			cursor: "",
			setupMocks: func(boardRepo *MockBoardRepository) {
				boards := []*domain.Board{}
				boardRepo.On("ListByUserID", mock.Anything, "user-123", 20, "", false, "", "updated_at").Return(boards, "", nil)
			},
			wantErr:        false,
			expectedLimit:  20,
			expectedCount:  0,
			expectedCursor: "",
		},
		{
			name:   "ошибка при получении списка",
			userID: "user-123",
			limit:  20,
			cursor: "",
			setupMocks: func(boardRepo *MockBoardRepository) {
				boardRepo.On("ListByUserID", mock.Anything, "user-123", 20, "", false, "", "updated_at").Return(nil, "", errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boardRepo := new(MockBoardRepository)

			tt.setupMocks(boardRepo)

			useCase := NewListBoardsUseCase(boardRepo)
			boards, cursor, err := useCase.Execute(context.Background(), tt.userID, tt.limit, tt.cursor, false, "", "updated_at")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, boards)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, boards)
				assert.Equal(t, tt.expectedCount, len(boards))
				assert.Equal(t, tt.expectedCursor, cursor)
			}

			boardRepo.AssertExpectations(t)
		})
	}
}
