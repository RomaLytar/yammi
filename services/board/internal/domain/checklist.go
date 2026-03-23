package domain

import (
	"time"

	"github.com/google/uuid"
)

// Checklist — чеклист карточки (card-scoped).
// Каждая карточка может иметь несколько чеклистов.
type Checklist struct {
	ID        string
	CardID    string
	BoardID   string
	Title     string
	Position  int
	Items     []ChecklistItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ChecklistItem — элемент чеклиста.
type ChecklistItem struct {
	ID          string
	ChecklistID string
	BoardID     string
	Title       string
	IsChecked   bool
	Position    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewChecklist создает новый чеклист с валидацией
func NewChecklist(id, cardID, boardID, title string, position int) (*Checklist, error) {
	if cardID == "" {
		return nil, ErrCardNotFound
	}

	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if title == "" {
		return nil, ErrEmptyChecklistTitle
	}

	if id == "" {
		id = uuid.NewString()
	}

	now := time.Now()
	return &Checklist{
		ID:        id,
		CardID:    cardID,
		BoardID:   boardID,
		Title:     title,
		Position:  position,
		Items:     nil,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update обновляет заголовок чеклиста
func (c *Checklist) Update(title string) error {
	if title == "" {
		return ErrEmptyChecklistTitle
	}

	c.Title = title
	c.UpdatedAt = time.Now()

	return nil
}

// Progress возвращает процент выполнения (0-100)
func (c *Checklist) Progress() int {
	if len(c.Items) == 0 {
		return 0
	}

	checked := 0
	for _, item := range c.Items {
		if item.IsChecked {
			checked++
		}
	}

	return checked * 100 / len(c.Items)
}

// NewChecklistItem создает новый элемент чеклиста с валидацией
func NewChecklistItem(id, checklistID, boardID, title string, position int) (*ChecklistItem, error) {
	if checklistID == "" {
		return nil, ErrChecklistNotFound
	}

	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if title == "" {
		return nil, ErrEmptyItemTitle
	}

	if id == "" {
		id = uuid.NewString()
	}

	now := time.Now()
	return &ChecklistItem{
		ID:          id,
		ChecklistID: checklistID,
		BoardID:     boardID,
		Title:       title,
		IsChecked:   false,
		Position:    position,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update обновляет заголовок элемента чеклиста
func (i *ChecklistItem) Update(title string) error {
	if title == "" {
		return ErrEmptyItemTitle
	}

	i.Title = title
	i.UpdatedAt = time.Now()

	return nil
}

// Toggle переключает состояние is_checked
func (i *ChecklistItem) Toggle() {
	i.IsChecked = !i.IsChecked
	i.UpdatedAt = time.Now()
}
