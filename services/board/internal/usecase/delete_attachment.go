package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteAttachmentUseCase struct {
	attachmentRepo AttachmentRepository
	activityRepo   ActivityRepository
	memberRepo     MembershipRepository
	storage        FileStorage
	publisher      EventPublisher
}

func NewDeleteAttachmentUseCase(attachmentRepo AttachmentRepository, activityRepo ActivityRepository, memberRepo MembershipRepository, storage FileStorage, publisher EventPublisher) *DeleteAttachmentUseCase {
	return &DeleteAttachmentUseCase{
		attachmentRepo: attachmentRepo,
		activityRepo:   activityRepo,
		memberRepo:     memberRepo,
		storage:        storage,
		publisher:      publisher,
	}
}

func (uc *DeleteAttachmentUseCase) Execute(ctx context.Context, attachmentID, boardID, userID string) error {
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	attachment, err := uc.attachmentRepo.GetByID(ctx, attachmentID, boardID)
	if err != nil {
		return err
	}

	if attachment.UploaderID != userID && role != domain.RoleOwner {
		return domain.ErrAccessDenied
	}

	_ = uc.storage.Delete(ctx, attachment.StorageKey)

	if err := uc.attachmentRepo.Delete(ctx, attachmentID, boardID); err != nil {
		return err
	}

	// Записываем в историю
	activity, _ := domain.NewActivity(attachment.CardID, boardID, userID, domain.ActivityAttachmentDeleted,
		fmt.Sprintf("Файл \"%s\" удалён", attachment.FileName),
		map[string]string{"file_name": attachment.FileName})
	if activity != nil {
		_ = uc.activityRepo.Create(ctx, activity)
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishAttachmentDeleted(ctx, AttachmentDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   attachment.CreatedAt,
			AttachmentID: attachment.ID,
			CardID:       attachment.CardID,
			BoardID:      attachment.BoardID,
			ActorID:      userID,
			FileName:     attachment.FileName,
		}); err != nil {
			slog.Error("failed to publish AttachmentDeleted", "error", err, "attachment_id", attachment.ID, "board_id", attachment.BoardID)
		}
	}()

	return nil
}
