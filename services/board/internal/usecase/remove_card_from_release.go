package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type RemoveCardFromReleaseUseCase struct {
	releaseRepo ReleaseRepository
	cardRepo    CardRepository
	memberRepo  MembershipRepository
	publisher   EventPublisher
}

func NewRemoveCardFromReleaseUseCase(releaseRepo ReleaseRepository, cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *RemoveCardFromReleaseUseCase {
	return &RemoveCardFromReleaseUseCase{
		releaseRepo: releaseRepo,
		cardRepo:    cardRepo,
		memberRepo:  memberRepo,
		publisher:   publisher,
	}
}

func (uc *RemoveCardFromReleaseUseCase) Execute(ctx context.Context, cardID, boardID, userID string) error {
	// 1. Проверка доступа (member может снимать)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Получаем карточку (проверяем, что есть release)
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return err
	}
	if card.ReleaseID == nil || *card.ReleaseID == "" {
		return domain.ErrReleaseNotFound
	}

	// 3. Получаем релиз (проверяем, что не завершён)
	release, err := uc.releaseRepo.GetByID(ctx, *card.ReleaseID, boardID)
	if err != nil {
		return err
	}
	if release.IsCompleted() {
		return domain.ErrReleaseCompleted
	}

	releaseID := *card.ReleaseID

	// 4. Снимаем release_id
	if err := uc.cardRepo.SetReleaseID(ctx, cardID, boardID, nil); err != nil {
		return err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardReleaseRemoved(ctx, CardReleaseRemovedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			CardID:       cardID,
			ReleaseID:    releaseID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardReleaseRemoved", "error", err, "card_id", cardID, "release_id", releaseID, "board_id", boardID)
		}
	}()

	return nil
}
