package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetActiveReleaseUseCase struct {
	releaseRepo ReleaseRepository
	memberRepo  MembershipRepository
}

func NewGetActiveReleaseUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository) *GetActiveReleaseUseCase {
	return &GetActiveReleaseUseCase{
		releaseRepo: releaseRepo,
		memberRepo:  memberRepo,
	}
}

func (uc *GetActiveReleaseUseCase) Execute(ctx context.Context, boardID, userID string) (*domain.Release, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем активный релиз (если нет — nil, nil)
	release, err := uc.releaseRepo.GetActiveByBoardID(ctx, boardID)
	if err == domain.ErrReleaseNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return release, nil
}
