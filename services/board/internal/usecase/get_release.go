package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetReleaseUseCase struct {
	releaseRepo ReleaseRepository
	memberRepo  MembershipRepository
}

func NewGetReleaseUseCase(releaseRepo ReleaseRepository, memberRepo MembershipRepository) *GetReleaseUseCase {
	return &GetReleaseUseCase{
		releaseRepo: releaseRepo,
		memberRepo:  memberRepo,
	}
}

func (uc *GetReleaseUseCase) Execute(ctx context.Context, releaseID, boardID, userID string) (*domain.Release, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем релиз
	return uc.releaseRepo.GetByID(ctx, releaseID, boardID)
}
