package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListCardActivityUseCase struct {
	activityRepo ActivityRepository
	memberRepo   MembershipRepository
}

func NewListCardActivityUseCase(activityRepo ActivityRepository, memberRepo MembershipRepository) *ListCardActivityUseCase {
	return &ListCardActivityUseCase{
		activityRepo: activityRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *ListCardActivityUseCase) Execute(ctx context.Context, cardID, boardID, userID string, limit int, cursor string) ([]*domain.Activity, string, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, "", err
	}
	if !isMember {
		return nil, "", domain.ErrAccessDenied
	}

	// 2. Загружаем активность
	return uc.activityRepo.ListByCardID(ctx, cardID, boardID, limit, cursor)
}
