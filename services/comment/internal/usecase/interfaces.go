package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

// CommentRepository определяет интерфейс для работы с комментариями
type CommentRepository interface {
	// Create создает новый комментарий
	Create(ctx context.Context, comment *domain.Comment) error

	// GetByID возвращает комментарий по ID
	GetByID(ctx context.Context, commentID string) (*domain.Comment, error)

	// ListByCardID возвращает комментарии карточки с курсорной пагинацией.
	// boardID используется для фильтрации — предотвращает cross-board IDOR.
	ListByCardID(ctx context.Context, cardID, boardID string, limit int, cursor string) ([]*domain.Comment, string, error)

	// Update обновляет комментарий
	Update(ctx context.Context, comment *domain.Comment) error

	// Delete удаляет комментарий по ID
	Delete(ctx context.Context, commentID string) error

	// CountByCardID возвращает количество комментариев к карточке.
	// boardID используется для фильтрации — предотвращает cross-board IDOR.
	CountByCardID(ctx context.Context, cardID, boardID string) (int, error)

	// IncrementReplyCount увеличивает счётчик ответов у родительского комментария
	IncrementReplyCount(ctx context.Context, commentID string) error

	// DecrementReplyCount уменьшает счётчик ответов у родительского комментария
	DecrementReplyCount(ctx context.Context, commentID string) error
}

// MembershipChecker проверяет членство пользователя в доске через Board Service gRPC
type MembershipChecker interface {
	// IsMember проверяет, является ли пользователь членом доски
	IsMember(ctx context.Context, boardID, userID string) (bool, error)

	// IsOwner проверяет, является ли пользователь владельцем доски
	IsOwner(ctx context.Context, boardID, userID string) (bool, error)

	// CardExistsInBoard проверяет, что карточка существует и принадлежит указанной доске.
	// Предотвращает создание orphan-комментариев к несуществующим или чужим карточкам.
	CardExistsInBoard(ctx context.Context, cardID, boardID, userID string) (bool, error)
}
