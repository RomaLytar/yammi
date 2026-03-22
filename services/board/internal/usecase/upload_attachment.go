package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UploadAttachmentUseCase struct {
	attachmentRepo AttachmentRepository
	activityRepo   ActivityRepository
	memberRepo     MembershipRepository
	storage        FileStorage
	publisher      EventPublisher
}

func NewUploadAttachmentUseCase(attachmentRepo AttachmentRepository, activityRepo ActivityRepository, memberRepo MembershipRepository, storage FileStorage, publisher EventPublisher) *UploadAttachmentUseCase {
	return &UploadAttachmentUseCase{
		attachmentRepo: attachmentRepo,
		activityRepo:   activityRepo,
		memberRepo:     memberRepo,
		storage:        storage,
		publisher:      publisher,
	}
}

// Execute создает метаданные вложения и возвращает pre-signed URL для загрузки
func (uc *UploadAttachmentUseCase) Execute(ctx context.Context, cardID, boardID, userID, fileName, contentType string, fileSize int64) (*domain.Attachment, string, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, "", err
	}
	if !isMember {
		return nil, "", domain.ErrAccessDenied
	}

	// 2. Проверка лимита вложений
	count, err := uc.attachmentRepo.CountByCardID(ctx, cardID, boardID)
	if err != nil {
		return nil, "", err
	}
	if count >= domain.MaxAttachmentsPerCard {
		return nil, "", domain.ErrMaxAttachmentsReached
	}

	// 3. Создаем доменную сущность (валидация внутри)
	attachment, err := domain.NewAttachment(cardID, boardID, fileName, fileSize, contentType, userID)
	if err != nil {
		return nil, "", err
	}

	// 4. Генерируем pre-signed URL для загрузки
	uploadURL, err := uc.storage.GenerateUploadURL(ctx, attachment.StorageKey, contentType, fileSize)
	if err != nil {
		log.Printf("ERROR: GenerateUploadURL failed: %v", err)
		return nil, "", err
	}

	// 5. Сохраняем метаданные
	if err := uc.attachmentRepo.Create(ctx, attachment); err != nil {
		log.Printf("ERROR: attachment repo create failed: %v", err)
		return nil, "", err
	}

	// 5.5 Записываем в историю
	activity, _ := domain.NewActivity(cardID, boardID, userID, domain.ActivityAttachmentAdded,
		fmt.Sprintf("Файл \"%s\" прикреплён", fileName),
		map[string]string{"file_name": fileName, "attachment_id": attachment.ID})
	if activity != nil {
		_ = uc.activityRepo.Create(ctx, activity)
	}

	// 6. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishAttachmentUploaded(context.Background(), AttachmentUploaded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   attachment.CreatedAt,
			AttachmentID: attachment.ID,
			CardID:       attachment.CardID,
			BoardID:      attachment.BoardID,
			ActorID:      userID,
			FileName:     attachment.FileName,
			FileSize:     attachment.FileSize,
		})
	}()

	return attachment, uploadURL, nil
}
