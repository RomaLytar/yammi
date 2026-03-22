package domain

import (
	"time"

	"github.com/google/uuid"
)

// ActivityType — тип активности по карточке
type ActivityType string

const (
	ActivityCardCreated    ActivityType = "card_created"
	ActivityCardUpdated    ActivityType = "card_updated"
	ActivityCardMoved      ActivityType = "card_moved"
	ActivityCardAssigned   ActivityType = "card_assigned"
	ActivityCardUnassigned ActivityType = "card_unassigned"
	ActivityCardDeleted       ActivityType = "card_deleted"
	ActivityAttachmentAdded   ActivityType = "attachment_added"
	ActivityAttachmentDeleted ActivityType = "attachment_deleted"
)

// Activity — запись в журнале активности карточки
type Activity struct {
	ID          string
	CardID      string
	BoardID     string
	ActorID     string
	Type        ActivityType
	Description string            // Human-readable: "Карточка создана", "Перемещена из 'To Do' в 'In Progress'"
	Changes     map[string]string // e.g. {"old_column": "To Do", "new_column": "In Progress"}
	CreatedAt   time.Time
}

// NewActivity создает новую запись активности с валидацией
func NewActivity(cardID, boardID, actorID string, activityType ActivityType, description string, changes map[string]string) (*Activity, error) {
	if cardID == "" {
		return nil, ErrCardNotFound
	}

	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if actorID == "" {
		return nil, ErrEmptyActorID
	}

	if activityType == "" {
		return nil, ErrInvalidActivityType
	}

	if changes == nil {
		changes = map[string]string{}
	}

	return &Activity{
		ID:          uuid.NewString(),
		CardID:      cardID,
		BoardID:     boardID,
		ActorID:     actorID,
		Type:        activityType,
		Description: description,
		Changes:     changes,
		CreatedAt:   time.Now(),
	}, nil
}
