package nats

import (
	"context"

	"github.com/RomaLytar/yammi/services/comment/internal/usecase"
)

// EventPublisher реализует интерфейс usecase.EventPublisher
type EventPublisher struct {
	publisher *Publisher
}

func NewEventPublisher(publisher *Publisher) *EventPublisher {
	return &EventPublisher{publisher: publisher}
}

func (p *EventPublisher) PublishCommentCreated(ctx context.Context, event usecase.CommentCreated) error {
	return p.publisher.Publish(ctx, "comment.created", event)
}

func (p *EventPublisher) PublishCommentUpdated(ctx context.Context, event usecase.CommentUpdated) error {
	return p.publisher.Publish(ctx, "comment.updated", event)
}

func (p *EventPublisher) PublishCommentDeleted(ctx context.Context, event usecase.CommentDeleted) error {
	return p.publisher.Publish(ctx, "comment.deleted", event)
}
