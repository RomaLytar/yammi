package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

// BoardRepository определяет интерфейс для работы с досками
type BoardRepository interface {
	// Create создает новую доску и сохраняет её в БД
	Create(ctx context.Context, board *domain.Board) error

	// GetByID возвращает доску по ID
	GetByID(ctx context.Context, boardID string) (*domain.Board, error)

	// ListByUserID возвращает список досок пользователя (с фильтрацией, поиском, сортировкой)
	ListByUserID(ctx context.Context, userID string, limit int, cursor string, ownerOnly bool, search string, sortBy string) ([]*domain.Board, string, error)

	// Update обновляет доску (с проверкой оптимистичной блокировки)
	Update(ctx context.Context, board *domain.Board) error

	// Delete удаляет доску по ID
	Delete(ctx context.Context, boardID string) error

	// BatchDelete удаляет несколько досок в одной транзакции
	BatchDelete(ctx context.Context, boardIDs []string) error

	// TouchUpdatedAt обновляет updated_at доски (при изменении карточек/колонок)
	TouchUpdatedAt(ctx context.Context, boardID string) error
}

// ColumnRepository определяет интерфейс для работы с колонками
type ColumnRepository interface {
	// Create создает новую колонку
	Create(ctx context.Context, column *domain.Column) error

	// GetByID возвращает колонку по ID
	GetByID(ctx context.Context, columnID string) (*domain.Column, error)

	// ListByBoardID возвращает все колонки доски в порядке сортировки
	ListByBoardID(ctx context.Context, boardID string) ([]*domain.Column, error)

	// Update обновляет колонку
	Update(ctx context.Context, column *domain.Column) error

	// Delete удаляет колонку по ID
	Delete(ctx context.Context, columnID string) error
}

// CardRepository определяет интерфейс для работы с карточками
type CardRepository interface {
	// Create создает новую карточку
	Create(ctx context.Context, card *domain.Card) error

	// GetByID возвращает карточку по ID (фильтруется по boardID для защиты от IDOR)
	GetByID(ctx context.Context, cardID, boardID string) (*domain.Card, error)

	// GetLastInColumn возвращает последнюю карточку в колонке (для генерации lexorank)
	GetLastInColumn(ctx context.Context, columnID string) (*domain.Card, error)

	// ListByColumnID возвращает все карточки колонки в порядке позиции
	ListByColumnID(ctx context.Context, columnID string) ([]*domain.Card, error)

	// Update обновляет карточку
	Update(ctx context.Context, card *domain.Card) error

	// Delete удаляет карточку по ID
	Delete(ctx context.Context, cardID string) error

	// BatchDelete удаляет несколько карточек по ID в рамках одной доски
	BatchDelete(ctx context.Context, boardID string, cardIDs []string) error
}

// MembershipRepository определяет интерфейс для работы с членством в досках
type MembershipRepository interface {
	// AddMember добавляет пользователя в доску с указанной ролью
	AddMember(ctx context.Context, boardID, userID string, role domain.Role) error

	// RemoveMember удаляет пользователя из доски
	RemoveMember(ctx context.Context, boardID, userID string) error

	// IsMember проверяет, является ли пользователь членом доски и возвращает его роль
	IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error)

	// ListMembers возвращает список членов доски с пагинацией
	ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error)
}
