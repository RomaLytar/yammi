package usecase

import (
	"context"
	"sync"

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
	// 1. Проверка доступа (member может видеть метки) — Redis cache, ~1ms
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Параллельно загружаем настройки, метки доски и данные доски (для ownerID)
	var (
		settings    *domain.BoardSettings
		boardLabels []*domain.Label
		board       *domain.Board
		settingsErr error
		labelsErr   error
		boardErr    error
	)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		settings, settingsErr = uc.settingsRepo.GetByBoardID(ctx, boardID)
	}()

	go func() {
		defer wg.Done()
		boardLabels, labelsErr = uc.labelRepo.ListByBoardID(ctx, boardID)
	}()

	go func() {
		defer wg.Done()
		board, boardErr = uc.boardRepo.GetByID(ctx, boardID)
	}()

	wg.Wait()

	if settingsErr != nil {
		return nil, settingsErr
	}
	if labelsErr != nil {
		return nil, labelsErr
	}
	if boardErr != nil {
		return nil, boardErr
	}

	result := &AvailableLabelsResult{
		BoardLabels:        boardLabels,
		UseBoardLabelsOnly: settings.UseBoardLabelsOnly,
	}

	// 3. Если разрешены глобальные метки — получаем user_labels владельца доски
	if !settings.UseBoardLabelsOnly {
		userLabels, err := uc.userLabelRepo.ListByUserID(ctx, board.OwnerID)
		if err != nil {
			return nil, err
		}
		result.UserLabels = userLabels
	}

	return result, nil
}
