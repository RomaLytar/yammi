package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AssignCardToReleaseUseCase struct {
	releaseRepo ReleaseRepository
	cardRepo    CardRepository
	memberRepo  MembershipRepository
	publisher   EventPublisher
}

func NewAssignCardToReleaseUseCase(releaseRepo ReleaseRepository, cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *AssignCardToReleaseUseCase {
	return &AssignCardToReleaseUseCase{
		releaseRepo: releaseRepo,
		cardRepo:    cardRepo,
		memberRepo:  memberRepo,
		publisher:   publisher,
	}
}

func (uc *AssignCardToReleaseUseCase) Execute(ctx context.Context, cardID, releaseID, boardID, userID string) error {
	// 1. Проверка доступа (member может назначать)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Получаем релиз (проверяем, что не завершён)
	release, err := uc.releaseRepo.GetByID(ctx, releaseID, boardID)
	if err != nil {
		return err
	}
	if release.IsCompleted() {
		return domain.ErrReleaseCompleted
	}

	// 3. Получаем карточку (проверяем принадлежность к доске)
	_, err = uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return err
	}

	// 4. Устанавливаем release_id
	if err := uc.cardRepo.SetReleaseID(ctx, cardID, boardID, &releaseID); err != nil {
		return err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardReleaseAssigned(ctx, CardReleaseAssignedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			CardID:       cardID,
			ReleaseID:    releaseID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardReleaseAssigned", "error", err, "card_id", cardID, "release_id", releaseID, "board_id", boardID)
		}
	}()

	return nil
}
