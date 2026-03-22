package domain

import (
	"time"
)

type NotificationType string

const (
	TypeWelcome       NotificationType = "welcome"
	TypeBoardCreated  NotificationType = "board_created"
	TypeBoardUpdated  NotificationType = "board_updated"
	TypeBoardDeleted  NotificationType = "board_deleted"
	TypeColumnCreated NotificationType = "column_created"
	TypeColumnUpdated NotificationType = "column_updated"
	TypeColumnDeleted NotificationType = "column_deleted"
	TypeCardCreated   NotificationType = "card_created"
	TypeCardUpdated   NotificationType = "card_updated"
	TypeCardMoved     NotificationType = "card_moved"
	TypeCardDeleted   NotificationType = "card_deleted"
	TypeMemberAdded   NotificationType = "member_added"
	TypeMemberRemoved NotificationType = "member_removed"
)

type Notification struct {
	ID        string
	UserID    string
	Type      NotificationType
	Title     string
	Message   string
	Metadata  map[string]string
	IsRead    bool
	CreatedAt time.Time
}

// NewNotification создает новое уведомление с валидацией полей.
func NewNotification(userID string, ntype NotificationType, title, message string, metadata map[string]string) (*Notification, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if ntype == "" {
		return nil, ErrEmptyType
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}

	if len(title) > 250 {
		title = title[:247] + "..."
	}

	return &Notification{
		ID:        generateUUID(),
		UserID:    userID,
		Type:      ntype,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		IsRead:    false,
		CreatedAt: time.Now(),
	}, nil
}
