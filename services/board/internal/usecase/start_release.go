package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type StartReleaseUseCase struct {
	releaseRepo  ReleaseRepository
	memberRepo   MembershipRepository
	settingsRepo BoardSettingsRepository
	publisher    EventPublisher
}

func NewStartReleaseUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository, settingsRepo BoardSettingsRepository, publisher EventPublisher) *StartReleaseUseCase {
	return &StartReleaseUseCase{
		releaseRepo:  releaseRepo,
		memberRepo:   memberRepo,
		settingsRepo: settingsRepo,
		publisher:    publisher,
	}
}

func (uc *StartReleaseUseCase) Execute(ctx context.Context, releaseID, boardID, userID string) (*domain.Release, error) {
	// 1. Проверка доступа (только owner)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return nil, domain.ErrNotOwner
	}

	// 2. Получаем релиз
	release, err := uc.releaseRepo.GetByID(ctx, releaseID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Проверяем нет ли другого активного релиза
	activeRelease, err := uc.releaseRepo.GetActiveByBoardID(ctx, boardID)
	if err != nil && err != domain.ErrReleaseNotFound {
		return nil, err
	}
	if activeRelease != nil && activeRelease.ID != release.ID {
		return nil, domain.ErrActiveReleaseExists
	}

	// 4. Получаем sprint_duration_days из настроек доски
	durationDays := 14 // default
	settings, err := uc.settingsRepo.GetByBoardID(ctx, boardID)
	if err == nil && settings.SprintDurationDays >= 7 {
		durationDays = settings.SprintDurationDays
	}

	// 5. Запускаем (валидация внутри)
	if err := release.Start(durationDays); err != nil {
		return nil, err
	}

	// 6. Сохраняем
	if err := uc.releaseRepo.Update(ctx, release); err != nil {
		return nil, err
	}

	// 7. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishReleaseStarted(ctx, ReleaseStartedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   release.UpdatedAt,
			BoardID:      boardID,
			ReleaseID:    release.ID,
			Name:         release.Name,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish ReleaseStarted", "error", err, "release_id", release.ID, "board_id", boardID)
		}
	}()

	return release, nil
}
