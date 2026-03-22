package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetDownloadURLUseCase struct {
	attachmentRepo AttachmentRepository
	memberRepo     MembershipRepository
	storage        FileStorage
}

func NewGetDownloadURLUseCase(attachmentRepo AttachmentRepository, memberRepo MembershipRepository, storage FileStorage) *GetDownloadURLUseCase {
	return &GetDownloadURLUseCase{
		attachmentRepo: attachmentRepo,
		memberRepo:     memberRepo,
		storage:        storage,
	}
}

// Execute генерирует pre-signed URL для скачивания вложения
func (uc *GetDownloadURLUseCase) Execute(ctx context.Context, attachmentID, boardID, userID string) (string, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return "", err
	}
	if !isMember {
		return "", domain.ErrAccessDenied
	}

	// 2. Получаем вложение
	attachment, err := uc.attachmentRepo.GetByID(ctx, attachmentID, boardID)
	if err != nil {
		return "", err
	}

	// 3. Генерируем pre-signed URL для скачивания
	return uc.storage.GenerateDownloadURL(ctx, attachment.StorageKey)
}
