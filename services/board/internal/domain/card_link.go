package domain

import (
	"time"

	"github.com/google/uuid"
)

// CardLinkType представляет тип связи между карточками
type CardLinkType string

const (
	LinkTypeSubtask CardLinkType = "subtask"
)

// IsValid проверяет валидность типа связи
func (lt CardLinkType) IsValid() bool {
	return lt == LinkTypeSubtask
}

// String реализует fmt.Stringer
func (lt CardLinkType) String() string {
	return string(lt)
}

// CardLink — связь parent->child между карточками (подзадача).
// Обе карточки могут быть на разных досках.
type CardLink struct {
	ID        string
	ParentID  string       // ID карточки-родителя
	ChildID   string       // ID карточки-потомка (подзадача)
	BoardID   string       // board_id родительской карточки (для партиционирования)
	LinkType  CardLinkType
	CreatedAt time.Time
}

// NewCardLink создает новую связь между карточками с валидацией
func NewCardLink(id, parentID, childID, boardID string, linkType CardLinkType) (*CardLink, error) {
	if parentID == "" {
		return nil, ErrCardNotFound
	}

	if childID == "" {
		return nil, ErrCardNotFound
	}

	if parentID == childID {
		return nil, ErrSelfLink
	}

	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if !linkType.IsValid() {
		return nil, ErrInvalidLinkType
	}

	if id == "" {
		id = uuid.NewString()
	}

	return &CardLink{
		ID:        id,
		ParentID:  parentID,
		ChildID:   childID,
		BoardID:   boardID,
		LinkType:  linkType,
		CreatedAt: time.Now(),
	}, nil
}
