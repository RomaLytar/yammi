package domain

import (
	"time"

	"github.com/google/uuid"
)

// Comment — сущность комментария к карточке.
// Поддерживает ответы на комментарии (parent_id), глубина вложенности — 1 уровень.
type Comment struct {
	ID         string
	CardID     string
	BoardID    string
	AuthorID   string
	ParentID   *string // nil для корневых комментариев
	Content    string
	ReplyCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewComment создает новый комментарий с валидацией
func NewComment(cardID, boardID, authorID, content string, parentID *string) (*Comment, error) {
	if cardID == "" {
		return nil, ErrEmptyCardID
	}

	if boardID == "" {
		return nil, ErrEmptyBoardID
	}

	if authorID == "" {
		return nil, ErrEmptyAuthorID
	}

	if content == "" {
		return nil, ErrEmptyText
	}

	if len(content) > 10000 {
		return nil, ErrContentTooLong
	}

	now := time.Now()
	return &Comment{
		ID:        uuid.NewString(),
		CardID:    cardID,
		BoardID:   boardID,
		AuthorID:  authorID,
		ParentID:  parentID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update обновляет текст комментария
func (c *Comment) Update(content string) error {
	if content == "" {
		return ErrEmptyText
	}

	if len(content) > 10000 {
		return ErrContentTooLong
	}

	c.Content = content
	c.UpdatedAt = time.Now()

	return nil
}

// IsReply возвращает true, если комментарий является ответом
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}
