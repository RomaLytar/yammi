package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteAttachmentUseCase_Execute(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		attachmentID string
		boardID      string
		userID       string
		setupMocks   func(*MockAttachmentRepository, *MockMembershipRepository, *MockFileStorage, *MockEventPublisher)
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "загрузивший удаляет своё вложение",
			attachmentID: "att-123",
			boardID:      "board-123",
			userID:       "user-456",
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-456").Return(true, domain.RoleMember, nil)
				attachRepo.On("GetByID", mock.Anything, "att-123", "board-123").Return(&domain.Attachment{
					ID:         "att-123",
					CardID:     "card-123",
					BoardID:    "board-123",
					FileName:   "document.pdf",
					FileSize:   1024,
					MimeType:   "application/pdf",
					StorageKey: "boards/board-123/cards/card-123/att-123/document.pdf",
					UploaderID: "user-456",
					CreatedAt:  now,
				}, nil)
				storage.On("Delete", mock.Anything, "boards/board-123/cards/card-123/att-123/document.pdf").Return(nil)
				attachRepo.On("Delete", mock.Anything, "att-123", "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "owner удаляет чужое вложение",
			attachmentID: "att-123",
			boardID:      "board-123",
			userID:       "user-owner",
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-owner").Return(true, domain.RoleOwner, nil)
				attachRepo.On("GetByID", mock.Anything, "att-123", "board-123").Return(&domain.Attachment{
					ID:         "att-123",
					CardID:     "card-123",
					BoardID:    "board-123",
					FileName:   "document.pdf",
					FileSize:   1024,
					MimeType:   "application/pdf",
					StorageKey: "boards/board-123/cards/card-123/att-123/document.pdf",
					UploaderID: "user-other",
					CreatedAt:  now,
				}, nil)
				storage.On("Delete", mock.Anything, "boards/board-123/cards/card-123/att-123/document.pdf").Return(nil)
				attachRepo.On("Delete", mock.Anything, "att-123", "board-123").Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "member не может удалить чужое вложение",
			attachmentID: "att-123",
			boardID:      "board-123",
			userID:       "user-member",
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-member").Return(true, domain.RoleMember, nil)
				attachRepo.On("GetByID", mock.Anything, "att-123", "board-123").Return(&domain.Attachment{
					ID:         "att-123",
					CardID:     "card-123",
					BoardID:    "board-123",
					FileName:   "document.pdf",
					FileSize:   1024,
					MimeType:   "application/pdf",
					StorageKey: "boards/board-123/cards/card-123/att-123/document.pdf",
					UploaderID: "user-other",
					CreatedAt:  now,
				}, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:         "вложение не найдено",
			attachmentID: "att-999",
			boardID:      "board-123",
			userID:       "user-123",
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleOwner, nil)
				attachRepo.On("GetByID", mock.Anything, "att-999", "board-123").Return(nil, domain.ErrAttachmentNotFound)
			},
			wantErr:     true,
			expectedErr: domain.ErrAttachmentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachRepo := new(MockAttachmentRepository)
			memberRepo := new(MockMembershipRepository)
			storage := new(MockFileStorage)
			publisher := new(MockEventPublisher)

			tt.setupMocks(attachRepo, memberRepo, storage, publisher)
			publisher.On("PublishAttachmentDeleted", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewDeleteAttachmentUseCase(attachRepo, memberRepo, storage, publisher)
			err := useCase.Execute(context.Background(), tt.attachmentID, tt.boardID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			attachRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			storage.AssertExpectations(t)
		})
	}
}
