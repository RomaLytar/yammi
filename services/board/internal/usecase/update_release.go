package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateReleaseUseCase struct {
	releaseRepo ReleaseRepository
	memberRepo  MembershipRepository
	publisher   EventPublisher
}

func NewUpdateReleaseUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateReleaseUseCase {
	return &UpdateReleaseUseCase{
		releaseRepo: releaseRepo,
		memberRepo:  memberRepo,
		publisher:   publisher,
	}
}

func (uc *UpdateReleaseUseCase) Execute(ctx context.Context, releaseID, boardID, userID, name, description string, startDate, endDate *time.Time) (*domain.Release, error) {
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

	// 3. Обновляем (валидация внутри)
	if err := release.Update(name, description, startDate, endDate); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.releaseRepo.Update(ctx, release); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishReleaseUpdated(ctx, ReleaseUpdatedEvent{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   release.UpdatedAt,
			BoardID:      boardID,
			ReleaseID:    release.ID,
			Name:         release.Name,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish ReleaseUpdated", "error", err, "release_id", release.ID, "board_id", boardID)
		}
	}()

	return release, nil
}
