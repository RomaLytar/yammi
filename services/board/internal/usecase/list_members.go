package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListMembersUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
}

func NewListMembersUseCase(boardRepo BoardRepository, memberRepo MembershipRepository) *ListMembersUseCase {
	return &ListMembersUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
	}
}

func (uc *ListMembersUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.Member, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем участников
	return uc.memberRepo.ListMembers(ctx, boardID, 100, 0)
}
