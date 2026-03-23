package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Label — метка доски (board-scoped colored tag).
// Может быть назначена на карточки (many-to-many через card_labels).
type Label struct {
	ID        string
	BoardID   string
	Name      string
	Color     string // hex color like "#ef4444"
	CreatedAt time.Time
}

// colorRegexp проверяет формат #rrggbb
var colorRegexp = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// ValidateColor проверяет формат цвета (#rrggbb)
func ValidateColor(color string) bool {
	return colorRegexp.MatchString(color)
}

// NewLabel создает новую метку с валидацией
func NewLabel(id, boardID, name, color string) (*Label, error) {
	if boardID == "" {
		return nil, ErrBoardNotFound
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

	return &Label{
		ID:        id,
		BoardID:   boardID,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}, nil
}

// Update обновляет метаданные метки
func (l *Label) Update(name, color string) error {
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
