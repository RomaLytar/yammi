package domain

import (
	"time"

	"github.com/google/uuid"
)

// Column — aggregate root для колонки доски.
// Отделен от Board для производительности (не загружаем все колонки при GetBoard).
type Column struct {
	ID        string
	BoardID   string
	Title     string
	Position  int       // INT позиция (колонок обычно <= 20, reorder редкий)
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewColumn создает новую колонку с валидацией
func NewColumn(boardID, title string, position int) (*Column, error) {
	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if title == "" {
		return nil, ErrEmptyColumnTitle
	}

	if position < 0 {
		return nil, ErrInvalidPosition
	}

	now := time.Now()
	return &Column{
		ID:        uuid.NewString(),
		BoardID:   boardID,
		Title:     title,
		Position:  position,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update обновляет заголовок колонки
func (c *Column) Update(title string) error {
	if title == "" {
		return ErrEmptyColumnTitle
	}

	c.Title = title
	c.UpdatedAt = time.Now()
	return nil
}

// UpdatePosition обновляет позицию колонки (для reorder)
func (c *Column) UpdatePosition(position int) error {
	if position < 0 {
		return ErrInvalidPosition
	}

	c.Position = position
	return nil
}
