package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type CreateCommentUseCase struct {
	commentRepo CommentRepository
	membership  MembershipChecker
	publisher   EventPublisher
}

func NewCreateCommentUseCase(commentRepo CommentRepository, membership MembershipChecker, publisher EventPublisher) *CreateCommentUseCase {
	return &CreateCommentUseCase{
		commentRepo: commentRepo,
		membership:  membership,
		publisher:   publisher,
	}
}

func (uc *CreateCommentUseCase) Execute(ctx context.Context, cardID, boardID, userID, content string, parentID *string) (*domain.Comment, error) {
	// 1. Проверка доступа
	isMember, err := uc.membership.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверяем что карточка существует и принадлежит указанной доске
	exists, err := uc.membership.CardExistsInBoard(ctx, cardID, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrEmptyCardID
	}

	// 3. Если parent_id задан — проверяем что родительский комментарий существует и является корневым
	if parentID != nil && *parentID != "" {
		parent, err := uc.commentRepo.GetByID(ctx, *parentID)
		if err != nil {
			if err == domain.ErrCommentNotFound {
				return nil, domain.ErrParentNotFound
			}
			return nil, err
		}

		// Проверяем что родитель принадлежит той же доске (cross-board IDOR защита)
		if parent.BoardID != boardID {
			return nil, domain.ErrParentNotFound
		}

		// Запрещаем ответы на ответы (глубина вложенности = 1)
		if parent.IsReply() {
			return nil, domain.ErrNestedReply
		}
	} else {
		parentID = nil
	}

	// 3. Создаем комментарий (валидация внутри)
	comment, err := domain.NewComment(cardID, boardID, userID, content, parentID)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// 5. Увеличиваем reply_count у родителя
	if parentID != nil {
		if err := uc.commentRepo.IncrementReplyCount(ctx, *parentID); err != nil {
			// Не фейлим операцию — комментарий уже создан
			_ = err
		}
	}

	// 6. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCommentCreated(ctx, CommentCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   comment.CreatedAt,
			CommentID:    comment.ID,
			CardID:       comment.CardID,
			BoardID:      comment.BoardID,
			AuthorID:     comment.AuthorID,
			ParentID:     comment.ParentID,
			Content:      comment.Content,
		}); err != nil {
			slog.Error("failed to publish CommentCreated", "error", err, "comment_id", comment.ID, "card_id", comment.CardID)
		}
	}()

	return comment, nil
}
