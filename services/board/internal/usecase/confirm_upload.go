package usecase

import (
	"context"
	"log/slog"

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

	// 3. Проверяем что файл загружен в хранилище и соответствует метаданным
	actualSize, actualType, err := uc.storage.Stat(ctx, attachment.StorageKey)
	if err != nil {
		return nil, domain.ErrAttachmentNotFound
	}

	// 4. Проверяем что реальный размер соответствует заявленному
	if actualSize != attachment.FileSize {
		uc.cleanupOrphan(ctx, attachment)
		return nil, domain.ErrFileSizeMismatch
	}

	// 5. Проверяем что реальный content-type соответствует заявленному
	if attachment.MimeType != "" && actualType != "" && actualType != attachment.MimeType {
		uc.cleanupOrphan(ctx, attachment)
		return nil, domain.ErrFileTypeMismatch
	}

	return attachment, nil
}

// cleanupOrphan удаляет S3-объект и метаданные вложения при несоответствии size/type,
// чтобы не накапливались orphan-объекты и не расходовались attachment-слоты.
func (uc *ConfirmUploadUseCase) cleanupOrphan(ctx context.Context, attachment *domain.Attachment) {
	if err := uc.storage.Delete(ctx, attachment.StorageKey); err != nil {
		slog.Error("failed to delete orphan S3 object", "error", err, "key", attachment.StorageKey)
	}
	if err := uc.attachmentRepo.Delete(ctx, attachment.ID, attachment.BoardID); err != nil {
		slog.Error("failed to delete orphan attachment metadata", "error", err, "id", attachment.ID)
	}
}
