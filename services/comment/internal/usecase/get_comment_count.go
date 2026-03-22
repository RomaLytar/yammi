package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type GetCommentCountUseCase struct {
	commentRepo CommentRepository
	membership  MembershipChecker
}

func NewGetCommentCountUseCase(commentRepo CommentRepository, membership MembershipChecker) *GetCommentCountUseCase {
	return &GetCommentCountUseCase{
		commentRepo: commentRepo,
		membership:  membership,
	}
}

func (uc *GetCommentCountUseCase) Execute(ctx context.Context, cardID, boardID, userID string) (int, error) {
	// 1. Проверка доступа
	isMember, err := uc.membership.IsMember(ctx, boardID, userID)
	if err != nil {
		return 0, err
	}
	if !isMember {
		return 0, domain.ErrAccessDenied
	}

	// 2. Считаем комментарии
	count, err := uc.commentRepo.CountByCardID(ctx, cardID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
