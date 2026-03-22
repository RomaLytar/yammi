package domain

import (
	"encoding/json"
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

// BoardEvent — одна запись на событие доски (вместо N записей в notifications).
type BoardEvent struct {
	ID        string
	BoardID   string
	ActorID   string
	EventType NotificationType
	Title     string
	Message   string
	Metadata  map[string]string
	CreatedAt time.Time
}

func NewBoardEvent(boardID, actorID string, eventType NotificationType, title, message string, metadata map[string]string) *BoardEvent {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	if len(title) > 250 {
		title = title[:247] + "..."
	}
	return &BoardEvent{
		ID:        generateUUID(),
		BoardID:   boardID,
		ActorID:   actorID,
		EventType: eventType,
		Title:     title,
		Message:   message,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}
}

// MetadataJSON сериализует metadata в JSON строку.
func (e *BoardEvent) MetadataJSON() string {
	data, _ := json.Marshal(e.Metadata)
	return string(data)
}

// ParseMetadataJSON десериализует JSON строку в map.
func ParseMetadataJSON(raw string) map[string]string {
	m := make(map[string]string)
	_ = json.Unmarshal([]byte(raw), &m)
	return m
}
