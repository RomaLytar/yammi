package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type ListCommentsUseCase struct {
	commentRepo CommentRepository
	membership  MembershipChecker
}

func NewListCommentsUseCase(commentRepo CommentRepository, membership MembershipChecker) *ListCommentsUseCase {
	return &ListCommentsUseCase{
		commentRepo: commentRepo,
		membership:  membership,
	}
}

func (uc *ListCommentsUseCase) Execute(ctx context.Context, cardID, boardID, userID string, limit int, cursor string) ([]*domain.Comment, string, error) {
	// 1. Проверка доступа
	isMember, err := uc.membership.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, "", err
	}
	if !isMember {
		return nil, "", domain.ErrAccessDenied
	}

	// 2. Валидация лимита
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	// 3. Запрашиваем комментарии (boardID фильтрация предотвращает cross-board IDOR)
	comments, nextCursor, err := uc.commentRepo.ListByCardID(ctx, cardID, boardID, limit, cursor)
	if err != nil {
		return nil, "", err
	}

	return comments, nextCursor, nil
}
