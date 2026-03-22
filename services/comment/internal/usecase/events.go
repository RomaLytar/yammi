package usecase

import (
	"time"

	"github.com/google/uuid"
)

// CommentCreated событие создания комментария
type CommentCreated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CommentID    string    `json:"comment_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	AuthorID     string    `json:"author_id"`
	ParentID     *string   `json:"parent_id,omitempty"`
	Content      string    `json:"content"`
}

// CommentUpdated событие обновления комментария
type CommentUpdated struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CommentID    string    `json:"comment_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
	Content      string    `json:"content"`
}

// CommentDeleted событие удаления комментария
type CommentDeleted struct {
	EventID      string    `json:"event_id"`
	EventVersion int       `json:"event_version"`
	OccurredAt   time.Time `json:"occurred_at"`
	CommentID    string    `json:"comment_id"`
	CardID       string    `json:"card_id"`
	BoardID      string    `json:"board_id"`
	ActorID      string    `json:"actor_id"`
}

func generateEventID() string {
	return uuid.New().String()
}
