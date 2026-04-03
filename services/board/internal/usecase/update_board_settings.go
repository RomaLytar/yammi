package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateBoardSettingsUseCase struct {
	settingsRepo BoardSettingsRepository
	memberRepo   MembershipRepository
}

func NewUpdateBoardSettingsUseCase(settingsRepo BoardSettingsRepository, memberRepo MembershipRepository) *UpdateBoardSettingsUseCase {
	return &UpdateBoardSettingsUseCase{
		settingsRepo: settingsRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *UpdateBoardSettingsUseCase) Execute(ctx context.Context, boardID, userID string, useBoardLabelsOnly bool, doneColumnID *string, sprintDurationDays int, releasesEnabled bool) (*domain.BoardSettings, error) {
	// 1. Проверка доступа (только owner может менять настройки)
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

	// 2. Валидация sprint_duration_days (минимум 7)
	if sprintDurationDays < 7 {
		return nil, domain.ErrInvalidSprintDuration
	}

	// 3. Создаем/обновляем настройки
	settings := domain.NewBoardSettings(boardID)
	settings.Update(useBoardLabelsOnly, doneColumnID, sprintDurationDays, releasesEnabled)

	// 4. Upsert (INSERT ON CONFLICT DO UPDATE)
	if err := uc.settingsRepo.Upsert(ctx, settings); err != nil {
		return nil, err
	}

	return settings, nil
}
