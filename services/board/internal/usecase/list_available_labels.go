package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// AvailableLabelsResult — результат запроса доступных меток для доски
type AvailableLabelsResult struct {
	BoardLabels        []*domain.Label
	UserLabels         []*domain.UserLabel
	UseBoardLabelsOnly bool
}

type ListAvailableLabelsUseCase struct {
	settingsRepo  BoardSettingsRepository
	labelRepo     LabelRepository
	userLabelRepo UserLabelRepository
	boardRepo     BoardRepository
	memberRepo    MembershipRepository
}

func NewListAvailableLabelsUseCase(
	settingsRepo BoardSettingsRepository,
	labelRepo LabelRepository,
	userLabelRepo UserLabelRepository,
	boardRepo BoardRepository,
	memberRepo MembershipRepository,
) *ListAvailableLabelsUseCase {
	return &ListAvailableLabelsUseCase{
		settingsRepo:  settingsRepo,
		labelRepo:     labelRepo,
		userLabelRepo: userLabelRepo,
		boardRepo:     boardRepo,
		memberRepo:    memberRepo,
	}
}

func (uc *ListAvailableLabelsUseCase) Execute(ctx context.Context, boardID, userID string) (*AvailableLabelsResult, error) {
	// 1. Проверка доступа (member может видеть метки)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем настройки доски
	settings, err := uc.settingsRepo.GetByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Получаем метки доски
	boardLabels, err := uc.labelRepo.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	result := &AvailableLabelsResult{
		BoardLabels:        boardLabels,
		UseBoardLabelsOnly: settings.UseBoardLabelsOnly,
	}

	// 4. Если разрешены глобальные метки — получаем user_labels владельца доски
	if !settings.UseBoardLabelsOnly {
		board, err := uc.boardRepo.GetByID(ctx, boardID)
		if err != nil {
			return nil, err
		}

		userLabels, err := uc.userLabelRepo.ListByUserID(ctx, board.OwnerID)
		if err != nil {
			return nil, err
		}

		result.UserLabels = userLabels
	}

	return result, nil
}
