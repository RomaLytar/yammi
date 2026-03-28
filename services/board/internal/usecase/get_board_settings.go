package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetBoardSettingsUseCase struct {
	settingsRepo BoardSettingsRepository
	memberRepo   MembershipRepository
}

func NewGetBoardSettingsUseCase(settingsRepo BoardSettingsRepository, memberRepo MembershipRepository) *GetBoardSettingsUseCase {
	return &GetBoardSettingsUseCase{
		settingsRepo: settingsRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *GetBoardSettingsUseCase) Execute(ctx context.Context, boardID, userID string) (*domain.BoardSettings, error) {
	// 1. Проверка доступа (member может смотреть настройки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем настройки (дефолтные, если нет записи)
	return uc.settingsRepo.GetByBoardID(ctx, boardID)
}
