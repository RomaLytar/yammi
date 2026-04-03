package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListReleasesUseCase struct {
	releaseRepo ReleaseRepository
	memberRepo  MembershipRepository
}

func NewListReleasesUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository) *ListReleasesUseCase {
	return &ListReleasesUseCase{
		releaseRepo: releaseRepo,
		memberRepo:  memberRepo,
	}
}

func (uc *ListReleasesUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.Release, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Список релизов
	return uc.releaseRepo.ListByBoardID(ctx, boardID)
}
