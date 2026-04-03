package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteReleaseUseCase struct {
	releaseRepo ReleaseRepository
	cardRepo    CardRepository
	memberRepo  MembershipRepository
	publisher   EventPublisher
}

func NewDeleteReleaseUseCase(releaseRepo ReleaseRepository, cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteReleaseUseCase {
	return &DeleteReleaseUseCase{
		releaseRepo: releaseRepo,
		cardRepo:    cardRepo,
		memberRepo:  memberRepo,
		publisher:   publisher,
	}
}

func (uc *DeleteReleaseUseCase) Execute(ctx context.Context, releaseID, boardID, userID string) error {
	// 1. Проверка доступа (только owner)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return domain.ErrNotOwner
	}

	// 2. Получаем релиз (проверяем существование)
	_, err = uc.releaseRepo.GetByID(ctx, releaseID, boardID)
	if err != nil {
		return err
	}

	// 3. Перемещаем карточки в бэклог
	uc.cardRepo.MoveToBacklog(ctx, boardID, releaseID)

	// 4. Удаляем релиз
	if err := uc.releaseRepo.Delete(ctx, releaseID, boardID); err != nil {
		return err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishReleaseDeleted(ctx, ReleaseDeletedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ReleaseID:    releaseID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish ReleaseDeleted", "error", err, "release_id", releaseID, "board_id", boardID)
		}
	}()

	return nil
}
