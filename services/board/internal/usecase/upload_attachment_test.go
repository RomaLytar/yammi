package usecase

import (
	"context"
	"testing"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAttachmentRepository - мок для AttachmentRepository
type MockAttachmentRepository struct {
	mock.Mock
}

func (m *MockAttachmentRepository) Create(ctx context.Context, attachment *domain.Attachment) error {
	args := m.Called(ctx, attachment)
	return args.Error(0)
}

func (m *MockAttachmentRepository) GetByID(ctx context.Context, attachmentID, boardID string) (*domain.Attachment, error) {
	args := m.Called(ctx, attachmentID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Attachment, error) {
	args := m.Called(ctx, cardID, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Attachment), args.Error(1)
}

func (m *MockAttachmentRepository) Delete(ctx context.Context, attachmentID, boardID string) error {
	args := m.Called(ctx, attachmentID, boardID)
	return args.Error(0)
}

func (m *MockAttachmentRepository) CountByCardID(ctx context.Context, cardID, boardID string) (int, error) {
	args := m.Called(ctx, cardID, boardID)
	return args.Int(0), args.Error(1)
}

// MockFileStorage - мок для FileStorage
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) GenerateUploadURL(ctx context.Context, key, contentType string, size int64) (string, error) {
	args := m.Called(ctx, key, contentType, size)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockFileStorage) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func TestUploadAttachmentUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		cardID      string
		boardID     string
		userID      string
		fileName    string
		contentType string
		fileSize    int64
		setupMocks  func(*MockAttachmentRepository, *MockMembershipRepository, *MockFileStorage, *MockEventPublisher)
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "успешная загрузка (возвращает pre-signed URL)",
			cardID:      "card-123",
			boardID:     "board-123",
			userID:      "user-123",
			fileName:    "document.pdf",
			contentType: "application/pdf",
			fileSize:    1024,
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleMember, nil)
				attachRepo.On("CountByCardID", mock.Anything, "card-123", "board-123").Return(0, nil)
				storage.On("GenerateUploadURL", mock.Anything, mock.AnythingOfType("string"), "application/pdf", int64(1024)).Return("https://minio/presigned-upload-url", nil)
				attachRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Attachment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "пользователь не является участником доски",
			cardID:      "card-123",
			boardID:     "board-123",
			userID:      "user-999",
			fileName:    "document.pdf",
			contentType: "application/pdf",
			fileSize:    1024,
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-999").Return(false, domain.Role(""), nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrAccessDenied,
		},
		{
			name:        "достигнут лимит вложений",
			cardID:      "card-123",
			boardID:     "board-123",
			userID:      "user-123",
			fileName:    "document.pdf",
			contentType: "application/pdf",
			fileSize:    1024,
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleMember, nil)
				attachRepo.On("CountByCardID", mock.Anything, "card-123", "board-123").Return(domain.MaxAttachmentsPerCard, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrMaxAttachmentsReached,
		},
		{
			name:        "файл слишком большой (доменная валидация)",
			cardID:      "card-123",
			boardID:     "board-123",
			userID:      "user-123",
			fileName:    "huge-file.zip",
			contentType: "application/zip",
			fileSize:    domain.MaxFileSize + 1,
			setupMocks: func(attachRepo *MockAttachmentRepository, memberRepo *MockMembershipRepository, storage *MockFileStorage, publisher *MockEventPublisher) {
				memberRepo.On("IsMember", mock.Anything, "board-123", "user-123").Return(true, domain.RoleMember, nil)
				attachRepo.On("CountByCardID", mock.Anything, "card-123", "board-123").Return(0, nil)
			},
			wantErr:     true,
			expectedErr: domain.ErrFileTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attachRepo := new(MockAttachmentRepository)
			memberRepo := new(MockMembershipRepository)
			storage := new(MockFileStorage)
			publisher := new(MockEventPublisher)

			tt.setupMocks(attachRepo, memberRepo, storage, publisher)
			publisher.On("PublishAttachmentUploaded", mock.Anything, mock.Anything).Return(nil).Maybe()

			useCase := NewUploadAttachmentUseCase(attachRepo, memberRepo, storage, publisher)
			attachment, uploadURL, err := useCase.Execute(context.Background(), tt.cardID, tt.boardID, tt.userID, tt.fileName, tt.contentType, tt.fileSize)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
				assert.Nil(t, attachment)
				assert.Empty(t, uploadURL)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
				assert.NotEmpty(t, uploadURL)
				assert.Equal(t, tt.fileName, attachment.FileName)
				assert.Equal(t, tt.fileSize, attachment.FileSize)
				assert.Equal(t, tt.userID, attachment.UploaderID)
			}

			attachRepo.AssertExpectations(t)
			memberRepo.AssertExpectations(t)
			storage.AssertExpectations(t)
		})
	}
}
