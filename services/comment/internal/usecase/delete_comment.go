package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type DeleteCommentUseCase struct {
	commentRepo CommentRepository
	membership  MembershipChecker
	publisher   EventPublisher
}

func NewDeleteCommentUseCase(commentRepo CommentRepository, membership MembershipChecker, publisher EventPublisher) *DeleteCommentUseCase {
	return &DeleteCommentUseCase{
		commentRepo: commentRepo,
		membership:  membership,
		publisher:   publisher,
	}
}

func (uc *DeleteCommentUseCase) Execute(ctx context.Context, commentID, boardID, userID string) error {
	// 1. Проверка доступа
	isMember, err := uc.membership.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Получаем комментарий
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// 3. Проверяем права: автор ИЛИ владелец доски
	if comment.AuthorID != userID {
		isOwner, err := uc.membership.IsOwner(ctx, boardID, userID)
		if err != nil {
			return err
		}
		if !isOwner {
			return domain.ErrNotAuthor
		}
	}

	// 4. Уменьшаем reply_count у родителя (если это ответ)
	if comment.ParentID != nil {
		if err := uc.commentRepo.DecrementReplyCount(ctx, *comment.ParentID); err != nil {
			// Не фейлим операцию
			_ = err
		}
	}

	// 5. Удаляем (каскадно удалит ответы через FK ON DELETE CASCADE)
	if err := uc.commentRepo.Delete(ctx, commentID); err != nil {
		return err
	}

	// 6. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishCommentDeleted(context.Background(), CommentDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   comment.UpdatedAt,
			CommentID:    comment.ID,
			CardID:       comment.CardID,
			BoardID:      comment.BoardID,
			ActorID:      userID,
		})
	}()

	return nil
}
