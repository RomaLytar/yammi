package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type UpdateCommentUseCase struct {
	commentRepo CommentRepository
	membership  MembershipChecker
	publisher   EventPublisher
}

func NewUpdateCommentUseCase(commentRepo CommentRepository, membership MembershipChecker, publisher EventPublisher) *UpdateCommentUseCase {
	return &UpdateCommentUseCase{
		commentRepo: commentRepo,
		membership:  membership,
		publisher:   publisher,
	}
}

func (uc *UpdateCommentUseCase) Execute(ctx context.Context, commentID, boardID, userID, content string) (*domain.Comment, error) {
	// 1. Проверка доступа
	isMember, err := uc.membership.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем комментарий
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// 3. Проверяем автора — редактировать может только автор
	if comment.AuthorID != userID {
		return nil, domain.ErrNotAuthor
	}

	// 4. Обновляем текст (валидация внутри)
	if err := comment.Update(content); err != nil {
		return nil, err
	}

	// 5. Сохраняем
	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	// 6. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCommentUpdated(ctx, CommentUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   comment.UpdatedAt,
			CommentID:    comment.ID,
			CardID:       comment.CardID,
			BoardID:      comment.BoardID,
			ActorID:      userID,
			Content:      comment.Content,
		}); err != nil {
			slog.Error("failed to publish CommentUpdated", "error", err, "comment_id", comment.ID, "card_id", comment.CardID)
		}
	}()

	return comment, nil
}
