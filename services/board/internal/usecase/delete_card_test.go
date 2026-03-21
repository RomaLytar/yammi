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

func TestDeleteCardUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		cardIDs     []string
		boardID     string
		userID      string
		setupMocks  func(*MockCardRepository, *MockBoardRepository, *MockMembershipRepository, *MockEventPublisher)
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "успешное удаление карточки owner'ом (чужая карточка)",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("BatchDelete", mock.Anything, "board-123", []string{"card-123"}).Return(nil)
				boardRepo.On("TouchUpdatedAt", mock.Anything, "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "успешное удаление карточки создателем (member)",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-456",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(&domain.Card{
					ID:        "card-123",
					ColumnID:  "column-1",
					Title:     "Test Card",
					Position:  "n",
					CreatorID: "user-456",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				cardRepo.On("BatchDelete", mock.Anything, "board-123", []string{"card-123"}).Return(nil)
				boardRepo.On("TouchUpdatedAt", mock.Anything, "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "доступ запрещен — не участник доски",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-999",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:    "доступ запрещен — member пытается удалить чужую карточку",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-456",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
				cardRepo.On("GetByID", mock.Anything, "card-123").Return(&domain.Card{
					ID:        "card-123",
					ColumnID:  "column-1",
					Title:     "Test Card",
					Position:  "n",
					CreatorID: "user-other",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:    "успешное batch-удаление нескольких карточек owner'ом",
			cardIDs: []string{"card-1", "card-2", "card-3"},
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("BatchDelete", mock.Anything, "board-123", []string{"card-1", "card-2", "card-3"}).Return(nil)
				boardRepo.On("TouchUpdatedAt", mock.Anything, "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "карточка не найдена при проверке создателя",
			cardIDs: []string{"card-999"},
			boardID: "board-123",
			userID:  "user-456",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
				cardRepo.On("GetByID", mock.Anything, "card-999").Return(nil, domain.ErrCardNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrCardNotFound,
		},
		{
			name:    "пустой список cardIDs — ничего не происходит",
			cardIDs: []string{},
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("BatchDelete", mock.Anything, "board-123", []string{}).Return(nil)
				boardRepo.On("TouchUpdatedAt", mock.Anything, "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "ошибка при проверке членства",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(false, domain.Role(""), errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: errors.New("database error"),
		},
		{
			name:    "ошибка при batch-удалении",
			cardIDs: []string{"card-123"},
			boardID: "board-123",
			userID:  "user-123",
			setupMocks: func(cardRepo *MockCardRepository, boardRepo *MockBoardRepository, memberRepo *MockMembershipRepository, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				cardRepo.On("BatchDelete", mock.Anything, "board-123", []string{"card-123"}).Return(errors.New("database error"))
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

			tt.setupMocks(cardRepo, boardRepo, memberRepo, publisher)
			publisher.On("PublishCardDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()
			boardRepo.On("TouchUpdatedAt", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewDeleteCardUseCase(cardRepo, boardRepo, memberRepo, publisher)
			err := useCase.Execute(context.Background(), tt.cardIDs, tt.boardID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			cardRepo.AssertExpectations(t)
			boardRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
		})
	}
}
