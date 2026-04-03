package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateReleaseUseCase struct {
	releaseRepo ReleaseRepository
	memberRepo  MembershipRepository
	publisher   EventPublisher
}

func NewCreateReleaseUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateReleaseUseCase {
	return &CreateReleaseUseCase{
		releaseRepo: releaseRepo,
		memberRepo:  memberRepo,
		publisher:   publisher,
	}
}

func (uc *CreateReleaseUseCase) Execute(ctx context.Context, boardID, userID, name, description string, startDate, endDate *time.Time) (*domain.Release, error) {
	// 1. Проверка доступа (member может создавать)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверка лимита релизов на доску
	count, err := uc.releaseRepo.CountByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if count >= domain.MaxReleasesPerBoard {
		return nil, domain.ErrMaxReleasesReached
	}

	// 3. Создаем релиз (валидация внутри)
	release, err := domain.NewRelease(boardID, name, description, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.releaseRepo.Create(ctx, release); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishReleaseCreated(ctx, ReleaseCreatedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   release.CreatedAt,
			BoardID:      boardID,
			ReleaseID:    release.ID,
			Name:         release.Name,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish ReleaseCreated", "error", err, "release_id", release.ID, "board_id", boardID)
		}
	}()

	return release, nil
}
