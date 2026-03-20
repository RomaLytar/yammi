package domain

import (
	"time"

	"github.com/google/uuid"
)

// Board — минимальный aggregate root (только метаданные, БЕЗ members/columns/cards).
// Members проверяются через MembershipRepository (query, не загрузка всех в память).
// Columns/Cards загружаются отдельными запросами (granular API).
type Board struct {
	ID          string
	Title       string
	Description string
	OwnerID     string // Создатель доски (всегда имеет роль RoleOwner)
	Version     int    // Optimistic locking (инкрементируется при каждом изменении)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewBoard создает новую доску с валидацией
func NewBoard(title, description, ownerID string) (*Board, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	if ownerID == "" {
		return nil, ErrEmptyOwnerID
	}

	now := time.Now()
	return &Board{
		ID:          uuid.NewString(),
		Title:       title,
		Description: description,
		OwnerID:     ownerID,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update обновляет метаданные доски
func (b *Board) Update(title, description string) error {
	if title == "" {
		return ErrEmptyTitle
	}

	b.Title = title
	b.Description = description
	b.UpdatedAt = time.Now()
	b.Version++ // Optimistic locking

	return nil
}

// IsOwner проверяет, является ли пользователь владельцем
func (b *Board) IsOwner(userID string) bool {
	return b.OwnerID == userID
}

// IncrementVersion увеличивает версию (для optimistic locking)
// Вызывается в usecase перед Save() при любом изменении доски
func (b *Board) IncrementVersion() {
	b.Version++
	b.UpdatedAt = time.Now()
}
