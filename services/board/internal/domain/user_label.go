package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// UserLabel — пользовательская метка (user-scoped, глобальная).
// Не привязана к доске, принадлежит пользователю.
type UserLabel struct {
	ID        string
	UserID    string
	Name      string
	Color     string // hex color like "#ef4444"
	CreatedAt time.Time
}

// NewUserLabel создает новую пользовательскую метку с валидацией
func NewUserLabel(id, userID, name, color string) (*UserLabel, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	if name == "" {
		return nil, ErrEmptyLabelName
	}

	if !ValidateColor(color) {
		return nil, ErrInvalidColor
	}

	if id == "" {
		id = uuid.NewString()
	}

	return &UserLabel{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}, nil
}

// Update обновляет метаданные пользовательской метки
func (l *UserLabel) Update(name, color string) error {
	if name == "" {
		return ErrEmptyLabelName
	}

	if !ValidateColor(color) {
		return ErrInvalidColor
	}

	l.Name = name
	l.Color = color

	return nil
}
