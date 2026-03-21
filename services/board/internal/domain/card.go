package domain

import (
	"time"

	"github.com/google/uuid"
)

// Card — aggregate root для карточки.
// Отделен от Board/Column для производительности и granular cache.
// Position использует lexorank (string) вместо INT для efficient reordering.
type Card struct {
	ID          string
	ColumnID    string
	Title       string
	Description string
	Position    string  // lexorank (a, am, b, c, ...) — НЕ INT!
	AssigneeID  *string // опциональный исполнитель
	CreatorID   string  // кто создал карточку
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewCard создает новую карточку с валидацией
func NewCard(columnID, title, description, position string, assigneeID *string, creatorID string) (*Card, error) {
	if columnID == "" {
		return nil, ErrColumnNotFound
	}

	if title == "" {
		return nil, ErrEmptyCardTitle
	}

	if err := ValidateLexorank(position); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Card{
		ID:          uuid.NewString(),
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Position:    position,
		AssigneeID:  assigneeID,
		CreatorID:   creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update обновляет метаданные карточки
func (c *Card) Update(title, description string, assigneeID *string) error {
	if title == "" {
		return ErrEmptyCardTitle
	}

	c.Title = title
	c.Description = description
	c.AssigneeID = assigneeID
	c.UpdatedAt = time.Now()

	return nil
}

// Move перемещает карточку в другую колонку с новой позицией
func (c *Card) Move(targetColumnID, newPosition string) error {
	if targetColumnID == "" {
		return ErrColumnNotFound
	}

	if err := ValidateLexorank(newPosition); err != nil {
		return err
	}

	c.ColumnID = targetColumnID
	c.Position = newPosition
	c.UpdatedAt = time.Now()

	return nil
}

// Reorder изменяет позицию карточки в текущей колонке
func (c *Card) Reorder(newPosition string) error {
	if err := ValidateLexorank(newPosition); err != nil {
		return err
	}

	c.Position = newPosition
	c.UpdatedAt = time.Now()

	return nil
}
