package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ConfirmUploadUseCase struct {
	attachmentRepo AttachmentRepository
	memberRepo     MembershipRepository
	storage        FileStorage
}

func NewConfirmUploadUseCase(attachmentRepo AttachmentRepository, memberRepo MembershipRepository, storage FileStorage) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{
		attachmentRepo: attachmentRepo,
		memberRepo:     memberRepo,
		storage:        storage,
	}
}

// Execute подтверждает загрузку файла — проверяет что файл реально загружен в хранилище
func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, attachmentID, boardID, userID string) (*domain.Attachment, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем вложение
	attachment, err := uc.attachmentRepo.GetByID(ctx, attachmentID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Проверяем что файл загружен в хранилище
	exists, err := uc.storage.Exists(ctx, attachment.StorageKey)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrAttachmentNotFound
	}

	return attachment, nil
}
