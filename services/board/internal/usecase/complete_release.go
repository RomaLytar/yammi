package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CompleteReleaseUseCase struct {
	releaseRepo  ReleaseRepository
	cardRepo     CardRepository
	settingsRepo BoardSettingsRepository
	memberRepo   MembershipRepository
	publisher    EventPublisher
}

func NewCompleteReleaseUseCase(releaseRepo ReleaseRepository, cardRepo CardRepository, settingsRepo BoardSettingsRepository, memberRepo MembershipRepository, publisher EventPublisher) *CompleteReleaseUseCase {
	return &CompleteReleaseUseCase{
		releaseRepo:  releaseRepo,
		cardRepo:     cardRepo,
		settingsRepo: settingsRepo,
		memberRepo:   memberRepo,
		publisher:    publisher,
	}
}

func (uc *CompleteReleaseUseCase) Execute(ctx context.Context, releaseID, boardID, userID string) (*domain.Release, int, error) {
	// 1. Проверка доступа (только owner)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return nil, 0, domain.ErrNotOwner
	}

	// 2. Получаем релиз
	release, err := uc.releaseRepo.GetByID(ctx, releaseID, boardID)
	if err != nil {
		return nil, 0, err
	}

	// 3. Перемещаем незавершённые карточки в бэклог
	var movedToBacklog int
	settings, err := uc.settingsRepo.GetByBoardID(ctx, boardID)
	if err != nil {
		return nil, 0, err
	}

	if settings.DoneColumnID != nil && *settings.DoneColumnID != "" {
		// Карточки в done колонке остаются в релизе, остальные — в бэклог
		movedToBacklog, err = uc.cardRepo.MoveToBacklogExceptColumn(ctx, boardID, releaseID, *settings.DoneColumnID)
	} else {
		// Нет done колонки — все карточки в бэклог
		movedToBacklog, err = uc.cardRepo.MoveToBacklog(ctx, boardID, releaseID)
	}
	if err != nil {
		return nil, 0, err
	}

	// 4. Завершаем релиз (валидация внутри)
	if err := release.Complete(); err != nil {
		return nil, 0, err
	}

	// 5. Сохраняем
	if err := uc.releaseRepo.Update(ctx, release); err != nil {
		return nil, 0, err
	}

	// 6. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishReleaseCompleted(ctx, ReleaseCompletedEvent{
			EventID:             generateEventID(),
			EventVersion:        1,
			OccurredAt:          release.UpdatedAt,
			BoardID:             boardID,
			ReleaseID:           release.ID,
			Name:                release.Name,
			ActorID:             userID,
			CardsMovedToBacklog: movedToBacklog,
		}); err != nil {
			slog.Error("failed to publish ReleaseCompleted", "error", err, "release_id", release.ID, "board_id", boardID)
		}
	}()

	return release, movedToBacklog, nil
}
